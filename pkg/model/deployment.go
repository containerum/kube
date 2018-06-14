package model

import (
	"fmt"

	"strconv"

	"path"
	"strings"

	"time"

	"github.com/blang/semver"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/pkg/errors"
	api_apps "k8s.io/api/apps/v1"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

const (
	deploymentKind       = "Deployment"
	deploymentAPIVersion = "apps/v1"

	minDeployCPU      = 10   //m
	minDeployMemory   = 10   //Mi
	maxDeployCPU      = 3000 //m
	maxDeployMemory   = 8000 //Mi
	maxDeployReplicas = 15

	volumePostfix = "-volume"
	cmPostfix     = "-cm"
)

type DeploymentKubeAPI kube_types.Deployment

// ParseKubeDeploymentList parses kubernetes v1.DeploymentList to more convenient []Deployment struct
func ParseKubeDeploymentList(deploys interface{}, parseforuser bool) (*kube_types.DeploymentsList, error) {
	deployList := deploys.(*api_apps.DeploymentList)
	if deployList == nil {
		return nil, ErrUnableConvertDeploymentList
	}

	deployments := make([]kube_types.Deployment, 0)
	for _, deployment := range deployList.Items {
		deployment, err := ParseKubeDeployment(&deployment, parseforuser)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, *deployment)
	}
	return &kube_types.DeploymentsList{deployments}, nil
}

// ParseKubeDeployment parses kubernetes v1.Deployment to more convenient Deployment struct
func ParseKubeDeployment(deployment interface{}, parseforuser bool) (*kube_types.Deployment, error) {
	deploy := deployment.(*api_apps.Deployment)
	if deploy == nil {
		return nil, ErrUnableConvertDeployment
	}

	replicas := 0
	if r := deploy.Spec.Replicas; r != nil {
		replicas = int(*r)
	}
	containers, totalcpu, totalmem := getContainers(deploy.Spec.Template.Spec.Containers, getVolumeMode(deploy.Spec.Template.Spec.Volumes), getVolumeStorageName(deploy.Spec.Template.Spec.Volumes), replicas)

	version, _ := semver.ParseTolerant(deploy.GetObjectMeta().GetLabels()["version"])

	newDeploy := kube_types.Deployment{
		Name:     deploy.GetName(),
		Replicas: replicas,
		Status: &kube_types.DeploymentStatus{
			Replicas:            int(deploy.Status.Replicas),
			ReadyReplicas:       int(deploy.Status.ReadyReplicas),
			AvailableReplicas:   int(deploy.Status.AvailableReplicas),
			UpdatedReplicas:     int(deploy.Status.UpdatedReplicas),
			UnavailableReplicas: int(deploy.Status.UnavailableReplicas),
		},
		CreatedAt:   deploy.ObjectMeta.CreationTimestamp.UTC().Format(time.RFC3339),
		SolutionID:  deploy.GetObjectMeta().GetLabels()[solutionLabel],
		Containers:  containers,
		TotalCPU:    uint(totalcpu.ScaledValue(api_resource.Milli)),
		TotalMemory: uint(totalmem.Value() / 1024 / 1024),
		Owner:       deploy.GetObjectMeta().GetLabels()[ownerLabel],
		Version:     version,
		Active:      true,
	}

	if parseforuser {
		newDeploy.Mask()
	}

	return &newDeploy, nil
}

func getVolumeMode(volumes []api_core.Volume) map[string]int32 {
	volumemap := make(map[string]int32, 0)
	for _, v := range volumes {
		if v.ConfigMap != nil {
			volumemap[v.Name] = *v.ConfigMap.DefaultMode
		}
	}
	return volumemap
}

func getVolumeStorageName(volumes []api_core.Volume) map[string]string {
	volumemap := make(map[string]string, 0)
	for _, v := range volumes {
		if v.PersistentVolumeClaim != nil {
			volumemap[v.Name] = v.PersistentVolumeClaim.ClaimName
		}
	}
	return volumemap
}

