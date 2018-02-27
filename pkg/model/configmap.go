package model

import (
	"errors"
	"fmt"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

type ConfigMapWithOwner struct {
	kube_types.ConfigMap
	Owner string `json:"owner,omitempty"`
}

// ParseConfigMapList parses kubernetes v1.ConfigMapList to more convenient []ConfigMap struct.
func ParseConfigMapList(cmi interface{}) ([]ConfigMapWithOwner, error) {
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
	return newCms, nil
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
	err := validateConfigMap(cm)
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

func validateConfigMap(cm ConfigMapWithOwner) []error {
	errs := []error{}
	if cm.Owner == "" {
		errs = append(errs, errors.New(noOwner))
	}
	if len(api_validation.IsDNS1123Subdomain(cm.Name)) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, cm.Name)))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
