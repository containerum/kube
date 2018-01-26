package kubernetes

import (
	"errors"

	log "github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	quotaName = "quota"
)

var (
	ErrUnableGetNamespaceQuotaList = errors.New("Unable to get namespace list")
	ErrUnableGetNamespaceQuota     = errors.New("Unable to get namespace")
)

func (k *Kube) GetNamespaceQuotaList(owner string) (interface{}, error) {
	quotas, err := k.CoreV1().ResourceQuotas("").List(meta_v1.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithError(err).WithField("Owner", owner).Error(ErrUnableGetNamespaceQuotaList)
		return nil, ErrUnableGetNamespaceQuotaList
	}
	return quotas, nil
}

func (k *Kube) GetNamespaceQuota(ns string) (interface{}, error) {
	quota, err := k.CoreV1().ResourceQuotas(ns).Get(quotaName, meta_v1.GetOptions{})
	if err != nil {
		log.WithError(err).WithField("Namespace", ns).Error(ErrUnableGetNamespaceQuota)
		return nil, ErrUnableGetNamespaceQuota
	}
	return quota, nil
}