//ToKube creates kubernetes v1.Deployment from Deployment struct and namespace labels
func (deploy *DeploymentKubeAPI) ToKube(nsName string, labels map[string]string) (*api_apps.Deployment, []error) {
	err := deploy.Validate()
	if err != nil {
		return nil, err
	}

	repl := int32(deploy.Replicas)
	containers, err := makeContainers(deploy.Containers)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{errors.New("invalid namespace labels")}
	}

	labels[appLabel] = deploy.Name

	if deploy.SolutionID != "" {
		labels[solutionLabel] = deploy.SolutionID
	}

	if deploy.Version.String() != "" {
		labels["version"] = deploy.Version.String()
	}

	volumes, verr := makeTemplateVolumes(deploy.Containers)
	if verr != nil {
		return nil, []error{verr}
	}

	newDeploy := api_apps.Deployment{
		TypeMeta: api_meta.TypeMeta{
			Kind:       deploymentKind,
			APIVersion: deploymentAPIVersion,
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      deploy.Name,
			Namespace: nsName,
		},
		Spec: api_apps.DeploymentSpec{
			Selector: &api_meta.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &repl,
			Strategy: api_apps.DeploymentStrategy{
				Type: api_apps.RecreateDeploymentStrategyType,
			},
			Template: api_core.PodTemplateSpec{
				Spec: api_core.PodSpec{
					Containers: containers,
					NodeSelector: map[string]string{
						"role": "slave",
					},
					Volumes: volumes,
				},
				ObjectMeta: api_meta.ObjectMeta{
					Labels: labels,
				},
			},
		},
	}

	return &newDeploy, nil
}

func makeContainers(containers []kube_types.Container) ([]api_core.Container, []error) {
	var containersAfter []api_core.Container

	for _, c := range containers {
		errs := validateContainer(c, c.Limits.CPU, c.Limits.Memory)
		if errs != nil {
			return nil, errs
		}

		container := api_core.Container{
			Name:    c.Name,
			Image:   c.Image,
			Command: makeContainerCommands(c.Commands),
		}

		if c.VolumeMounts != nil || c.ConfigMaps != nil {
			container.VolumeMounts = makeContainerVolumes(c.VolumeMounts, c.ConfigMaps)
		}

		if c.Env != nil {
			container.Env = makeContainerEnv(c.Env)
		}

		if c.Ports != nil {
			container.Ports = makeContainerPorts(c.Ports)
		}

		rq := makeContainerResourceQuota(c.Limits.CPU, c.Limits.Memory)

		container.Resources = *rq

		containersAfter = append(containersAfter, container)
	}
	return containersAfter, nil
}

func makeContainerVolumes(volumes []kube_types.ContainerVolume, configMaps []kube_types.ContainerVolume) []api_core.VolumeMount {
	volumeMounts := make([]api_core.VolumeMount, 0)
	for _, v := range volumes {
		var subpath string
		if v.SubPath != nil {
			subpath = *v.SubPath
		}
		if v.PersistentVolumeClaimName != nil {
			volumeMounts = append(volumeMounts, api_core.VolumeMount{Name: *v.PersistentVolumeClaimName + volumePostfix, MountPath: v.MountPath, SubPath: subpath})
		}
	}
	for _, v := range configMaps {
		var subpath string
		if v.SubPath != nil {
			subpath = *v.SubPath
		}
		volumeMounts = append(volumeMounts, api_core.VolumeMount{Name: v.Name + cmPostfix, MountPath: v.MountPath, SubPath: subpath})
	}

	return volumeMounts
}

func makeContainerEnv(env []kube_types.Env) []api_core.EnvVar {
	envvar := make([]api_core.EnvVar, 0)
	if env != nil {
		for _, v := range env {
			envvar = append(envvar, api_core.EnvVar{Name: v.Name, Value: v.Value})
		}
	}
	return envvar
}

func makeContainerPorts(ports []kube_types.ContainerPort) []api_core.ContainerPort {
	contports := make([]api_core.ContainerPort, 0)
	if ports != nil {
		for _, v := range ports {
			contports = append(contports, api_core.ContainerPort{ContainerPort: int32(v.Port), Protocol: api_core.Protocol(v.Protocol), Name: v.Name})
		}
	}
	return contports
}

func makeContainerCommands(commands []string) []string {
	contcommands := make([]string, 0)
	if commands != nil {
		contcommands = commands
	}
	return contcommands
}

