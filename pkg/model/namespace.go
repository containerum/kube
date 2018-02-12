package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	ownerLabel = "owner"
)

func ParseResourceQuotaList(quotas interface{}) []kube_types.Namespace {
	objects := quotas.(*api_core.ResourceQuotaList)
	var namespaces []kube_types.Namespace
	for _, quota := range objects.Items {
		ns := ParseResourceQuota(&quota)
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

func ParseResourceQuota(quota interface{}) kube_types.Namespace {
	obj := quota.(*api_core.ResourceQuota)
	cpuLimit := obj.Spec.Hard[api_core.ResourceLimitsCPU]
	memoryLimit := obj.Spec.Hard[api_core.ResourceLimitsMemory]
	cpuUsed := obj.Status.Used[api_core.ResourceLimitsCPU]
	memoryUsed := obj.Status.Used[api_core.ResourceLimitsMemory]
	owner := obj.GetLabels()[ownerLabel]
	return kube_types.Namespace{
		Name:    obj.GetNamespace(),
		Owner:   owner,
		Created: obj.ObjectMeta.CreationTimestamp.Unix(),
		Resources: kube_types.Resources{
			Hard: kube_types.Resource{
				CPU:    cpuLimit.String(),
				Memory: memoryLimit.String(),
			},
			Used: &kube_types.Resource{
				CPU:    cpuUsed.String(),
				Memory: memoryUsed.String(),
			},
		},
	}
}

func MakeResourceQuota(cpu, memory string) (*api_core.ResourceQuota, error) {
	cpuq, err := api_resource.ParseQuantity(cpu)
	if err != nil {
		return nil, ErrInvalidCPUFormat
	}

	memoryq, err := api_resource.ParseQuantity(memory)
	if err != nil {
		return nil, ErrInvalidMemoryFormat
	}

	return &api_core.ResourceQuota{
		Spec: api_core.ResourceQuotaSpec{
			Hard: api_core.ResourceList{
				api_core.ResourceRequestsCPU:    cpuq,
				api_core.ResourceLimitsCPU:      cpuq,
				api_core.ResourceRequestsMemory: memoryq,
				api_core.ResourceLimitsMemory:   memoryq,
			},
		},
	}, nil
}

func MakeNamespace(ns kube_types.Namespace) *api_core.Namespace {
	newNs := api_core.Namespace{}
	newNs.Kind = "Namespace"
	newNs.APIVersion = "v1"
	newNs.Spec = api_core.NamespaceSpec{}
	newNs.ObjectMeta.Name = ns.Name
	newNs.ObjectMeta.Labels = make(map[string]string)
	if ns.Owner != "" {
		newNs.ObjectMeta.Labels["owner"] = ns.Owner
	}

	return &newNs
}
