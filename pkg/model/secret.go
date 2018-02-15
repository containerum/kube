package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ParseSecretList parses kubernetes v1.SecretList to more convenient []Secret struct.
func ParseSecretList(secreti interface{}) ([]kube_types.Secret, error) {
	secrets := secreti.(*api_core.SecretList)
	if secrets == nil {
		return nil, ErrUnableConvertSecretList
	}

	newSecrets := make([]kube_types.Secret, 0)
	for _, secret := range secrets.Items {
		newSecret, err := ParseSecret(&secret)
		if err != nil {
			return nil, err
		}
		newSecrets = append(newSecrets, *newSecret)
	}
	return newSecrets, nil
}

// ParseSecret parses kubernetes v1.Secret to more convenient Secret struct.
func ParseSecret(secreti interface{}) (*kube_types.Secret, error) {
	secret := secreti.(*api_core.Secret)
	if secret == nil {
		return nil, ErrUnableConvertSecret
	}

	newData := make(map[string]string)
	for k, v := range secret.Data {
		newData[k] = string(v)
	}

	owner := secret.GetObjectMeta().GetLabels()[ownerLabel]
	createdAt := secret.CreationTimestamp.Unix()

	return &kube_types.Secret{
		Name:      secret.GetName(),
		CreatedAt: &createdAt,
		Data:      newData,
		Owner:     &owner,
	}, nil
}

// MakeSecret creates kubernetes v1.Secret from Secret struct and namespace labels
func MakeSecret(nsName string, secret kube_types.Secret, labels map[string]string) *api_core.Secret {
	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = secret.Name
	labels[ownerLabel] = *secret.Owner
	labels[nameLabel] = secret.Name

	newSecret := api_core.Secret{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      secret.Name,
			Namespace: nsName,
		},
		Data: makeSecretData(secret.Data),
		Type: "Opaque",
	}
	return &newSecret
}

func makeSecretData(data map[string]string) map[string][]byte {
	newData := make(map[string][]byte, 0)
	if data != nil {
		for k, v := range data {
			newData[k] = []byte(v)
		}
	}
	return newData
}
