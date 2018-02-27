package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kube *Kube) GetConfigMapList(namespace string) (*api_core.ConfigMapList, error) {
	cmAfter, err := kube.CoreV1().ConfigMaps(namespace).List(api_meta.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
		}).Error(ErrUnableGetConfigMapList)
		return nil, err
	}
	return cmAfter, nil
}

func (kube *Kube) GetConfigMap(namespace, cm string) (*api_core.ConfigMap, error) {
	cmAfter, err := kube.CoreV1().ConfigMaps(namespace).Get(cm, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"ConfigMap": cm,
		}).Error(ErrUnableGetConfigMap)
		return nil, err
	}
	return cmAfter, nil
}

func (kube *Kube) CreateConfigMap(cm *api_core.ConfigMap) (*api_core.ConfigMap, error) {
	cmAfter, err := kube.CoreV1().ConfigMaps(cm.Namespace).Create(cm)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": cm.Namespace,
			"ConfigMap": cm.Name,
		}).Error(ErrUnableCreateConfigMap)
		return nil, err
	}
	return cmAfter, nil
}

func (kube *Kube) UpdateConfigMap(cm *api_core.ConfigMap) (*api_core.ConfigMap, error) {
	cmAfter, err := kube.CoreV1().ConfigMaps(cm.Namespace).Update(cm)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": cm.Namespace,
			"ConfigMap": cm.Name,
		}).Error(ErrUnableUpdateConfigMap)
		return nil, err
	}
	return cmAfter, nil
}

func (kube *Kube) DeleteConfigMap(namespace, cm string) error {
	err := kube.CoreV1().ConfigMaps(namespace).Delete(cm, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"ConfigMap": cm,
		}).Error(ErrUnableDeleteConfigMap)
		return err
	}
	return nil
}
