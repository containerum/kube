package model

import (
	"git.containerum.net/ch/kube-client/pkg/model"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
)

type PodWithOwner struct {
	kube_types.Pod
	Owner string `json:"owner,omitempty" binding:"required,uuid"`
}

// ParsePodList parses kubernetes v1.PodList to more convenient []Pod struct.
func ParsePodList(pods interface{}) []PodWithOwner {
	objects := pods.(*api_core.PodList)
	var pos []PodWithOwner
	for _, po := range objects.Items {
		pos = append(pos, ParsePod(&po))
	}
	return pos
}

// ParsePod parses kubernetes v1.PodList to more convenient Pod struct.
func ParsePod(pod interface{}) PodWithOwner {
	obj := pod.(*api_core.Pod)
	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	containers := getContainers(obj.Spec.Containers)
	return PodWithOwner{
		Pod: model.Pod{
			Name:       obj.GetName(),
			Containers: containers,
			Hostname:   &obj.Spec.Hostname,
			Status: &model.PodStatus{
				Phase: string(obj.Status.Phase),
			},
		},
		Owner: owner,
	}
}

func getContainers(cListi interface{}) []model.Container {
	cList := cListi.([]api_core.Container)
	var containers []model.Container
	for _, c := range cList {
		env := getEnv(c.Env)
		volumes := getVolumes(c.VolumeMounts)

		cpu := c.Resources.Limits["cpu"]
		mem := c.Resources.Limits["memory"]

		containers = append(containers, model.Container{
			Name:    c.Name,
			Image:   c.Image,
			Env:     &env,
			Volume:  &volumes,
			Command: &c.Command,
			Limits: model.Limits{
				CPU:    cpu.String(),
				Memory: mem.String(),
			},
		})
	}
	return containers
}

func getVolumes(vListi interface{}) []model.Volume {
	vList := vListi.([]api_core.VolumeMount)
	volumes := make([]model.Volume, 0)
	for _, v := range vList {
		volumes = append(volumes, model.Volume{
			Name:      v.Name,
			MountPath: v.MountPath,
			SubPath:   &v.SubPath,
		})
	}
	return volumes
}

func getEnv(eListi interface{}) []model.Env {
	eList := eListi.([]api_core.EnvVar)
	envs := make([]model.Env, 0)
	for _, e := range eList {
		envs = append(envs, model.Env{
			Name:  e.Name,
			Value: e.Value,
		})
	}
	return envs
}
