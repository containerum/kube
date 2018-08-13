package model

import (
	api_storage "k8s.io/api/storage/v1"
)

type StorageList []string

// ParseStoragesList parses kubernetes v1.StorageClassList to more convenient StorageList struct.
func ParseStoragesList(storages interface{}) (*StorageList, error) {
	nativeStorages := storages.(*api_storage.StorageClassList)
	if nativeStorages == nil {
		return nil, ErrUnableConvertStorageList
	}

	storageList := make(StorageList, len(nativeStorages.Items))
	for i := range nativeStorages.Items {
		storageList[i] = nativeStorages.Items[i].Name
	}
	return &storageList, nil
}
