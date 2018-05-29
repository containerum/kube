package model

import (
	"fmt"
	"strings"

	"time"

	"encoding/base64"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

type ConfigMapKubeAPI kube_types.ConfigMap

// ParseKubeConfigMapList parses kubernetes v1.ConfigMapList to more convenient []ConfigMap struct.
func ParseKubeConfigMapList(cmi interface{}, parseforuser bool) (*kube_types.ConfigMapsList, error) {
	cmList := cmi.(*api_core.ConfigMapList)
	if cmList == nil {
		return nil, ErrUnableConvertConfigMapList
	}

	newCms := make([]kube_types.ConfigMap, 0)
	for _, cm := range cmList.Items {
		newCm, err := ParseKubeConfigMap(&cm, parseforuser)
		if err != nil {
			return nil, err
		}
		newCms = append(newCms, *newCm)
	}
	return &kube_types.ConfigMapsList{newCms}, nil
}

// ParseKubeConfigMap parses kubernetes v1.ConfigMap to more convenient ConfigMap struct.
func ParseKubeConfigMap(cmi interface{}, parseforuser bool) (*kube_types.ConfigMap, error) {
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

	newCm := kube_types.ConfigMap{
		Name:      cm.GetName(),
		CreatedAt: &createdAt,
		Data:      kube_types.ConfigMapData(newData),
		Owner:     owner,
	}

	if parseforuser {
		newCm.Mask()
	}
	return &newCm, nil
}

// ToKube creates kubernetes v1.ConfigMap from ConfigMap struct and namespace labels
func (cm *ConfigMapKubeAPI) ToKube(nsName string, labels map[string]string) (*api_core.ConfigMap, []error) {
	if err := cm.Validate(); err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError().AddDetails("invalid namespace labels")}
	}

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

func (cm *ConfigMapKubeAPI) Validate() []error {
	var errs []error
	if cm.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1123Label(cm.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, cm.Name, strings.Join(err, ",")))
	}
	if len(cm.Data) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "data"))
	} else {
		for k := range cm.Data {
			if err := api_validation.IsConfigMapKey(k); len(err) > 0 {
				errs = append(errs, fmt.Errorf(invalidName, k, strings.Join(err, ",")))
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
