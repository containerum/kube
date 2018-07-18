package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	quotaName = "quota"
)

//GetNamespaceList returns namespaces list
func (k *Kube) GetNamespaceList(owner string) (*api_core.NamespaceList, error) {
	quotas, err := k.CoreV1().Namespaces().List(api_meta.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithField("Owner", owner).Error(err)
		return nil, err
	}
	return quotas, nil
}

//GetNamespaceQuotaList returns namespaces (quotas) list
func (k *Kube) GetNamespaceQuotaList(owner string) (*api_core.ResourceQuotaList, error) {
	quotas, err := k.CoreV1().ResourceQuotas("").List(api_meta.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithField("Owner", owner).Error(err)
		return nil, err
	}
	return quotas, nil
}

//GetNamespace returns namespace
func (k *Kube) GetNamespace(nsName string) (*api_core.Namespace, error) {
	ns, err := k.CoreV1().Namespaces().Get(nsName, api_meta.GetOptions{})
	if err != nil {
		log.WithField("Namespace", ns).Error(err)
		return nil, err
	}
	return ns, nil
}

//GetNamespaceQuota returns namespace (quota)
func (k *Kube) GetNamespaceQuota(ns string) (*api_core.ResourceQuota, error) {
	quota, err := k.CoreV1().ResourceQuotas(ns).Get(quotaName, api_meta.GetOptions{})
	if err != nil {
		log.WithField("Namespace", ns).Error(err)
		return nil, err
	}
	return quota, nil
}

//CreateNamespace creates namespace
func (k *Kube) CreateNamespace(ns *api_core.Namespace) (*api_core.Namespace, error) {
	nsAfter, err := k.CoreV1().Namespaces().Create(ns)
	if err != nil {
		log.WithField("Namespace", ns.Name).Error(err)
		return nil, err
	}
	return nsAfter, nil
}

//CreateNamespaceQuota creates namespace quota
func (k *Kube) CreateNamespaceQuota(nsName string, quota *api_core.ResourceQuota) (*api_core.ResourceQuota, error) {
	quotaAfter, err := k.CoreV1().ResourceQuotas(nsName).Create(quota)
	if err != nil {
		log.WithField("Namespace", nsName).Error(err)
		return nil, err
	}
	return quotaAfter, nil
}

//UpdateNamespaceQuota updates namespace quota
func (k *Kube) UpdateNamespaceQuota(nsName string, quota *api_core.ResourceQuota) (*api_core.ResourceQuota, error) {
	quotaAfter, err := k.CoreV1().ResourceQuotas(nsName).Update(quota)
	if err != nil {
		log.WithField("Namespace", nsName).Error(err)
		return nil, err
	}
	return quotaAfter, nil
}

//DeleteNamespace deletes namespace
func (k *Kube) DeleteNamespace(nsName string) error {
	err := k.CoreV1().Namespaces().Delete(nsName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithField("Namespace", nsName).Error(err)
		return err
	}
	return nil
}
