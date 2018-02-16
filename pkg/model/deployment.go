package model

import (
	"crypto/sha256"
	"fmt"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	"gopkg.in/inf.v0"
	api_apps "k8s.io/api/apps/v1"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const requestCoeffUnscaled = 5
const requestCoeffScale = 1

const glusterFSEndpoint = "ch-glusterfs"

type DeploymentWithOwner struct {
	kube_types.Deployment
	Owner string `json:"owner,omitempty" binding:"required,uuid"`
}

// ParseDeploymentList parses kubernetes v1.DeploymentList to more convenient []Deployment struct
func ParseDeploymentList(deploys interface{}) ([]DeploymentWithOwner, error) {
	objects := deploys.(*api_apps.DeploymentList)
	if objects == nil {
		return nil, ErrUnableConvertDeploymentList
	}

	deployments := make([]DeploymentWithOwner, 0)
	for _, deployment := range objects.Items {
		deployment, err := ParseDeployment(&deployment)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, *deployment)
	}
	return deployments, nil
}

// ParseDeployment parses kubernetes v1.Deployment to more convenient Deployment struct
func ParseDeployment(deployment interface{}) (*DeploymentWithOwner, error) {
	obj := deployment.(*api_apps.Deployment)
	if obj == nil {
		return nil, ErrUnableConvertDeployment
	}

	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	replicas := 0
	containers := getContainers(obj.Spec.Template.Spec.Containers)
	updated := obj.ObjectMeta.CreationTimestamp.Unix()
	if r := obj.Spec.Replicas; r != nil {
		replicas = int(*r)
	}
	for _, c := range obj.Status.Conditions {
		if t := c.LastUpdateTime.Unix(); t > updated {
			updated = t
		}
	}
	return &DeploymentWithOwner{
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
			Hostname:   &obj.Spec.Template.Spec.Hostname,
		},
		Owner: owner,
	}, nil
}

//MakeDeployment creates kubernetes v1.Deployment from Deployment struct and namespace labels
func MakeDeployment(nsName string, depl *DeploymentWithOwner, labels map[string]string) (*api_apps.Deployment, error) {
	repl := int32(depl.Replicas)
	containers, volumes, err := makeContainers(depl.Containers)
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
					Volumes: makeTemplateVolumes(volumes, depl.Owner),
				},
				ObjectMeta: api_meta.ObjectMeta{
					Labels: labels,
				},
			},
		},
	}

	return &deployment, nil
}

func makeContainers(containers []kube_types.Container) ([]api_core.Container, []string, error) {
	var containersAfter []api_core.Container
	if len(containers) == 0 {
		return nil, nil, ErrNoContainerInRequest
	}

	volumes := make([]string, 0)
	for _, c := range containers {
		container := api_core.Container{
			Name:    c.Name,
			Image:   c.Image,
			Command: makeContainerCommands(c.Command),
		}

		if c.Volume != nil {
			vm, vnames := makeContainerVolumes(*c.Volume)
			volumes = append(volumes, vnames...)
			container.VolumeMounts = vm
		}

		if c.Env != nil {
			container.Env = makeContainerEnv(*c.Env)
		}

		if c.Ports != nil {
			container.Ports = makeContainerPorts(*c.Ports)
		}

		if rq, err := makeContainerResourceQuota(c.Limits.CPU, c.Limits.Memory); err != nil {
			return nil, nil, err
		} else {
			container.Resources = *rq
		}

		containersAfter = append(containersAfter, container)
	}
	return containersAfter, volumes, nil
}

func makeContainerVolumes(volumes []kube_types.Volume) ([]api_core.VolumeMount, []string) {
	mounts := make([]api_core.VolumeMount, 0)
	vnames := make([]string, 0)
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
	return mounts, vnames
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

func makeContainerPorts(ports []kube_types.Port) []api_core.ContainerPort {
	contports := make([]api_core.ContainerPort, 0)
	if ports != nil {
		for _, v := range ports {
			contports = append(contports, api_core.ContainerPort{ContainerPort: int32(v.Port), Protocol: api_core.Protocol(v.Protocol), Name: v.Name})
		}
	}
	return contports
}

func makeContainerCommands(commands *[]string) []string {
	contcommands := make([]string, 0)
	if commands != nil {
		contcommands = *commands
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

func makeTemplateVolumes(volumes []string, owner string) []api_core.Volume {
	tvolumes := make([]api_core.Volume, 0)
	if volumes != nil {
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
	return tvolumes
}
