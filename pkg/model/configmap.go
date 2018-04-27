package model

import (
	"errors"
	"fmt"
	"strings"

	"time"

	"encoding/base64"

	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

// SelectedConfigMapsList -- model for config maps list from all namespaces
//
// swagger:model
type SelectedConfigMapsList map[string]ConfigMapsList

// ConfigMapsList -- model for config maps list
//
// swagger:model
type ConfigMapsList struct {
	ConfigMaps []ConfigMapWithOwner `json:"configmaps"`
}

// ConfigMapWithOwner -- model for config map with owner
//
// swagger:model
type ConfigMapWithOwner struct {
	// swagger: allOf
	kube_types.ConfigMap
	// required: true
	Owner string `json:"owner,omitempty"`
}

// ParseKubeConfigMapList parses kubernetes v1.ConfigMapList to more convenient []ConfigMap struct.
func ParseKubeConfigMapList(cmi interface{}, parseforuser bool) (*ConfigMapsList, error) {
	cmList := cmi.(*api_core.ConfigMapList)
	if cmList == nil {
		return nil, ErrUnableConvertConfigMapList
	}

	newCms := make([]ConfigMapWithOwner, 0)
	for _, cm := range cmList.Items {
		newCm, err := ParseKubeConfigMap(&cm, parseforuser)
		if err != nil {
			return nil, err
		}
		newCms = append(newCms, *newCm)
	}
	return &ConfigMapsList{newCms}, nil
}

// ParseKubeConfigMap parses kubernetes v1.ConfigMap to more convenient ConfigMap struct.
func ParseKubeConfigMap(cmi interface{}, parseforuser bool) (*ConfigMapWithOwner, error) {
	cm := cmi.(*api_core.ConfigMap)
	if cm == nil {
		return nil, ErrUnableConvertConfigMap
	}

	newData := make(map[string]string)
	for k, v := range cm.Data {
		newData[k] = string(v)
	}

	owner := cm.GetObjectMeta().GetLabels()[ownerLabel]
	createdAt := cm.CreationTimestamp.UTC().Format(time.RFC3339)

	newCm := ConfigMapWithOwner{
		ConfigMap: kube_types.ConfigMap{
			Name:      cm.GetName(),
			CreatedAt: &createdAt,
			Data:      newData,
		},
		Owner: owner,
	}

	if parseforuser {
		newCm.Owner = ""
	}

	return &newCm, nil
}

// ToKube creates kubernetes v1.ConfigMap from ConfigMap struct and namespace labels
func (cm *ConfigMapWithOwner) ToKube(nsName string, labels map[string]string) (*api_core.ConfigMap, []error) {
	err := cm.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[ownerLabel] = cm.Owner

	for k, v := range cm.Data {
		dec, err := base64.StdEncoding.DecodeString(v)
		if err == nil {
			cm.Data[k] = string(dec)
		}
	}

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

func (cm *ConfigMapWithOwner) Validate() []error {
	errs := []error{}
	if cm.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else if !IsValidUUID(cm.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}
	if cm.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(cm.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, cm.Name, strings.Join(err, ",")))
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
