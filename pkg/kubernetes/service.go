package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) GetServiceList(nsname string) (interface{}, error) {
	svcAfter, err := k.CoreV1().Services(nsname).List(api_meta.ListOptions{})
	if err != nil {
		log.WithError(err).WithField("Namespace", nsname).Error(ErrUnableGetServiceList)
		return nil, ErrUnableGetServiceList
	}
	return svcAfter, nil
}

func (k *Kube) CreateService(svc *api_core.Service) (*api_core.Service, error) {
	svcAfter, err := k.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil {
		log.WithError(err).WithField("Namespace", svc.Name).Error(ErrUnableCreateService)
		return nil, err
	}
	return svcAfter, nil
}
