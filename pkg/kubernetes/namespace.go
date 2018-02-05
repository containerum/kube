package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	quotaName = "quota"
)

func (k *Kube) GetNamespaceQuotaList(owner string) (interface{}, error) {
	quotas, err := k.CoreV1().ResourceQuotas("").List(api_meta.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithError(err).WithField("Owner", owner).Error(ErrUnableGetNamespaceQuotaList)
		return nil, ErrUnableGetNamespaceQuotaList
	}
	return quotas, nil
}

func (k *Kube) GetNamespaceQuota(ns string) (interface{}, error) {
	quota, err := k.CoreV1().ResourceQuotas(ns).Get(quotaName, api_meta.GetOptions{})
	if err != nil {
		log.WithError(err).WithField("Namespace", ns).Error(ErrUnableGetNamespaceQuota)
		return nil, ErrUnableGetNamespaceQuota
	}
	return quota, nil
}

func (k *Kube) CreateNamespace(ns *api_core.Namespace) (*api_core.Namespace, error) {
	nsAfter, err := k.CoreV1().Namespaces().Create(ns)
	if err != nil {
		log.WithError(err).WithField("Namespace", ns.Name).Error(ErrUnableCreateNamespace)
		return nil, err
	}
	return nsAfter, nil
}

func (k *Kube) CreateNamespaceQuota(nsName string, quota *api_core.ResourceQuota) error {
	quota.SetNamespace(nsName)
	quota.SetName("quota")

	_, err := k.CoreV1().ResourceQuotas(nsName).Create(quota)
	if err != nil {
		log.WithError(err).WithField("Namespace", nsName).Error(ErrUnableCreateNamespaceQuota)
		return err
	}
	return nil
}

func (k *Kube) UpdateNamespaceQuota(nsName string, quota *api_core.ResourceQuota) (*api_core.ResourceQuota, error) {
	quota.SetNamespace(nsName)
	quota.SetName("quota")

	quotaAfter, err := k.CoreV1().ResourceQuotas(nsName).Update(quota)
	if err != nil {
		log.WithError(err).WithField("Namespace", nsName).Error(ErrUnableUpdateNamespaceQuota)
		return nil, err
	}
	return quotaAfter, nil
}

func (k *Kube) DeleteNamespace(nsName string) error {
	err := k.CoreV1().Namespaces().Delete(nsName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithError(err).WithField("Namespace", nsName).Error(ErrUnableDeleteNamespace)
		return err
	}
	return nil
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
