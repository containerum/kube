package kubernetes

import (
	log "github.com/sirupsen/logrus"
	kubeCoreV1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetService returns Service with given name
// from provided namespace.
// In case of trouble returns ErrUnableGetService
func (kube *Kube) GetService(namespace, serviceName string) (*kubeCoreV1.Service, error) {
	nativeService, err := kube.CoreV1().
		Services(namespace).
		Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": namespace,
			"Service":   serviceName,
		}).Error(ErrUnableGetService)
		return nil, ErrUnableGetService
	}
	return nativeService, nil
}

func (kube *Kube) GetServiceList(nsname string) (interface{}, error) {
	svcAfter, err := kube.CoreV1().Services(nsname).List(meta_v1.ListOptions{})
	if err != nil {
		log.WithError(err).WithField("Namespace", nsname).Error(ErrUnableGetServiceList)
		return nil, ErrUnableGetServiceList
	}
	return svcAfter, nil
}

func (kube *Kube) CreateService(svc *kubeCoreV1.Service) (*kubeCoreV1.Service, error) {
	svcAfter, err := kube.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil {
		log.WithError(err).WithField("Namespace", svc.Name).Error(ErrUnableCreateService)
		return nil, err
	}
	return svcAfter, nil
}
