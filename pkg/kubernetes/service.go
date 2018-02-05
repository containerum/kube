package kubernetes

import (
	"errors"

	log "github.com/sirupsen/logrus"
	kubeCoreV1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrUnableGetService     = errors.New("Unable to get service")
	ErrUnableGetServiceList = errors.New("Unable to get service list")
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

// GetServiceList returns kubernete ServiceList or ErrUnableGetServiceList
func (kube *Kube) GetServiceList(namespace, owner string) (*kubeCoreV1.ServiceList, error) {
	services, err := kube.CoreV1().
		Services(namespace).
		List(meta_v1.ListOptions{
			LabelSelector: getOwnerLabel(owner),
		})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": namespace,
			"Owner":     owner,
		}).Error(ErrUnableGetServiceList)
		return nil, ErrUnableGetServiceList
	}
	return services, nil
}
