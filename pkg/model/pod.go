package model

import (
	"strconv"

	"git.containerum.net/ch/kube-client/pkg/model"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

type PodsList struct {
	Pods []PodWithOwner `json:"pods"`
}

type PodWithOwner struct {
	kube_types.Pod
	Owner string `json:"owner,omitempty"`
}

// ParsePodList parses kubernetes v1.PodList to more convenient []Pod struct.
func ParsePodList(pods interface{}, parseforuser bool) *PodsList {
	objects := pods.(*api_core.PodList)
	pos := make([]PodWithOwner, 0)
	for _, po := range objects.Items {
		pos = append(pos, ParsePod(&po, parseforuser))
	}
	return &PodsList{pos}
}

// ParsePod parses kubernetes v1.PodList to more convenient Pod struct.
func ParsePod(pod interface{}, parseforuser bool) PodWithOwner {
	obj := pod.(*api_core.Pod)
	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	containers, _, _ := getContainers(obj.Spec.Containers, nil, 0)
	deploy := obj.GetObjectMeta().GetLabels()[appLabel]

	newPod := PodWithOwner{
		Pod: model.Pod{
			Deploy:     &deploy,
			Name:       obj.GetName(),
			Containers: containers,
			Hostname:   &obj.Spec.Hostname,
			Status: &model.PodStatus{
				Phase: string(obj.Status.Phase),
			},
		},
		Owner: owner,
	}

	if parseforuser {
		newPod.Owner = ""
	}

	return newPod
}

func getContainers(cListi interface{}, mode map[string]int32, replicas int) (containers []model.Container, totalcpu api_resource.Quantity, totalmem api_resource.Quantity) {
	cList := cListi.([]api_core.Container)
	for _, c := range cList {
		env := getEnv(c.Env)
		volumes, configMaps := getVolumes(c.VolumeMounts, mode)

		cpu := c.Resources.Limits["cpu"]
		mem := c.Resources.Limits["memory"]

		for i := 0; i < replicas; i++ {
			totalcpu.Add(c.Resources.Limits["cpu"])
			totalmem.Add(c.Resources.Limits["memory"])
		}

		containers = append(containers, model.Container{
			Name:         c.Name,
			Image:        c.Image,
			Env:          env,
			VolumeMounts: volumes,
			ConfigMaps:   configMaps,
			Commands:     c.Command,
			Limits: model.Resource{
				CPU:    cpu.String(),
				Memory: mem.String(),
			},
		})
	}
	return containers, totalcpu, totalmem
}

func getVolumes(vListi interface{}, mode map[string]int32) ([]model.ContainerVolume, []model.ContainerVolume) {
	vList := vListi.([]api_core.VolumeMount)
	volumes := make([]model.ContainerVolume, 0)
	configMaps := make([]model.ContainerVolume, 0)
	for _, v := range vList {

		subpath := v.SubPath
		newvol := model.ContainerVolume{
			Name:      v.Name,
			MountPath: v.MountPath,
		}

		if subpath != "" {
			newvol.SubPath = &subpath
		}

		mode, ok := mode[v.Name]
		if ok {
			formated := strconv.FormatInt(int64(mode), 8)
			newvol.Mode = &formated
			configMaps = append(configMaps, newvol)
		} else {
			volumes = append(volumes, newvol)
		}
	}
	return volumes, configMaps
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
