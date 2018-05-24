package model

import (
	"errors"
	"fmt"
	"strings"

	"time"

	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

const (
	ownerLabel = "owner"
	appLabel   = "app"

	minNamespaceCPU    = 300   //m
	minNamespaceMemory = 512   //Mi
	maxNamespaceCPU    = 12000 //m
	maxNamespaceMemory = 28672 //Mi
)

// NamespacesList -- model for namespaces list
//
// swagger:model
type NamespacesList struct {
	Namespaces []NamespaceWithOwner `json:"namespaces"`
}

// NamespaceWithOwner -- model for namespace with owner
//
// swagger:model
type NamespaceWithOwner struct {
	// swagger: allOf
	kube_types.Namespace
	//hosting-internal name
	Owner string `json:"owner,omitempty"`
}

// ParseKubeResourceQuotaList parses kubernetes v1.ResourceQuotaList to more convenient []Namespace struct.
// (resource quouta contains all fields that parent namespace contains)
func ParseKubeResourceQuotaList(quotas interface{}, parseforuser bool) (*NamespacesList, error) {
	objects := quotas.(*api_core.ResourceQuotaList)
	if objects == nil {
		return nil, ErrUnableConvertNamespaceList
	}

	namespaces := make([]NamespaceWithOwner, 0, objects.Size())
	for _, quota := range objects.Items {
		ns, err := ParseKubeResourceQuota(&quota, parseforuser)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, *ns)
	}
	return &NamespacesList{namespaces}, nil
}

// ParseKubeResourceQuota parses kubernetes v1.ResourceQuota to more convenient Namespace struct.
// (resource quouta contains all fields that parent namespace contains)
func ParseKubeResourceQuota(quota interface{}, parseforuser bool) (*NamespaceWithOwner, error) {
	obj := quota.(*api_core.ResourceQuota)
	if obj == nil {
		return nil, ErrUnableConvertNamespace
	}

	cpuLimit := obj.Spec.Hard[api_core.ResourceLimitsCPU]
	memoryLimit := obj.Spec.Hard[api_core.ResourceLimitsMemory]
	cpuUsed := obj.Status.Used[api_core.ResourceLimitsCPU]
	memoryUsed := obj.Status.Used[api_core.ResourceLimitsMemory]
	owner := obj.GetObjectMeta().GetLabels()[ownerLabel]
	createdAt := obj.ObjectMeta.CreationTimestamp.UTC().Format(time.RFC3339)

	ns := NamespaceWithOwner{
		Owner: owner,
		Namespace: kube_types.Namespace{
			ID:        obj.GetNamespace(),
			CreatedAt: &createdAt,
			Resources: kube_types.Resources{
				Hard: kube_types.Resource{
					CPU:    uint(cpuLimit.ScaledValue(api_resource.Milli)),
					Memory: uint(memoryLimit.Value() / 1024 / 1024),
				},
				Used: &kube_types.Resource{
					CPU:    uint(cpuUsed.ScaledValue(api_resource.Milli)),
					Memory: uint(memoryUsed.Value() / 1024 / 1024),
				},
			},
		},
	}

	return &ns, nil
}

// ToKube creates kubernetes v1.Namespace from Namespace struct
func (ns *NamespaceWithOwner) ToKube() (*api_core.Namespace, []error) {
	err := ns.Validate()
	if err != nil {
		return nil, err
	}

	newNs := api_core.Namespace{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels: map[string]string{ownerLabel: ns.Owner},
			Name:   ns.ID,
		},
		Spec: api_core.NamespaceSpec{},
	}
	return &newNs, nil
}

// MakeResourceQuota creates kubernetes v1.ResourceQuota from cpu, memory, labels and namespace name
func MakeResourceQuota(ns string, labels map[string]string, resources kube_types.Resource) (*api_core.ResourceQuota, []error) {
	errs := ValidateResourceQuota(resources.CPU, resources.Memory)
	if errs != nil {
		return nil, errs
	}

	cpuLim := api_resource.NewScaledQuantity(int64(resources.CPU), api_resource.Milli)
	memLim := api_resource.NewQuantity(int64(resources.Memory)*1024*1024, api_resource.BinarySI)
	//Requests is equal to Limits
	cpuReq := cpuLim
	memReq := memLim

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
				api_core.ResourceRequestsCPU:    *cpuReq,
				api_core.ResourceLimitsCPU:      *cpuLim,
				api_core.ResourceRequestsMemory: *memReq,
				api_core.ResourceLimitsMemory:   *memLim,
			},
		},
	}
	return &newRq, nil
}

func ParseNamespaceListForUser(headers UserHeaderDataMap, nsl []NamespaceWithOwner) *NamespacesList {
	nso := make([]NamespaceWithOwner, 0)
	ret := NamespacesList{nso}
	for _, ns := range nsl {
		ns.ParseForUser(headers)
		if ns.Label != "" {
			ret.Namespaces = append(ret.Namespaces, ns)
		}
	}
	return &ret
}

func (ns *NamespaceWithOwner) ParseForUser(headers UserHeaderDataMap) {
	ns.Label = ""
	for _, n := range headers {
		if ns.ID == n.ID {
			ns.Label = n.Label
			ns.Access = string(n.Access)
		}
	}
	ns.Owner = ""
}

func (ns *NamespaceWithOwner) Validate() []error {
	errs := []error{}

	if ns.ID == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "id"))
	} else if err := api_validation.IsDNS1123Label(ns.ID); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, ns.ID, strings.Join(err, ",")))
	}
	if ns.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "owner"))
	} else if !IsValidUUID(ns.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateResourceQuota(cpu, mem uint) []error {
	errs := []error{}

	if cpu < minNamespaceCPU || cpu > maxNamespaceCPU {
		errs = append(errs, fmt.Errorf(invalidCPUQuota, cpu, minNamespaceCPU, maxNamespaceCPU))
	}

	if mem < minNamespaceMemory || mem > maxNamespaceMemory {
		errs = append(errs, fmt.Errorf(invalidMemoryQuota, mem, minNamespaceMemory, maxNamespaceMemory))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
