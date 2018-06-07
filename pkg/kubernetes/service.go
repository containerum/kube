package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetServiceList returns services list
func (k *Kube) GetServiceList(nsname string) (*api_core.ServiceList, error) {
	services, err := k.CoreV1().Services(nsname).List(api_meta.ListOptions{})
	if err != nil {
		log.WithField("Namespace", nsname).Error(err)
		return nil, err
	}
	return services, nil
}

func (k *Kube) GetServiceSolutionList(ns string, solutionID string) (*api_core.ServiceList, error) {
	services, err := k.CoreV1().Services(ns).List(api_meta.ListOptions{
		LabelSelector: "solution=" + solutionID,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Solution":  solutionID,
		}).Error(err)
		return nil, err
	}
	return services, nil
}

//GetService returns service
func (k *Kube) GetService(namespace, serviceName string) (*api_core.Service, error) {
	nativeService, err := k.CoreV1().Services(namespace).Get(serviceName, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"Service":   serviceName,
		}).Error(err)
		return nil, err
	}
	return nativeService, nil
}

//CreateService creates service
func (k *Kube) CreateService(svc *api_core.Service) (*api_core.Service, error) {
	svcAfter, err := k.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": svc.Namespace,
			"Service":   svc.Name,
		}).Error(err)
		return nil, err
	}
	return svcAfter, nil
}

//UpdateService updates service
func (k *Kube) UpdateService(service *api_core.Service) (*api_core.Service, error) {
	newService, err := k.CoreV1().
		Services(service.ObjectMeta.Namespace).
		Update(service)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": service.Namespace,
			"Service":   service.Name,
		}).Error(err)
		return nil, err
	}
	return newService, nil
}

//DeleteService deletes service
func (k *Kube) DeleteService(namespace, serviceName string) error {
	err := k.CoreV1().Services(namespace).
		Delete(serviceName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"Service":   serviceName,
		}).Error(err)
		return err
	}
	return nil
}
