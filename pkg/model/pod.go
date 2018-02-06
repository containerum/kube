package model

import (
	"k8s.io/api/core/v1"
	"git.containerum.net/ch/kube-client/pkg/model"
)

func ParsePodList(pods interface{}) []model.Pod {
	objects := pods.(*v1.PodList)
	var pos []model.Pod
	for _, po := range objects.Items {
		pos = append(pos, ParsePod(&po))
	}
	return pos
}

func ParsePod(pod interface{}) model.Pod {
	obj := pod.(*v1.Pod)
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

func getContainers(cList []v1.Container) []model.Container {
	var containers []model.Container
	for _, c := range cList {
		env := getEnv(c.Env)
		volumes := getVolumes(c.VolumeMounts)
		containers = append(containers, model.Container{
			Name:   c.Name,
			Image:  c.Image,
			Env:    &env,
			Volume: &volumes,
		})
	}
	return containers
}

func getVolumes(vList []v1.VolumeMount) []model.Volume {
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

func getEnv(eList []v1.EnvVar) []model.Env {
	var envs []model.Env
	for _, e := range eList {
		envs = append(envs, model.Env{
			Name:  e.Name,
			Value: e.Value,
		})
	}
	return envs
}