func makeContainerResourceQuota(cpu, memory uint) *api_core.ResourceRequirements {
	limits := make(map[api_core.ResourceName]api_resource.Quantity)
	requests := make(map[api_core.ResourceName]api_resource.Quantity)

	lcpu := api_resource.NewScaledQuantity(int64(cpu), api_resource.Milli)
	lmem := api_resource.NewQuantity(int64(memory)*1024*1024, api_resource.BinarySI)
	rcpu := api_resource.NewScaledQuantity(int64(cpu/2), api_resource.Milli)
	rmem := api_resource.NewQuantity(int64(memory/2)*1024*1024, api_resource.BinarySI)

	limits["cpu"] = *lcpu
	limits["memory"] = *lmem
	requests["cpu"] = *rcpu
	requests["memory"] = *rmem

	return &api_core.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func makeTemplateVolumes(containers []kube_types.Container) ([]api_core.Volume, error) {
	templateVolumes := make([]api_core.Volume, 0)
	existingVolume := make(map[string]bool, 0)
	existingMountPath := make(map[string]bool, 0)

	for _, c := range containers {
		for _, v := range c.VolumeMounts {
			newVolume := api_core.Volume{
				Name: *v.PersistentVolumeClaimName + volumePostfix,
				VolumeSource: api_core.VolumeSource{
					PersistentVolumeClaim: &api_core.PersistentVolumeClaimVolumeSource{
						ClaimName: *v.PersistentVolumeClaimName,
					},
				},
			}
			if !existingMountPath[v.MountPath] {
				existingMountPath[v.MountPath] = true
			} else {
				return nil, fmt.Errorf(duplicateMountPath, v.MountPath)
			}

			if !existingVolume[newVolume.Name] {
				templateVolumes = append(templateVolumes, newVolume)
				existingVolume[newVolume.Name] = true
			} else {
				continue
			}
		}

		for _, v := range c.ConfigMaps {
			defMode := int32(0644)
			if v.Mode != nil {
				if mode, err := strconv.ParseInt(*v.Mode, 8, 32); err == nil {
					defMode = int32(mode)
				}
			}

			newVolume := api_core.Volume{
				Name: v.Name + cmPostfix,
				VolumeSource: api_core.VolumeSource{
					ConfigMap: &api_core.ConfigMapVolumeSource{
						LocalObjectReference: api_core.LocalObjectReference{
							Name: v.Name,
						},
						DefaultMode: &defMode,
					},
				},
			}
			if !existingMountPath[v.MountPath] {
				existingMountPath[v.MountPath] = true
			} else {
				return nil, fmt.Errorf(duplicateMountPath, v.MountPath)
			}

			if !existingVolume[newVolume.Name] {
				templateVolumes = append(templateVolumes, newVolume)
				existingVolume[newVolume.Name] = true
			} else {
				continue
			}
		}
	}

	return templateVolumes, nil
}

func UpdateImage(deployment interface{}, containerName, newimage string) (*api_apps.Deployment, error) {
	deploy := deployment.(*api_apps.Deployment)

	updated := false
	for i, v := range deploy.Spec.Template.Spec.Containers {
		if v.Name == containerName {
			deploy.Spec.Template.Spec.Containers[i].Image = newimage
			updated = true
			break
		}
	}
	if updated == false {
		return nil, fmt.Errorf(noContainer, containerName)
	}

	return deploy, nil
}

func (deploy *DeploymentKubeAPI) Validate() []error {
	var errs []error
	if deploy.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1123Label(deploy.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, deploy.Name, strings.Join(err, ",")))
	}
	if len(api_validation.IsInRange(deploy.Replicas, 0, maxDeployReplicas)) > 0 {
		errs = append(errs, fmt.Errorf(invalidReplicas, deploy.Replicas, maxDeployReplicas))
	}
	if deploy.Containers == nil || len(deploy.Containers) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Containers"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateContainer(container kube_types.Container, cpu, mem uint) []error {
	var errs []error
	if container.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1123Label(container.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, container.Name, strings.Join(err, ",")))
	}

	if cpu < minDeployCPU || cpu > maxDeployCPU {
		errs = append(errs, fmt.Errorf(invalidCPUQuota, cpu, minDeployCPU, maxDeployCPU))
	}

	if mem < minDeployMemory || mem > maxDeployMemory {
		errs = append(errs, fmt.Errorf(invalidMemoryQuota, mem, minDeployMemory, maxDeployMemory))
	}

	for _, v := range container.Ports {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.ports.name"))
		} else if err := api_validation.IsValidPortName(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.Protocol == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.ports.protocol"))
		} else if v.Protocol != kube_types.UDP && v.Protocol != kube_types.TCP {
			errs = append(errs, fmt.Errorf(invalidProtocol, v.Protocol))
		}
		if len(api_validation.IsValidPortNum(v.Port)) > 0 {
			errs = append(errs, fmt.Errorf(invalidPort, v.Port, minport, maxport))
		}
	}

	for _, v := range container.Env {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.env.name"))
		} else if err := api_validation.IsEnvVarName(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
	}

	for _, v := range container.VolumeMounts {
		if v.MountPath == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.volume_mounts.mount_path"))
		}
		if v.SubPath != nil && path.IsAbs(*v.SubPath) {
			errs = append(errs, fmt.Errorf(subPathRelative, *v.SubPath))
		}
		if v.PersistentVolumeClaimName == nil {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.volume_mounts.pvc_name"))
		}
	}

	for _, v := range container.ConfigMaps {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.config_maps.name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.MountPath == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "container.config_maps.mount_path"))
		}
		if v.SubPath != nil && path.IsAbs(*v.SubPath) {
			errs = append(errs, fmt.Errorf(subPathRelative, *v.SubPath))
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
