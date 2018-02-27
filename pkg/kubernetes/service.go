package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kube *Kube) GetServiceList(nsname string) (interface{}, error) {
	svcAfter, err := kube.CoreV1().Services(nsname).List(api_meta.ListOptions{})
	if err != nil {
		log.WithField("Namespace", nsname).Error(ErrUnableGetServiceList)
		return nil, err
	}
	return svcAfter, nil
}

// GetService returns Service with given name
// from provided namespace.
// In case of trouble returns ErrUnableGetService
func (kube *Kube) GetService(namespace, serviceName string) (*api_core.Service, error) {
	nativeService, err := kube.CoreV1().Services(namespace).Get(serviceName, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"Service":   serviceName,
		}).Error(ErrUnableGetService)
		return nil, err
	}
	return nativeService, nil
}

func (kube *Kube) CreateService(svc *api_core.Service) (*api_core.Service, error) {
	svcAfter, err := kube.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": svc.Namespace,
			"Service":   svc.Name,
		}).Error(ErrUnableCreateService)
		return nil, err
	}
	return svcAfter, nil
}

func (kube *Kube) DeleteService(namespace, serviceName string) error {
	err := kube.CoreV1().Services(namespace).
		Delete(serviceName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"Service":   serviceName,
		}).Error(ErrUnableDeleteService)
		return err
	}
	return nil
}

func (kube *Kube) UpdateService(service *api_core.Service) (*api_core.Service, error) {
	newService, err := kube.CoreV1().
		Services(service.ObjectMeta.Namespace).
		Update(service)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": service.Namespace,
			"Service":   service.Name,
		}).Error(ErrUnableUpdateService)
		return nil, err
	}
	return newService, nil
}
