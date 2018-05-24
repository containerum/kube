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

// PersistentVolumeClaimList -- model for pvc list
//
// swagger:model
type PersistentVolumeClaimList struct {
	PersistentVolumeClaims []PersistentVolumeClaimWithOwner `json:"volumes"`
}

// PersistentVolumeClaimWithOwner -- model for pvc with owner
//
// swagger:model
type PersistentVolumeClaimWithOwner struct {
	// swagger: allOf
	kube_types.PersistentVolumeClaim
	// required: true
	Owner string `json:"owner,omitempty"`
}

// ParseKubePersistentVolumeClaimList parses kubernetes v1.PersistentVolumeClaimList to more convenient PersistentVolumeClaimList struct.
func ParseKubePersistentVolumeClaimList(ns interface{}, parseforuser bool) (*PersistentVolumeClaimList, error) {
	nativePvc := ns.(*api_core.PersistentVolumeClaimList)
	if nativePvc == nil {
		return nil, ErrUnableConvertVolumeList
	}

	pvcList := make([]PersistentVolumeClaimWithOwner, 0)
	for _, nativeService := range nativePvc.Items {
		pvc, err := ParseKubePersistentVolumeClaim(&nativeService, parseforuser)
		if err != nil {
			return nil, err
		}
		pvcList = append(pvcList, *pvc)
	}
	return &PersistentVolumeClaimList{pvcList}, nil
}

// ParseKubePersistentVolumeClaim parses kubernetes v1.PersistentVolume to more convenient PersistentVolumeClaimWithOwner struct.
func ParseKubePersistentVolumeClaim(pvci interface{}, parseforuser bool) (*PersistentVolumeClaimWithOwner, error) {
	native := pvci.(*api_core.PersistentVolumeClaim)
	if native == nil {
		return nil, ErrUnableConvertVolume
	}

	owner := native.GetObjectMeta().GetLabels()[ownerLabel]
	size := native.Spec.Resources.Requests["storage"]
	createdAt := native.GetCreationTimestamp().UTC().UTC().Format(time.RFC3339)

	pvc := PersistentVolumeClaimWithOwner{
		PersistentVolumeClaim: kube_types.PersistentVolumeClaim{
			Name:         native.Name,
			CreatedAt:    &createdAt,
			StorageClass: native.ObjectMeta.Annotations[api_core.BetaStorageClassAnnotation],
			AccessMode:   kube_types.PersistentVolumeAccessMode(native.Spec.AccessModes[0]),
			Size:         uint(size.Value() / 1024 / 1024),
		},
		Owner: owner,
	}

	if parseforuser {
		pvc.ParseForUser()
	}

	return &pvc, nil
}

// ToKube creates kubernetes v1.Service from Service struct and namespace labels
func (pvc *PersistentVolumeClaimWithOwner) ToKube(nsName string, labels map[string]string) (*api_core.PersistentVolumeClaim, []error) {
	err := pvc.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError().AddDetails("invalid namespace labels")}
	}

	memsize := api_resource.NewQuantity(int64(pvc.Size)*1024*1024, api_resource.BinarySI)

	newPvc := api_core.PersistentVolumeClaim{
		TypeMeta: api_meta.TypeMeta{
			Kind:       pvcKind,
			APIVersion: pvcAPIVersion,
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:      labels,
			Name:        pvc.Name,
			Annotations: map[string]string{api_core.BetaStorageClassAnnotation: pvc.StorageClass},
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

func (pvc *PersistentVolumeClaimWithOwner) Validate() []error {
	errs := []error{}
	if pvc.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1035Label(pvc.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, pvc.Name, strings.Join(err, ",")))
	}
	if pvc.StorageClass == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "storage_class"))
	}
	if pvc.AccessMode == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "access_mode"))
	}
	if pvc.Size == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "size"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ParseForUser removes information not interesting for users
func (pvc *PersistentVolumeClaimWithOwner) ParseForUser() {
	pvc.Owner = ""
}
