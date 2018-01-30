package model

import (
	"k8s.io/api/core/v1"
)

type Pod struct {
	Name            string            `json:"name"`
	Owner           string            `json:"owner_id,omitempty"`
	Containers      []Container       `json:"containers"`
	ImagePullSecret map[string]string `json:"image_pull_secret,omitempty"`
	Status          PodStatus         `json:"status,omitempty"`
	Hostname        string            `json:"hostname,omitempty"`
}

type PodStatus struct {
	Phase string `json:"phase"`
}

type Container struct {
	Name   string   `json:"name"`
	Env    []Env    `json:"env,omitempty"`
	Image  string   `json:"image"`
	Volume []Volume `json:"volume,omitempty"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Volume struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
	SubPath   string `json:"sub_path,omitempty"`
}

func ParsePodList(pods interface{}) []Pod {
	objects := pods.(*v1.PodList)
	var pos []Pod
	for _, po := range objects.Items {
		pos = append(pos, ParsePod(&po))
	}
	return pos
}

func ParsePod(pod interface{}) Pod {
	obj := pod.(*v1.Pod)
	owner := obj.GetLabels()[ownerLabel]
	containers := getContainers(obj.Spec.Containers)
	return Pod{
		Name:       obj.GetName(),
		Owner:      owner,
		Containers: containers,
		Hostname:   obj.Spec.Hostname,
		Status: PodStatus{
			Phase: string(obj.Status.Phase),
		},
	}
}

func getContainers(cList []v1.Container) []Container {
	var containers []Container
	for _, c := range cList {
		containers = append(containers, Container{
			Name:   c.Name,
			Image:  c.Image,
			Env:    getEnv(c.Env),
			Volume: getVolumes(c.VolumeMounts),
		})
	}
	return containers
}

func getVolumes(vList []v1.VolumeMount) []Volume {
	var volumes []Volume
	for _, v := range vList {
		volumes = append(volumes, Volume{
			Name:      v.Name,
			MountPath: v.MountPath,
			SubPath:   v.SubPath,
		})
	}
	return volumes
}

func getEnv(eList []v1.EnvVar) []Env {
	var envs []Env
	for _, e := range eList {
		envs = append(envs, Env{
			Name:  e.Name,
			Value: e.Value,
		})
	}
	return envs
}
