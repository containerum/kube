package model

import (
	json_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	ownerLabel = "owner"
)

func ParseResourceQuotaList(quotas interface{}) []json_types.Namespace {
	objects := quotas.(*api_core.ResourceQuotaList)
	var namespaces []json_types.Namespace
	for _, quota := range objects.Items {
		ns := ParseResourceQuota(&quota)
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

func ParseResourceQuota(quota interface{}) json_types.Namespace {
	obj := quota.(*api_core.ResourceQuota)
	cpuLimit := obj.Spec.Hard[api_core.ResourceLimitsCPU]
	memoryLimit := obj.Spec.Hard[api_core.ResourceLimitsMemory]
	cpuUsed := obj.Status.Used[api_core.ResourceLimitsCPU]
	memoryUsed := obj.Status.Used[api_core.ResourceLimitsMemory]
	owner := obj.GetLabels()[ownerLabel]
	return json_types.Namespace{
		Name:  obj.GetNamespace(),
		Owner: owner,
		Resources: json_types.Resources{
			Hard: json_types.Resource{
				CPU:    cpuLimit.String(),
				Memory: memoryLimit.String(),
			},
			Used: &json_types.Resource{
				CPU:    cpuUsed.String(),
				Memory: memoryUsed.String(),
			},
		},
	}
}

func MakeResourceQuota(cpu, memory api_resource.Quantity) *api_core.ResourceQuota {
	return &api_core.ResourceQuota{
		Spec: api_core.ResourceQuotaSpec{
			Hard: api_core.ResourceList{
				api_core.ResourceRequestsCPU:    cpu,
				api_core.ResourceLimitsCPU:      cpu,
				api_core.ResourceRequestsMemory: memory,
				api_core.ResourceLimitsMemory:   memory,
			},
		},
	}
}

func MakeNamespace(ns json_types.Namespace) *api_core.Namespace {
	newNs := api_core.Namespace{}
	newNs.Kind = "Namespace"
	newNs.APIVersion = "v1"
	newNs.Spec = api_core.NamespaceSpec{}
	newNs.ObjectMeta.Name = ns.Name
	newNs.ObjectMeta.Labels = make(map[string]string)
	newNs.ObjectMeta.Labels["name"] = ns.Name
	if ns.Owner != "" {
		newNs.ObjectMeta.Labels["owner"] = ns.Owner
	}

	return &newNs
}
