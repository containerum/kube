package model

import (
	"errors"
	"fmt"
	"strings"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

type ConfigMapsList struct {
	ConfigMaps []ConfigMapWithOwner `json:"configmaps"`
}

type ConfigMapWithOwner struct {
	kube_types.ConfigMap
	Owner string `json:"owner,omitempty"`
}

// ParseConfigMapList parses kubernetes v1.ConfigMapList to more convenient []ConfigMap struct.
func ParseConfigMapList(cmi interface{}) (*ConfigMapsList, error) {
	cm := cmi.(*api_core.ConfigMapList)
	if cm == nil {
		return nil, ErrUnableConvertConfigMapList
	}

	newCms := make([]ConfigMapWithOwner, 0)
	for _, cm := range cm.Items {
		newCm, err := ParseConfigMap(&cm)
		if err != nil {
			return nil, err
		}
		newCms = append(newCms, *newCm)
	}
	return &ConfigMapsList{newCms}, nil
}

// ParseConfigMap parses kubernetes v1.ConfigMap to more convenient ConfigMap struct.
func ParseConfigMap(cmi interface{}) (*ConfigMapWithOwner, error) {
	cm := cmi.(*api_core.ConfigMap)
	if cm == nil {
		return nil, ErrUnableConvertConfigMap
	}

	newData := make(map[string]string)
	for k, v := range cm.Data {
		newData[k] = string(v)
	}

	owner := cm.GetObjectMeta().GetLabels()[ownerLabel]
	createdAt := cm.CreationTimestamp.Unix()

	return &ConfigMapWithOwner{
		ConfigMap: kube_types.ConfigMap{
			Name:      cm.GetName(),
			CreatedAt: &createdAt,
			Data:      newData,
		},
		Owner: owner,
	}, nil
}

// MakeConfigMap creates kubernetes v1.ConfigMap from ConfigMap struct and namespace labels
func MakeConfigMap(nsName string, cm ConfigMapWithOwner, labels map[string]string) (*api_core.ConfigMap, []error) {
	err := ValidateConfigMap(cm)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = cm.Name
	labels[ownerLabel] = cm.Owner
	labels[nameLabel] = cm.Name

	newCm := api_core.ConfigMap{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      cm.Name,
			Namespace: nsName,
		},
		Data: cm.Data,
	}
	return &newCm, nil
}

func ValidateConfigMap(cm ConfigMapWithOwner) []error {
	errs := []error{}
	if cm.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else if !IsValidUUID(cm.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}

	if cm.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(cm.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, cm.Name, strings.Join(err, ","))))
	}
	for k := range cm.Data {
		if err := api_validation.IsConfigMapKey(k); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, k, strings.Join(err, ",")))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
