package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_storage "k8s.io/api/storage/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) GetStorageClassesList() (*api_storage.StorageClassList, error) {
	storages, err := k.StorageV1().StorageClasses().List(api_meta.ListOptions{})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return storages, nil
}
