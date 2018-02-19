package model

import (
	"fmt"
	"net/http"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

const (
	ownerLabel = "owner"
	appLabel   = "app"
	nameLabel  = "name"
)

const (
	minNamespaceCPU    = "0.3"
	minNamespaceMemory = "0.5Gi"
	maxNamespaceCPU    = "10"
	maxNamespaceMemory = "10Gi"
)

type NamespaceWithOwner struct {
	kube_types.Namespace
	Owner string `json:"owner,omitempty" binding:"omitempty,uuid"`
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
func MakeResourceQuota(cpu, memory string, labels map[string]string, ns string) (*api_core.ResourceQuota, []error) {
	cpuq, err := api_resource.ParseQuantity(cpu)
	if err != nil {
		return nil, []error{ErrInvalidCPUFormat}
	}
	memoryq, err := api_resource.ParseQuantity(memory)
	if err != nil {
		return nil, []error{ErrInvalidMemoryFormat}
	}

	errors := validateResourceQuota(cpuq, memoryq)
	if errors != nil {
		return nil, errors
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
func MakeNamespace(ns NamespaceWithOwner) (*api_core.Namespace, error) {
	err := validateNamespace(ns.Namespace)
	if err != nil {
		return nil, err
	}

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

	return &newNs, nil
}

func validateNamespace(ns kube_types.Namespace) error {
	if len(api_validation.IsDNS1123Subdomain(ns.Name)) > 0 {
		return NewErrorWithCode(fmt.Sprintf(invalidName, ns.Name), http.StatusBadRequest)
	}
	return nil
}

func validateResourceQuota(cpu, mem api_resource.Quantity) []error {
	errors := []error{}

	mincpu, _ := api_resource.ParseQuantity(minNamespaceCPU)
	maxcpu, _ := api_resource.ParseQuantity(maxNamespaceCPU)
	minmem, _ := api_resource.ParseQuantity(minNamespaceMemory)
	maxmem, _ := api_resource.ParseQuantity(maxNamespaceMemory)

	if cpu.Cmp(mincpu) == -1 || cpu.Cmp(maxcpu) == 1 {
		errors = append(errors, NewError(fmt.Sprintf(invalidCPUQuota, cpu.String(), minNamespaceCPU, maxNamespaceCPU)))
	}

	if mem.Cmp(minmem) == -1 || mem.Cmp(maxmem) == 1 {
		errors = append(errors, NewError(fmt.Sprintf(invalidMemoryQuota, mem.String(), minNamespaceMemory, maxNamespaceMemory)))
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}
