package model

import (
	"strconv"

	"time"

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

// ParseKubePodList parses kubernetes v1.PodList to more convenient []Pod struct.
func ParseKubePodList(pods interface{}, parseforuser bool) *PodsList {
	podList := pods.(*api_core.PodList)
	ret := make([]PodWithOwner, 0)
	for _, po := range podList.Items {
		ret = append(ret, ParseKubePod(&po, parseforuser))
	}
	return &PodsList{ret}
}

// ParseKubePod parses kubernetes v1.PodList to more convenient Pod struct.
func ParseKubePod(pod interface{}, parseforuser bool) PodWithOwner {
	obj := pod.(*api_core.Pod)
	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	containers, cpu, mem := getContainers(obj.Spec.Containers, nil, 1)
	deploy := obj.GetObjectMeta().GetLabels()[appLabel]
	createdAt := obj.ObjectMeta.CreationTimestamp.Format(time.RFC3339)

	newPod := PodWithOwner{
		Pod: model.Pod{
			CreatedAt:   &createdAt,
			TotalMemory: mem.String(),
			TotalCPU:    cpu.String(),
			Deploy:      &deploy,
			Name:        obj.GetName(),
			Containers:  containers,
			Hostname:    &obj.Spec.Hostname,
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
