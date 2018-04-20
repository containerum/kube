package model

import (
	"crypto/sha256"
	"fmt"

	"strconv"

	"path"
	"strings"

	"time"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
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

	glusterFSEndpoint = "ch-glusterfs"

	minDeployCPU      = 10   //m
	minDeployMemory   = 10   //Mi
	maxDeployCPU      = 3000 //m
	maxDeployMemory   = 8000 //Mi
	maxDeployReplicas = 15
)

type DeploymentsList struct {
	Deployments []DeploymentWithOwner `json:"deployments"`
}

type DeploymentWithOwner struct {
	kube_types.Deployment
	Owner string `json:"owner,omitempty"`
}

// ParseKubeDeploymentList parses kubernetes v1.DeploymentList to more convenient []Deployment struct
func ParseKubeDeploymentList(deploys interface{}, parseforuser bool) (*DeploymentsList, error) {
	deployList := deploys.(*api_apps.DeploymentList)
	if deployList == nil {
		return nil, ErrUnableConvertDeploymentList
	}

	deployments := make([]DeploymentWithOwner, 0)
	for _, deployment := range deployList.Items {
		deployment, err := ParseKubeDeployment(&deployment, parseforuser)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, *deployment)
	}
	return &DeploymentsList{deployments}, nil
}

// ParseKubeDeployment parses kubernetes v1.Deployment to more convenient Deployment struct
func ParseKubeDeployment(deployment interface{}, parseforuser bool) (*DeploymentWithOwner, error) {
	deploy := deployment.(*api_apps.Deployment)
	if deploy == nil {
		return nil, ErrUnableConvertDeployment
	}

	owner := deploy.GetObjectMeta().GetLabels()[ownerLabel]
	replicas := 0
	if r := deploy.Spec.Replicas; r != nil {
		replicas = int(*r)
	}
	containers, totalcpu, totalmem := getContainers(deploy.Spec.Template.Spec.Containers, getVolumeMode(deploy.Spec.Template.Spec.Volumes), replicas)
	updated := deploy.ObjectMeta.CreationTimestamp
	for _, c := range deploy.Status.Conditions {
		if c.LastUpdateTime.After(updated.Time) {
			updated = c.LastUpdateTime
		}
	}

	newDeploy := DeploymentWithOwner{
		Deployment: kube_types.Deployment{
			Name:     deploy.GetName(),
			Replicas: replicas,
			Status: &kube_types.DeploymentStatus{
				CreatedAt:           deploy.ObjectMeta.CreationTimestamp.Format(time.RFC3339),
				UpdatedAt:           updated.Format(time.RFC3339),
				Replicas:            int(deploy.Status.Replicas),
				ReadyReplicas:       int(deploy.Status.ReadyReplicas),
				AvailableReplicas:   int(deploy.Status.AvailableReplicas),
				UpdatedReplicas:     int(deploy.Status.UpdatedReplicas),
				UnavailableReplicas: int(deploy.Status.UnavailableReplicas),
			},
			Containers:  containers,
			TotalCPU:    uint(totalcpu.ScaledValue(api_resource.Milli)),
			TotalMemory: uint(totalmem.Value() / 1024 / 1024),
		},
		Owner: owner,
	}

	if parseforuser {
		newDeploy.Owner = ""
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

//ToKube creates kubernetes v1.Deployment from Deployment struct and namespace labels
func (deploy *DeploymentWithOwner) ToKube(nsName string, labels map[string]string) (*api_apps.Deployment, []error) {
	err := deploy.Validate()
	if err != nil {
		return nil, err
	}

	repl := int32(deploy.Replicas)
	containers, volumes, cmaps, err := makeContainers(deploy.Containers)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = deploy.Name
	labels[ownerLabel] = deploy.Owner

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
			Template: api_core.PodTemplateSpec{
				Spec: api_core.PodSpec{
					Containers: containers,
					NodeSelector: map[string]string{
						"role": "slave",
					},
					Volumes: makeTemplateVolumes(volumes, cmaps, deploy.Owner),
				},
				ObjectMeta: api_meta.ObjectMeta{
					Labels: labels,
				},
			},
		},
	}

	return &newDeploy, nil
}

