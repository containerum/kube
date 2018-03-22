package model

import (
	"crypto/sha256"
	"fmt"

	"strconv"

	"strings"

	"path"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	"github.com/pkg/errors"
	"gopkg.in/inf.v0"
	api_apps "k8s.io/api/apps/v1"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

const requestCoeffUnscaled = 5
const requestCoeffScale = 1

const glusterFSEndpoint = "ch-glusterfs"

const (
	minDeployCPU    = "10m"
	minDeployMemory = "10Mi"
	maxDeployCPU    = "4"
	maxDeployMemory = "4Gi"

	maxDeployReplicas = 10
)

type DeploymentsList struct {
	Deployments []DeploymentWithOwner `json:"deployments"`
}

type DeploymentWithOwner struct {
	kube_types.Deployment
	Owner string `json:"owner,omitempty"`
}

// ParseDeploymentList parses kubernetes v1.DeploymentList to more convenient []Deployment struct
func ParseDeploymentList(deploys interface{}, parseforuser bool) (*DeploymentsList, error) {
	objects := deploys.(*api_apps.DeploymentList)
	if objects == nil {
		return nil, ErrUnableConvertDeploymentList
	}

	deployments := make([]DeploymentWithOwner, 0)
	for _, deployment := range objects.Items {
		deployment, err := ParseDeployment(&deployment, parseforuser)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, *deployment)
	}
	return &DeploymentsList{deployments}, nil
}

// ParseDeployment parses kubernetes v1.Deployment to more convenient Deployment struct
func ParseDeployment(deployment interface{}, parseforuser bool) (*DeploymentWithOwner, error) {
	obj := deployment.(*api_apps.Deployment)
	if obj == nil {
		return nil, ErrUnableConvertDeployment
	}

	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	replicas := 0
	containers := getContainers(obj.Spec.Template.Spec.Containers, getVolumeMode(obj.Spec.Template.Spec.Volumes))
	updated := obj.ObjectMeta.CreationTimestamp.Unix()
	if r := obj.Spec.Replicas; r != nil {
		replicas = int(*r)
	}
	for _, c := range obj.Status.Conditions {
		if t := c.LastUpdateTime.Unix(); t > updated {
			updated = t
		}
	}

	newDeploy := DeploymentWithOwner{
		Deployment: kube_types.Deployment{
			Name:     obj.GetName(),
			Replicas: replicas,
			Status: &kube_types.DeploymentStatus{
				CreatedAt:           obj.ObjectMeta.CreationTimestamp.Unix(),
				UpdatedAt:           updated,
				Replicas:            int(obj.Status.Replicas),
				ReadyReplicas:       int(obj.Status.ReadyReplicas),
				AvailableReplicas:   int(obj.Status.AvailableReplicas),
				UpdatedReplicas:     int(obj.Status.UpdatedReplicas),
				UnavailableReplicas: int(obj.Status.UnavailableReplicas),
			},
			Containers: containers,
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

//MakeDeployment creates kubernetes v1.Deployment from Deployment struct and namespace labels
func MakeDeployment(nsName string, depl DeploymentWithOwner, labels map[string]string) (*api_apps.Deployment, []error) {
	err := ValidateDeployment(depl)
	if err != nil {
		return nil, err
	}

	repl := int32(depl.Replicas)
	containers, volumes, cmaps, err := makeContainers(depl.Containers)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = depl.Name
	labels[ownerLabel] = depl.Owner
	labels[nameLabel] = depl.Name

	deployment := api_apps.Deployment{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      depl.Name,
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
					Volumes: makeTemplateVolumes(volumes, cmaps, depl.Owner),
				},
				ObjectMeta: api_meta.ObjectMeta{
					Labels: labels,
				},
			},
		},
	}

	return &deployment, nil
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

		rq, err := makeContainerResourceQuota(c.Limits.CPU, c.Limits.Memory)
		if err != nil {
			return nil, nil, nil, []error{err}
		}
		container.Resources = *rq

		errs := ValidateContainer(c, *container.Resources.Limits.Cpu(), *container.Resources.Limits.Memory())
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

func makeContainerResourceQuota(cpu string, memory string) (*api_core.ResourceRequirements, error) {
	limits := make(map[api_core.ResourceName]api_resource.Quantity)

	var err error
	limits["cpu"], err = api_resource.ParseQuantity(cpu)
	if err != nil {
		return nil, ErrInvalidCPUFormat
	}
	limits["memory"], err = api_resource.ParseQuantity(memory)
	if err != nil {
		return nil, ErrInvalidMemoryFormat
	}

	requests := make(map[api_core.ResourceName]api_resource.Quantity)
	reqCPU := limits["cpu"]
	reqMem := limits["memory"]

	//TODO Think how to divide Quantity values in adequate way
	requests["cpu"] = *api_resource.NewScaledQuantity(reqCPU.AsDec().Mul(reqCPU.AsDec(), inf.NewDec(requestCoeffUnscaled, requestCoeffScale)).UnscaledBig().Int64(), api_resource.Scale(0-reqCPU.AsDec().Scale()))
	requests["memory"] = *api_resource.NewScaledQuantity(reqMem.AsDec().Mul(reqMem.AsDec(), inf.NewDec(requestCoeffUnscaled, requestCoeffScale)).UnscaledBig().Int64(), api_resource.Scale(0-reqMem.AsDec().Scale()))

	return &api_core.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}, nil
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

func ValidateDeployment(deploy DeploymentWithOwner) []error {
	errs := []error{}
	if deploy.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else {
		if !IsValidUUID(deploy.Owner) {
			errs = append(errs, errors.New(invalidOwner))
		}
	}
	if deploy.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(deploy.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, deploy.Name, strings.Join(err, ","))))
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

func ValidateContainer(container kube_types.Container, cpu, mem api_resource.Quantity) []error {
	errs := []error{}

	mincpu, _ := api_resource.ParseQuantity(minDeployCPU)
	maxcpu, _ := api_resource.ParseQuantity(maxDeployCPU)
	minmem, _ := api_resource.ParseQuantity(minDeployMemory)
	maxmem, _ := api_resource.ParseQuantity(maxDeployMemory)

	if container.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(container.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, container.Name, strings.Join(err, ","))))
	}

	if cpu.Cmp(mincpu) == -1 || cpu.Cmp(maxcpu) == 1 {
		errs = append(errs, fmt.Errorf(invalidCPUQuota, cpu.String(), minDeployCPU, maxDeployCPU))
	}

	if mem.Cmp(minmem) == -1 || mem.Cmp(maxmem) == 1 {
		errs = append(errs, fmt.Errorf(invalidMemoryQuota, mem.String(), minDeployMemory, maxDeployMemory))
	}

	for _, v := range container.Ports {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Port: Name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, errors.New(fmt.Sprintf(invalidName, v.Name, strings.Join(err, ","))))
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
			errs = append(errs, errors.New(fmt.Sprintf(invalidName, v.Name, strings.Join(err, ","))))
		}
	}

	for _, v := range container.VolumeMounts {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Volume: Name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, errors.New(fmt.Sprintf(invalidName, v.Name, strings.Join(err, ","))))
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
			errs = append(errs, errors.New(fmt.Sprintf(invalidName, v.Name, strings.Join(err, ","))))
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
