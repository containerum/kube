package kubernetes

import (
	api_core "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

//GetPersistentVolumeClaimsList returns pvc list
func (k *Kube) GetPersistentVolumeClaimsList(ns string) (interface{}, error) {
	pods, err := k.CoreV1().PersistentVolumeClaims(ns).List(meta_v1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
		}).Error(err)
		return nil, err
	}
	return pods, nil
}

//GetPersistentVolumeClaim returns pvc
func (k *Kube) GetPersistentVolumeClaim(ns string, pvcName string) (*api_core.PersistentVolumeClaim, error) {
	pvc, err := k.CoreV1().PersistentVolumeClaims(ns).Get(pvcName, meta_v1.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"PVC":       pvcName,
		}).Error(err)
		return nil, err
	}
	return pvc, nil
}

//CreatePersistentVolumeClaim creates pvc
func (k *Kube) CreatePersistentVolumeClaim(pvc *api_core.PersistentVolumeClaim) (*api_core.PersistentVolumeClaim, error) {
	newpvc, err := k.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(pvc)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": pvc.Namespace,
			"PVC":       pvc.Name,
		}).Error(err)
		return nil, err
	}
	return newpvc, nil
}

//UpdatePersistentVolumeClaim updates pvc
func (k *Kube) UpdatePersistentVolumeClaim(pvc *api_core.PersistentVolumeClaim) (*api_core.PersistentVolumeClaim, error) {
	updpvc, err := k.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(pvc)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": pvc.Namespace,
			"PVC":       pvc.Name,
		}).Error(err)
		return nil, err
	}
	return updpvc, nil
}

//DeletePersistentVolumeClaim deletes pvc
func (k *Kube) DeletePersistentVolumeClaim(ns string, pvc string) error {
	err := k.CoreV1().Pods(ns).Delete(pvc, &meta_v1.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"PVC":       pvc,
		}).Error(err)
		return err
	}
	return nil
}
