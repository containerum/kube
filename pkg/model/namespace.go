package model

import (
	v1 "k8s.io/api/core/v1"
)

const (
	ownerLabel = "owner"
)

type Namespace struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner_id,omitempty"`
	Resources Resources `json:"resources"`
}

type Resources struct {
	Hard Resource `json:"hard"`
	Used Resource `json:"used"`
}

type Resource struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

func ParseResourceQuotaList(quotas interface{}) []Namespace {
	objects := quotas.(*v1.ResourceQuotaList)
	var namespaces []Namespace
	for _, quota := range objects.Items {
		ns := ParseResourceQuota(&quota)
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

func ParseResourceQuota(quota interface{}) Namespace {
	obj := quota.(*v1.ResourceQuota)
	cpuLimit := obj.Spec.Hard[v1.ResourceLimitsCPU]
	memoryLimit := obj.Spec.Hard[v1.ResourceLimitsMemory]
	cpuUsed := obj.Status.Used[v1.ResourceLimitsCPU]
	memoryUsed := obj.Status.Used[v1.ResourceLimitsMemory]
	owner := obj.GetLabels()[ownerLabel]
	return Namespace{
		Name:  obj.GetNamespace(),
		Owner: owner,
		Resources: Resources{
			Hard: Resource{
				CPU:    cpuLimit.String(),
				Memory: memoryLimit.String(),
			},
			Used: Resource{
				CPU:    cpuUsed.String(),
				Memory: memoryUsed.String(),
			},
		},
	}
}
