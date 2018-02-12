package model

import (
	"git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
)

func ParsePodList(pods interface{}) []model.Pod {
	objects := pods.(*api_core.PodList)
	var pos []model.Pod
	for _, po := range objects.Items {
		pos = append(pos, ParsePod(&po))
	}
	return pos
}

func ParsePod(pod interface{}) model.Pod {
	obj := pod.(*api_core.Pod)
	owner := obj.GetLabels()[ownerLabel]
	containers := getContainers(obj.Spec.Containers)
	return model.Pod{
		Name:       obj.GetName(),
		Owner:      &owner,
		Containers: containers,
		Hostname:   &obj.Spec.Hostname,
		Status: &model.PodStatus{
			Phase: string(obj.Status.Phase),
		},
	}
}

func getContainers(cListi interface{}) []model.Container {
	cList := cListi.([]*api_core.Container)
	var containers []model.Container
	for _, c := range cList {
		env := getEnv(c.Env)
		volumes := getVolumes(c.VolumeMounts)

		cpu := c.Resources.Limits["cpu"]
		mem := c.Resources.Limits["memory"]

		containers = append(containers, model.Container{
			Name:   c.Name,
			Image:  c.Image,
			Env:    &env,
			Volume: &volumes,
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
	var volumes []model.Volume
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
	var envs []model.Env
	for _, e := range eList {
		envs = append(envs, model.Env{
			Name:  e.Name,
			Value: e.Value,
		})
	}
	return envs
}
