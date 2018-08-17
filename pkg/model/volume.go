package model

import (
	"fmt"

	"time"

	"strings"

	"git.containerum.net/ch/kube-api/pkg/kubeerrors"
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
	return &kube_types.VolumesList{Volumes: pvcList}, nil
}

// ParseKubePersistentVolumeClaim parses kubernetes v1.PersistentVolume to more convenient PersistentVolumeClaimWithOwner struct.
func ParseKubePersistentVolumeClaim(pvci interface{}, parseforuser bool) (*kube_types.Volume, error) {
	native := pvci.(*api_core.PersistentVolumeClaim)
	if native == nil {
		return nil, ErrUnableConvertVolume
	}

	capacity := native.Spec.Resources.Requests["storage"]

	pvc := kube_types.Volume{
		Name:        native.Name,
		Namespace:   native.Namespace,
		Status:      string(native.Status.Phase),
		CreatedAt:   native.GetCreationTimestamp().UTC().Format(time.RFC3339),
		StorageName: native.ObjectMeta.Annotations[api_core.BetaStorageClassAnnotation],
		AccessMode:  kube_types.PersistentVolumeAccessMode(native.Spec.AccessModes[0]),
		Capacity:    uint(capacity.Value() / 1024 / 1024 / 1024), //in Gi
		Owner:       native.GetObjectMeta().GetLabels()[ownerLabel],
	}

	if parseforuser {
		pvc.Mask()
	}

	return &pvc, nil
}

// ToKube creates kubernetes v1.Service from Service struct and namespace labels
func (pvc *VolumeKubeAPI) ToKube(nsName string, labels map[string]string) (*api_core.PersistentVolumeClaim, []error) {
	//TODO Maybe we should use different access modes
	pvc.AccessMode = kube_types.ReadWriteMany
	err := pvc.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeerrors.ErrInternalError().AddDetails("invalid project labels")}
	}

	memsize := api_resource.NewQuantity(int64(pvc.Capacity)*1024*1024*1024, api_resource.BinarySI)

	newPvc := api_core.PersistentVolumeClaim{
		TypeMeta: api_meta.TypeMeta{
			Kind:       pvcKind,
			APIVersion: pvcAPIVersion,
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:      labels,
			Name:        pvc.Name,
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

// ToKube creates kubernetes v1.Service from Service struct and namespace labels
func (pvc *VolumeKubeAPI) Resize(oldpvc *api_core.PersistentVolumeClaim) (*api_core.PersistentVolumeClaim, error) {
	if oldpvc.Status.Phase != api_core.ClaimBound {
		return nil, kubeerrors.ErrVolumeNotReady()
	}

	memsize := api_resource.NewQuantity(int64(pvc.Capacity)*1024*1024*1024, api_resource.BinarySI)
	if memsize.Cmp(oldpvc.Spec.Resources.Requests["storage"]) < 1 {
		return nil, kubeerrors.ErrUnableDownsizeVolume()
	}

	oldpvc.Spec.Resources.Requests["storage"] = *memsize

	return oldpvc, nil
}

func (pvc *VolumeKubeAPI) Validate() []error {
	var errs []error
	if pvc.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "id"))
	} else if err := api_validation.IsDNS1123Label(pvc.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, pvc.Name, strings.Join(err, ",")))
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
