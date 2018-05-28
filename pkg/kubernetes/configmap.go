package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetConfigMapList returns config maps list
func (k *Kube) GetConfigMapList(namespace string) (*api_core.ConfigMapList, error) {
	cmAfter, err := k.CoreV1().ConfigMaps(namespace).List(api_meta.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
		}).Error(err)
		return nil, err
	}
	return cmAfter, nil
}

//GetConfigMap returns config map
func (k *Kube) GetConfigMap(namespace, cm string) (*api_core.ConfigMap, error) {
	cmAfter, err := k.CoreV1().ConfigMaps(namespace).Get(cm, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"ConfigMap": cm,
		}).Error(err)
		return nil, err
	}
	return cmAfter, nil
}

//CreateConfigMap creates config map
func (k *Kube) CreateConfigMap(cm *api_core.ConfigMap) (*api_core.ConfigMap, error) {
	cmAfter, err := k.CoreV1().ConfigMaps(cm.Namespace).Create(cm)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": cm.Namespace,
			"ConfigMap": cm.Name,
		}).Error(err)
		return nil, err
	}
	return cmAfter, nil
}

//UpdateConfigMap updates config map
func (k *Kube) UpdateConfigMap(cm *api_core.ConfigMap) (*api_core.ConfigMap, error) {
	cmAfter, err := k.CoreV1().ConfigMaps(cm.Namespace).Update(cm)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": cm.Namespace,
			"ConfigMap": cm.Name,
		}).Error(err)
		return nil, err
	}
	return cmAfter, nil
}

//DeleteConfigMap deletes config map
func (k *Kube) DeleteConfigMap(namespace, cm string) error {
	err := k.CoreV1().ConfigMaps(namespace).Delete(cm, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": namespace,
			"ConfigMap": cm,
		}).Error(err)
		return err
	}
	return nil
}
