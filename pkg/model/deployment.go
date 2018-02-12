package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/inf.v0"
	api_apps "k8s.io/api/apps/v1"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const requestCoeffUnscaled = 5
const requestCoeffScale = 1

func ParseDeploymentList(deploys interface{}) []kube_types.Deployment {
	objects := deploys.(*api_apps.DeploymentList)
	var deployments []kube_types.Deployment
	for _, deployment := range objects.Items {
		deployment := ParseDeployment(&deployment)
		deployments = append(deployments, deployment)
	}
	return deployments
}

func ParseDeployment(deployment interface{}) kube_types.Deployment {
	obj := deployment.(*api_apps.Deployment)
	owner := obj.GetLabels()[ownerLabel]
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
	return kube_types.Deployment{
		Name:     obj.GetName(),
		Owner:    &owner,
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
	}
}

func MakeDeployment(nsName string, depl *kube_types.Deployment, containers []api_core.Container) *api_apps.Deployment {
	repl := int32(depl.Replicas)

	newDepl := api_apps.Deployment{}
	newDepl.Kind = "Deployment"
	newDepl.APIVersion = "apps/v1"
	newDepl.SetName(depl.Name)
	newDepl.SetNamespace(nsName)
	newDepl.SetLabels(map[string]string{"app": depl.Name, "owner": *depl.Owner})
	newDepl.Spec.Selector = &api_meta.LabelSelector{MatchLabels: map[string]string{"app": depl.Name, "owner": *depl.Owner}}
	newDepl.Spec.Replicas = &repl
	newDepl.Spec.Template.Spec.Containers = containers
	newDepl.Spec.Template.Spec.NodeSelector = map[string]string{"role": "slave"}
	newDepl.Spec.Template.SetLabels(map[string]string{"app": depl.Name, "owner": *depl.Owner})

	return &newDepl
}

func MakeContainers(containers []kube_types.Container) ([]api_core.Container, error) {
	var containersAfter []api_core.Container
	if len(containers) == 0 {
		return nil, ErrNoContainerInRequest
	}

	for _, container := range containers {
		err := binding.Validator.ValidateStruct(container)
		if err != nil {
			return nil, err
		}

		parsedContainer, err := MakeContainer(container)
		if err != nil {
			return nil, err
		}
		containersAfter = append(containersAfter, *parsedContainer)
	}
	return containersAfter, nil
}

func MakeContainer(container kube_types.Container) (*api_core.Container, error) {
	//Adding mounted volumes
	var mounts []api_core.VolumeMount
	if container.Volume != nil {
		for _, v := range *container.Volume {
			err := binding.Validator.ValidateStruct(v)
			if err != nil {
				return nil, err
			}

			var subpath string
			if v.SubPath != nil {
				subpath = *v.SubPath
			}
			mounts = append(mounts, api_core.VolumeMount{Name: v.Name, MountPath: v.MountPath, SubPath: subpath})
		}
	}

	//Adding enviroment variables
	var env []api_core.EnvVar
	if container.Env != nil {
		for _, v := range *container.Env {
			err := binding.Validator.ValidateStruct(v)
			if err != nil {
				return nil, err
			}
			env = append(env, api_core.EnvVar{Name: v.Name, Value: v.Value})
		}
	}

	//Adding ports
	var ports []api_core.ContainerPort
	if container.Ports != nil {
		for _, v := range *container.Ports {
			err := binding.Validator.ValidateStruct(v)
			if err != nil {
				return nil, err
			}
			ports = append(ports, api_core.ContainerPort{ContainerPort: int32(v.Port), Protocol: api_core.Protocol(v.Protocol), Name: v.Name})
		}
	}

	limits := make(map[api_core.ResourceName]api_resource.Quantity)

	var err error
	limits["cpu"], err = api_resource.ParseQuantity(container.Limits.CPU)
	if err != nil {
		return nil, ErrInvalidCPUFormat
	}
	limits["memory"], err = api_resource.ParseQuantity(container.Limits.Memory)
	if err != nil {
		return nil, ErrInvalidMemoryFormat
	}

	requests := make(map[api_core.ResourceName]api_resource.Quantity)
	reqCPU := limits["cpu"]
	reqMem := limits["memory"]

	//TODO Think how to divide Quantity values in adequate way
	requests["cpu"] = *api_resource.NewScaledQuantity(reqCPU.AsDec().Mul(reqCPU.AsDec(), inf.NewDec(requestCoeffUnscaled, requestCoeffScale)).UnscaledBig().Int64(), api_resource.Scale(0-reqCPU.AsDec().Scale()))
	requests["memory"] = *api_resource.NewScaledQuantity(reqMem.AsDec().Mul(reqMem.AsDec(), inf.NewDec(requestCoeffUnscaled, requestCoeffScale)).UnscaledBig().Int64(), api_resource.Scale(0-reqMem.AsDec().Scale()))

	var command []string
	if container.Command != nil {
		command = *container.Command
	}

	return &api_core.Container{
		Name:         container.Name,
		Image:        container.Image,
		VolumeMounts: mounts,
		Env:          env,
		Ports:        ports,
		Resources: api_core.ResourceRequirements{
			Limits:   limits,
			Requests: requests,
		},
		Command: command,
	}, nil
}
