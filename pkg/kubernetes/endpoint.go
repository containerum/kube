package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kube *Kube) GetEndpointList(namespace string) (*api_core.EndpointsList, error) {
	endpointsAfter, err := kube.CoreV1().Endpoints(namespace).List(api_meta.ListOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": namespace,
		}).Error(ErrUnableGetEndpointList)
		return nil, err
	}
	return endpointsAfter, nil
}

func (kube *Kube) GetEndpoint(namespace, endpoint string) (*api_core.Endpoints, error) {
	endpointAfter, err := kube.CoreV1().Endpoints(namespace).Get(endpoint, api_meta.GetOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": namespace,
			"Endpoint":  endpoint,
		}).Error(ErrUnableGetEndpoint)
		return nil, err
	}
	return endpointAfter, nil
}

func (kube *Kube) CreateEndpoint(endpoint *api_core.Endpoints) (*api_core.Endpoints, error) {
	endpointAfter, err := kube.CoreV1().Endpoints(endpoint.Namespace).Create(endpoint)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": endpoint.Namespace,
			"Endpoint":  endpoint.Name,
		}).Error(ErrUnableCreateEndpoint)
		return nil, err
	}
	return endpointAfter, nil
}

func (kube *Kube) UpdateEndpoint(endpoint *api_core.Endpoints) (*api_core.Endpoints, error) {
	endpointAfter, err := kube.CoreV1().Endpoints(endpoint.Namespace).Update(endpoint)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": endpoint.Namespace,
			"Endpoint":  endpoint.Name,
		}).Error(ErrUnableUpdateEndpoint)
		return nil, err
	}
	return endpointAfter, nil
}

func (kube *Kube) DeleteEndpoint(namespace, endpoint string) error {
	err := kube.CoreV1().Endpoints(namespace).Delete(endpoint, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": namespace,
			"Endpoint":  endpoint,
		}).Error(ErrUnableDeleteEndpoint)
		return err
	}
	return nil
}
