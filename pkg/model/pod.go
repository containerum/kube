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
		po := ParsePod(&po)
		pos = append(pos, po)
	}
	return pos
}

func ParsePod(pod interface{}) Pod {
	obj := pod.(*v1.Pod)
	return Pod{
		Name: obj.GetName(),
	}
}
