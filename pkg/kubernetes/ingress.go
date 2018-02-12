package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_extensions "k8s.io/api/extensions/v1beta1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) GetIngressList(ns string) (*api_extensions.IngressList, error) {
	ingressList, err := k.ExtensionsV1beta1().Ingresses(ns).List(api_meta.ListOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
		}).Error(ErrUnableGetIngressList)
		return nil, err
	}
	return ingressList, nil
}

func (k *Kube) GetIngress(ns string, ingress string) (*api_extensions.Ingress, error) {
	ingressAfter, err := k.ExtensionsV1beta1().Ingresses(ns).Get(ingress, api_meta.GetOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
		}).Error(ErrUnableGetIngress)
		return nil, err
	}
	return ingressAfter, nil
}

func (k *Kube) CreateIngress(ns string, ingress *api_extensions.Ingress) (*api_extensions.Ingress, error) {
	ingressAfter, err := k.ExtensionsV1beta1().Ingresses(ns).Create(ingress)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
			"Ingress":   ingress.Name,
		}).Error(ErrUnableCreateIngress)
		return nil, err
	}
	return ingressAfter, nil
}

func (k *Kube) UpdateIngress(ns string, ingress *api_extensions.Ingress) (*api_extensions.Ingress, error) {
	ingressAfter, err := k.ExtensionsV1beta1().Ingresses(ns).Update(ingress)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
			"Ingress":   ingress.Name,
		}).Error(ErrUnableUpdateIngress)
		return nil, err
	}
	return ingressAfter, nil
}

func (k *Kube) DeleteIngress(ns string, ingress string) error {
	err := k.ExtensionsV1beta1().Ingresses(ns).Delete(ingress, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
			"Ingress":   ingress,
		}).Error(ErrUnableDeleteIngress)
		return err
	}
	return nil
}
