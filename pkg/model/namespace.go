package model

import (
	api_core "k8s.io/api/core/v1"
	"git.containerum.net/ch/kube-client/pkg/model"
)

const (
	ownerLabel = "owner"
)

func ParseResourceQuotaList(quotas interface{}) []model.Namespace {
	objects := quotas.(*api_core.ResourceQuotaList)
	var namespaces []model.Namespace
	for _, quota := range objects.Items {
		ns := ParseResourceQuota(&quota)
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

func ParseResourceQuota(quota interface{}) model.Namespace {
	obj := quota.(*api_core.ResourceQuota)
	cpuLimit := obj.Spec.Hard[api_core.ResourceLimitsCPU]
	memoryLimit := obj.Spec.Hard[api_core.ResourceLimitsMemory]
	cpuUsed := obj.Status.Used[api_core.ResourceLimitsCPU]
	memoryUsed := obj.Status.Used[api_core.ResourceLimitsMemory]
	owner := obj.GetLabels()[ownerLabel]
	return model.Namespace{
		Name:  obj.GetNamespace(),
		Owner: &owner,
		Resources: model.Resources{
			Hard: model.Resource{
				CPU:    cpuLimit.String(),
				Memory: memoryLimit.String(),
			},
			Used: &model.Resource{
				CPU:    cpuUsed.String(),
				Memory: memoryUsed.String(),
			},
		},
	}
}
