package model

import (
	"errors"
	"fmt"

	"strings"

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

type NamespacesList struct {
	Namespaces []NamespaceWithOwner `json:"namespaces"`
}

type NamespaceWithOwner struct {
	kube_types.Namespace
	Name  string `json:"name,omitempty"`
	Owner string `json:"owner,omitempty"`
}

// ParseResourceQuotaList parses kubernetes v1.ResourceQuotaList to more convenient []Namespace struct.
// (resource quouta contains all fields that parent namespace contains)
func ParseResourceQuotaList(quotas interface{}) (*NamespacesList, error) {
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
	return &NamespacesList{namespaces}, nil
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
	createdAt := obj.ObjectMeta.CreationTimestamp.Unix()

	return &NamespaceWithOwner{
		Owner: owner,
		Name:  obj.GetNamespace(),
		Namespace: kube_types.Namespace{
			//Label:     obj.GetNamespace(),
			CreatedAt: &createdAt,
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

// MakeNamespace creates kubernetes v1.Namespace from Namespace struct
func MakeNamespace(ns NamespaceWithOwner) (*api_core.Namespace, []error) {
	err := validateNamespace(ns)
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

func validateNamespace(ns NamespaceWithOwner) []error {
	errs := []error{}

	if ns.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(ns.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, ns.Name, strings.Join(err, ","))))
	}
	if ns.Owner != "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else if !IsValidUUID(ns.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// MakeResourceQuota creates kubernetes v1.ResourceQuota from cpu, memory, labels and namespace name
func MakeResourceQuota(ns string, labels map[string]string, resources kube_types.Resource) (*api_core.ResourceQuota, []error) {
	cpuq, err := api_resource.ParseQuantity(resources.CPU)
	if err != nil {
		return nil, []error{ErrInvalidCPUFormat}
	}
	memoryq, err := api_resource.ParseQuantity(resources.Memory)
	if err != nil {
		return nil, []error{ErrInvalidMemoryFormat}
	}

	errs := ValidateResourceQuota(cpuq, memoryq)
	if errs != nil {
		return nil, errs
	}

	newRq := api_core.ResourceQuota{
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
	}
	return &newRq, nil
}

func ValidateResourceQuota(cpu, mem api_resource.Quantity) []error {
	errs := []error{}

	mincpu, _ := api_resource.ParseQuantity(minNamespaceCPU)
	maxcpu, _ := api_resource.ParseQuantity(maxNamespaceCPU)
	minmem, _ := api_resource.ParseQuantity(minNamespaceMemory)
	maxmem, _ := api_resource.ParseQuantity(maxNamespaceMemory)

	if cpu.Cmp(mincpu) == -1 || cpu.Cmp(maxcpu) == 1 {
		errs = append(errs, fmt.Errorf(invalidCPUQuota, cpu.String(), minNamespaceCPU, maxNamespaceCPU))
	}

	if mem.Cmp(minmem) == -1 || mem.Cmp(maxmem) == 1 {
		errs = append(errs, fmt.Errorf(invalidMemoryQuota, mem.String(), minNamespaceMemory, maxNamespaceMemory))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ParseNamespaceListForUser(headers UserHeaderDataMap, nsl []NamespaceWithOwner) *NamespacesList {
	ret := NamespacesList{}
	for _, ns := range nsl {
		nsp := ParseNamespaceForUser(headers, &ns)
		if nsp.Label != "" {
			ret.Namespaces = append(ret.Namespaces, *nsp)
		}
	}
	return &ret
}

func ParseNamespaceForUser(headers UserHeaderDataMap, ns *NamespaceWithOwner) *NamespaceWithOwner {
	for _, n := range headers {
		if ns.Name == n.ID {
			ns.Label = n.Label
		}
	}
	ns.Name = ""
	ns.Owner = ""
	return ns
}
