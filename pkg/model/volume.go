package model

import (
	"fmt"

	"time"

	"strings"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

const (
	pvcKind       = "PersistentVolumeClaim"
	pvcAPIVersion = "v1"
)

type VolumeKubeAPI kube_types.Volume

// ParseKubePersistentVolumeClaimList parses kubernetes v1.PersistentVolumeClaimList to more convenient PersistentVolumeClaimList struct.
func ParseKubePersistentVolumeClaimList(ns interface{}, parseforuser bool) (*kube_types.VolumesList, error) {
	nativePvc := ns.(*api_core.PersistentVolumeClaimList)
	if nativePvc == nil {
		return nil, ErrUnableConvertVolumeList
	}

	pvcList := make([]kube_types.Volume, 0)
	for _, nativeService := range nativePvc.Items {
		pvc, err := ParseKubePersistentVolumeClaim(&nativeService, parseforuser)
		if err != nil {
			return nil, err
		}
		pvcList = append(pvcList, *pvc)
	}
	return &kube_types.VolumesList{pvcList}, nil
}

// ParseKubePersistentVolumeClaim parses kubernetes v1.PersistentVolume to more convenient PersistentVolumeClaimWithOwner struct.
func ParseKubePersistentVolumeClaim(pvci interface{}, parseforuser bool) (*kube_types.Volume, error) {
	native := pvci.(*api_core.PersistentVolumeClaim)
	if native == nil {
		return nil, ErrUnableConvertVolume
	}

	owner := native.GetObjectMeta().GetLabels()[ownerLabel]
	capacity := native.Spec.Resources.Requests["storage"]
	createdAt := native.GetCreationTimestamp().UTC().UTC().Format(time.RFC3339)

	pvc := kube_types.Volume{
		ID:          native.Name,
		CreatedAt:   &createdAt,
		StorageName: native.ObjectMeta.Annotations[api_core.BetaStorageClassAnnotation],
		AccessMode:  kube_types.PersistentVolumeAccessMode(native.Spec.AccessModes[0]),
		Capacity:    uint(capacity.Value() / 1024 / 1024 / 1024), //in Gi
		Owner:       owner,
	}

	if parseforuser {
		pvc.Mask()
	}

	return &pvc, nil
}

// ToKube creates kubernetes v1.Service from Service struct and namespace labels
func (pvc *VolumeKubeAPI) ToKube(nsName string, labels map[string]string) (*api_core.PersistentVolumeClaim, []error) {
	err := pvc.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError().AddDetails("invalid namespace labels")}
	}

	memsize := api_resource.NewQuantity(int64(pvc.Capacity)*1024*1024*1024, api_resource.BinarySI)

	newPvc := api_core.PersistentVolumeClaim{
		TypeMeta: api_meta.TypeMeta{
			Kind:       pvcKind,
			APIVersion: pvcAPIVersion,
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:      labels,
			Name:        pvc.ID,
			Annotations: map[string]string{api_core.BetaStorageClassAnnotation: pvc.StorageName},
			Namespace:   nsName,
		},
		Spec: api_core.PersistentVolumeClaimSpec{
			AccessModes: []api_core.PersistentVolumeAccessMode{api_core.PersistentVolumeAccessMode(pvc.AccessMode)},
			Resources: api_core.ResourceRequirements{
				Requests: api_core.ResourceList{
					"storage": *memsize,
				},
			},
		},
	}

	return &newPvc, nil
}

func (pvc *VolumeKubeAPI) Validate() []error {
	var errs []error
	if pvc.ID == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "id"))
	} else if err := api_validation.IsDNS1123Label(pvc.ID); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, pvc.ID, strings.Join(err, ",")))
	}
	if pvc.StorageName == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "storage_name"))
	}
	if pvc.AccessMode == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "access_mode"))
	}
	if pvc.Capacity == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "capacity"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
