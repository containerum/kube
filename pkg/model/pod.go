package model

import (
	"strconv"

	"git.containerum.net/ch/kube-client/pkg/model"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
)

type PodWithOwner struct {
	kube_types.Pod
	Owner string `json:"owner,omitempty"`
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
	containers := getContainers(obj.Spec.Containers, nil)
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

func getContainers(cListi interface{}, mode map[string]int32) []model.Container {
	cList := cListi.([]api_core.Container)
	var containers []model.Container
	for _, c := range cList {
		env := getEnv(c.Env)
		volumes, configMaps := getVolumes(c.VolumeMounts, mode)

		cpu := c.Resources.Limits["cpu"]
		mem := c.Resources.Limits["memory"]

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
	return containers
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
