package model

import (
	"errors"
	"fmt"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

type SecretWithOwner struct {
	kube_types.Secret
	Owner string `json:"owner,omitempty"`
}

// ParseSecretList parses kubernetes v1.SecretList to more convenient []Secret struct.
func ParseSecretList(secreti interface{}) ([]SecretWithOwner, error) {
	secrets := secreti.(*api_core.SecretList)
	if secrets == nil {
		return nil, ErrUnableConvertSecretList
	}

	newSecrets := make([]SecretWithOwner, 0)
	for _, secret := range secrets.Items {
		newSecret, err := ParseSecret(&secret)
		if err != nil {
			return nil, err
		}

		if newSecret.Owner != "" {
			newSecrets = append(newSecrets, *newSecret)
		}
	}
	return newSecrets, nil
}

// ParseSecret parses kubernetes v1.Secret to more convenient Secret struct.
func ParseSecret(secreti interface{}) (*SecretWithOwner, error) {
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

	return &SecretWithOwner{
		Secret: kube_types.Secret{
			Name:      secret.GetName(),
			CreatedAt: &createdAt,
			Data:      newData,
		},
		Owner: owner,
	}, nil
}

// MakeSecret creates kubernetes v1.Secret from Secret struct and namespace labels
func MakeSecret(nsName string, secret SecretWithOwner, labels map[string]string) (*api_core.Secret, []error) {
	err := ValidateSecret(secret)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = secret.Name
	labels[ownerLabel] = secret.Owner
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

	return &newSecret, nil
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

func ValidateSecret(secret SecretWithOwner) []error {
	errs := []error{}
	if secret.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else {
		if !IsValidUUID(secret.Owner) {
			errs = append(errs, errors.New(invalidOwner))
		}
	}
	if secret.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if len(api_validation.IsDNS1123Subdomain(secret.Name)) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, secret.Name))
	}
	for k := range secret.Data {
		if len(api_validation.IsConfigMapKey(k)) > 0 {
			errs = append(errs, fmt.Errorf(invalidKey, k))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
