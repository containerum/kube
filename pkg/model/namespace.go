package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ownerLabel = "owner"
	appLabel   = "app"
	nameLabel  = "name"
)

type NamespaceWithOwner struct {
	kube_types.Namespace
	Owner string `json:"owner,omitempty" binding:"required,uuid"`
}

// ParseResourceQuotaList parses kubernetes v1.ResourceQuotaList to more convenient []Namespace struct.
// (resource quouta contains all fields that parent namespace contains)
func ParseResourceQuotaList(quotas interface{}) ([]NamespaceWithOwner, error) {
	objects := quotas.(*api_core.ResourceQuotaList)
	if objects == nil {
		return nil, ErrUnableConvertNamespaceList
	}

	namespaces := make([]NamespaceWithOwner, 0)
	for _, quota := range objects.Items {
		ns, err := ParseResourceQuota(&quota)

		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, *ns)
	}
	return namespaces, nil
}

// ParseResourceQuota parses kubernetes v1.ResourceQuota to more convenient Namespace struct.
// (resource quouta contains all fields that parent namespace contains)
func ParseResourceQuota(quota interface{}) (*NamespaceWithOwner, error) {
	obj := quota.(*api_core.ResourceQuota)
	if obj == nil {
		return nil, ErrUnableConvertNamespace
	}

	cpuLimit := obj.Spec.Hard[api_core.ResourceLimitsCPU]
	memoryLimit := obj.Spec.Hard[api_core.ResourceLimitsMemory]
	cpuUsed := obj.Status.Used[api_core.ResourceLimitsCPU]
	memoryUsed := obj.Status.Used[api_core.ResourceLimitsMemory]
	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]

	return &NamespaceWithOwner{
		Owner: owner,
		Namespace: kube_types.Namespace{
			Name:    obj.GetNamespace(),
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
		},
	}, nil
}

// MakeResourceQuota creates kubernetes v1.ResourceQuota from cpu, memory, labels and namespace name
func MakeResourceQuota(cpu, memory string, labels map[string]string, ns string) (*api_core.ResourceQuota, error) {
	cpuq, err := api_resource.ParseQuantity(cpu)
	if err != nil {
		return nil, ErrInvalidCPUFormat
	}
	memoryq, err := api_resource.ParseQuantity(memory)
	if err != nil {
		return nil, ErrInvalidMemoryFormat
	}

	return &api_core.ResourceQuota{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "ResourceQuota",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      "quota",
			Namespace: ns,
		},
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

// MakeNamespace creates kubernetes v1.Namespace from Namespace struct
func MakeNamespace(ns NamespaceWithOwner) *api_core.Namespace {
	newNs := api_core.Namespace{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels: map[string]string{},
			Name:   ns.Name,
		},
		Spec: api_core.NamespaceSpec{},
	}
	if ns.Owner != "" {
		newNs.ObjectMeta.Labels[ownerLabel] = ns.Owner
	}

	return &newNs
}
