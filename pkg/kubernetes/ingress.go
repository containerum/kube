package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_extensions "k8s.io/api/extensions/v1beta1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetIngressList returns ingresses list
func (k *Kube) GetIngressList(ns string) (*api_extensions.IngressList, error) {
	ingressList, err := k.ExtensionsV1beta1().Ingresses(ns).List(api_meta.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
		}).Error(err)
		return nil, err
	}
	return ingressList, nil
}

//GetIngress returns ingress
func (k *Kube) GetIngress(ns string, ingress string) (*api_extensions.Ingress, error) {
	ingressAfter, err := k.ExtensionsV1beta1().Ingresses(ns).Get(ingress, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Ingress":   ingress,
		}).Error(err)
		return nil, err
	}
	return ingressAfter, nil
}

//CreateIngress creates ingress
func (k *Kube) CreateIngress(ingress *api_extensions.Ingress) (*api_extensions.Ingress, error) {
	ingressAfter, err := k.ExtensionsV1beta1().Ingresses(ingress.Namespace).Create(ingress)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ingress.Namespace,
			"Ingress":   ingress.Name,
		}).Error(err)
		return nil, err
	}
	return ingressAfter, nil
}

//UpdateIngress updates ingress
func (k *Kube) UpdateIngress(ingress *api_extensions.Ingress) (*api_extensions.Ingress, error) {
	ingressAfter, err := k.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(ingress)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ingress.Namespace,
			"Ingress":   ingress.Name,
		}).Error(err)
		return nil, err
	}
	return ingressAfter, nil
}

//DeleteIngress deletes ingress
func (k *Kube) DeleteIngress(ns string, ingress string) error {
	err := k.ExtensionsV1beta1().Ingresses(ns).Delete(ingress, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Ingress":   ingress,
		}).Error(err)
		return err
	}
	return nil
}