func makeContainers(containers []kube_types.Container) ([]api_core.Container, []string, map[string]int64, []error) {
	var containersAfter []api_core.Container

	volumes := make([]string, 0)
	cmaps := make(map[string]int64, 0)
	for _, c := range containers {
		container := api_core.Container{
			Name:    c.Name,
			Image:   c.Image,
			Command: makeContainerCommands(c.Commands),
		}

		if c.VolumeMounts != nil || c.ConfigMaps != nil {
			vm, vnames, cmnames := makeContainerVolumes(c.VolumeMounts, c.ConfigMaps)
			volumes = append(volumes, vnames...)
			for k, v := range cmnames {
				cmaps[k] = v
			}
			container.VolumeMounts = vm
		}

		if c.Env != nil {
			container.Env = makeContainerEnv(c.Env)
		}

		if c.Ports != nil {
			container.Ports = makeContainerPorts(c.Ports)
		}

		rq := makeContainerResourceQuota(c.Limits.CPU, c.Limits.Memory)

		container.Resources = *rq

		errs := validateContainer(c, c.Limits.CPU, c.Limits.Memory)
		if errs != nil {
			return nil, nil, nil, errs
		}

		containersAfter = append(containersAfter, container)
	}
	return containersAfter, volumes, cmaps, nil
}

func makeContainerVolumes(volumes []kube_types.ContainerVolume, configMaps []kube_types.ContainerVolume) ([]api_core.VolumeMount, []string, map[string]int64) {
	mounts := make([]api_core.VolumeMount, 0)
	vnames := make([]string, 0)
	cmnames := make(map[string]int64, 0)
	if volumes != nil {
		for _, v := range volumes {
			var subpath string

			if v.SubPath != nil {
				subpath = *v.SubPath
			}
			vnames = append(vnames, v.Name)
			mounts = append(mounts, api_core.VolumeMount{Name: v.Name, MountPath: v.MountPath, SubPath: subpath})
		}
	}
	if configMaps != nil {
		for _, v := range configMaps {
			var subpath string
			if v.SubPath != nil {
				subpath = *v.SubPath
			}

			mode := int64(0644)
			if v.Mode != nil {
				if newMode, err := strconv.ParseInt(*v.Mode, 8, 32); err == nil {
					mode = newMode
				}
			}
			cmnames[v.Name] = mode
			mounts = append(mounts, api_core.VolumeMount{Name: v.Name, MountPath: v.MountPath, SubPath: subpath})
		}
	}

	return mounts, vnames, cmnames
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

func makeTemplateVolumes(volumes []string, cmaps map[string]int64, owner string) []api_core.Volume {
	tvolumes := make([]api_core.Volume, 0)
	if len(volumes) != 0 {
		for _, v := range volumes {
			newVolume := api_core.Volume{
				Name: v,
				VolumeSource: api_core.VolumeSource{
					Glusterfs: &api_core.GlusterfsVolumeSource{
						EndpointsName: glusterFSEndpoint,
						Path:          fmt.Sprintf("cli_%x", (sha256.Sum256([]byte(v + owner)))),
					},
				},
			}
			tvolumes = append(tvolumes, newVolume)
		}
	}
	if len(cmaps) != 0 {
		for k, v := range cmaps {
			mode := int32(v)
			newVolume := api_core.Volume{
				Name: k,
				VolumeSource: api_core.VolumeSource{
					ConfigMap: &api_core.ConfigMapVolumeSource{
						LocalObjectReference: api_core.LocalObjectReference{
							Name: k,
						},
						DefaultMode: &mode,
					},
				},
			}
			tvolumes = append(tvolumes, newVolume)
		}
	}
	return tvolumes
}

func (deploy *DeploymentWithOwner) Validate() []error {
	errs := []error{}
	if deploy.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else if !IsValidUUID(deploy.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}
	if deploy.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(deploy.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, deploy.Name, strings.Join(err, ",")))
	}
	if len(api_validation.IsInRange(deploy.Replicas, 1, maxDeployReplicas)) > 0 {
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
	errs := []error{}

	if container.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
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
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Port: Name"))
		} else if err := api_validation.IsValidPortName(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.Protocol != kube_types.UDP && v.Protocol != kube_types.TCP {
			errs = append(errs, fmt.Errorf(invalidProtocol, v.Protocol))
		}
		if len(api_validation.IsValidPortNum(v.Port)) > 0 {
			errs = append(errs, fmt.Errorf(invalidPort, v.Port, minport, maxport))
		}
	}

	for _, v := range container.Env {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Env: Name"))
		} else if err := api_validation.IsEnvVarName(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
	}

	for _, v := range container.VolumeMounts {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Volume: Name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.MountPath == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Volume: Mount path"))
		}
		if v.SubPath != nil && path.IsAbs(*v.SubPath) {
			errs = append(errs, fmt.Errorf(subPathRelative, *v.SubPath))
		}
	}

	for _, v := range container.ConfigMaps {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "ConfigMap: Name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.MountPath == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "ConfigMap: Mount path"))
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
