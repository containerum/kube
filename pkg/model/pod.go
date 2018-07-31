package model

import (
	"strconv"

	"time"

	"strings"

	"github.com/containerum/kube-client/pkg/model"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

// ParseKubePodList parses kubernetes v1.PodList to more convenient []Pod struct.
func ParseKubePodList(pods interface{}, parseforuser bool) *kube_types.PodsList {
	podList := pods.(*api_core.PodList)
	ret := make([]kube_types.Pod, 0)
	for _, po := range podList.Items {
		ret = append(ret, ParseKubePod(&po, parseforuser))
	}
	return &kube_types.PodsList{Pods: ret}
}

// ParseKubePod parses kubernetes v1.PodList to more convenient Pod struct.
func ParseKubePod(pod interface{}, parseforuser bool) kube_types.Pod {
	obj := pod.(*api_core.Pod)
	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	containers, cpu, mem := getContainers(obj.Spec.Containers, nil, nil, 1)
	deploy := obj.GetObjectMeta().GetLabels()[appLabel]
	createdAt := obj.ObjectMeta.CreationTimestamp.UTC().Format(time.RFC3339)

	newPod := kube_types.Pod{
		CreatedAt:  &createdAt,
		Deploy:     &deploy,
		Name:       obj.GetName(),
		Containers: containers,
		Status: &model.PodStatus{
			Phase: string(obj.Status.Phase),
		},
		ImagePullSecrets: getImagePullSecrets(obj.Spec.ImagePullSecrets),
		TotalCPU:         uint(cpu.ScaledValue(api_resource.Milli)),
		TotalMemory:      uint(mem.Value() / 1024 / 1024),
		Owner:            owner,
	}

	if parseforuser {
		newPod.Mask()
	}

	return newPod
}

func getContainers(cListi interface{}, mode map[string]int32, storageName map[string]string, replicas int) (containers []model.Container, totalcpu, totalmem api_resource.Quantity) {
	cList := cListi.([]api_core.Container)
	for _, c := range cList {
		env := getEnv(c.Env)
		volumes, configMaps := getVolumes(c.VolumeMounts, mode, storageName)

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
				CPU:    uint(cpu.ScaledValue(api_resource.Milli)),
				Memory: uint(mem.Value() / 1024 / 1024),
			},
		})
	}
	return containers, totalcpu, totalmem
}

func getVolumes(vListi interface{}, mode map[string]int32, storageName map[string]string) ([]model.ContainerVolume, []model.ContainerVolume) {
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
			newvol.Name = strings.TrimRight(newvol.Name, "-cm")
			configMaps = append(configMaps, newvol)
		} else {
			storage, ok := storageName[v.Name]
			if ok {
				newvol.Name = storage
			}
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
