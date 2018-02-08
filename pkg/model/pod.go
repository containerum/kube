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

func getContainers(cList []api_core.Container) []model.Container {
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

func getVolumes(vList []api_core.VolumeMount) []model.Volume {
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

func getEnv(eList []api_core.EnvVar) []model.Env {
	var envs []model.Env
	for _, e := range eList {
		envs = append(envs, model.Env{
			Name:  e.Name,
			Value: e.Value,
		})
	}
	return envs
}

func getDeploymentVolumes(vList []api_core.Volume) []model.DeploymentVolume {
	var volumes []model.DeploymentVolume
	for _, v := range vList {
		volumes = append(volumes, model.DeploymentVolume{
			Name: v.Name,
			GlusterFS: model.GlusterFS{
				Endpoint: v.VolumeSource.Glusterfs.EndpointsName,
				Path:     v.VolumeSource.Glusterfs.Path,
			},
		})
	}
	return volumes
}
