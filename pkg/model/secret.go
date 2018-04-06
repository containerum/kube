package model

import (
	"errors"
	"fmt"
	"strings"

	"time"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

type SecretsList struct {
	Secrets []SecretWithOwner `json:"secrets"`
}

type SecretWithOwner struct {
	kube_types.Secret
	Owner string `json:"owner,omitempty"`
}

const (
	secretKind       = "Secret"
	secretApiVersion = "v1"
)

// ParseSecretList parses kubernetes v1.SecretList to more convenient []Secret struct.
func ParseSecretList(secreti interface{}, parseforuser bool) (*SecretsList, error) {
	secrets := secreti.(*api_core.SecretList)
	if secrets == nil {
		return nil, ErrUnableConvertSecretList
	}

	newSecrets := make([]SecretWithOwner, 0)
	for _, secret := range secrets.Items {
		newSecret, err := ParseSecret(&secret, false)
		if err != nil {
			return nil, err
		}

		if newSecret.Owner != "" || !parseforuser {
			if parseforuser {
				newSecret.Owner = ""
			}
			newSecrets = append(newSecrets, *newSecret)
		}
	}
	return &SecretsList{newSecrets}, nil
}

// ParseSecret parses kubernetes v1.Secret to more convenient Secret struct.
func ParseSecret(secreti interface{}, parseforuser bool) (*SecretWithOwner, error) {
	secret := secreti.(*api_core.Secret)
	if secret == nil {
		return nil, ErrUnableConvertSecret
	}

	newData := make(map[string]string)
	for k, v := range secret.Data {
		newData[k] = string(v)
	}

	owner := secret.GetObjectMeta().GetLabels()[ownerLabel]
	createdAt := secret.CreationTimestamp.Format(time.RFC3339)

	newSecret := SecretWithOwner{
		Secret: kube_types.Secret{
			Name:      secret.GetName(),
			CreatedAt: &createdAt,
			Data:      newData,
		},
		Owner: owner,
	}

	if parseforuser {
		newSecret.Owner = ""
	}

	return &newSecret, nil

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
	labels[ownerLabel] = secret.Owner

	newSecret := api_core.Secret{
		TypeMeta: api_meta.TypeMeta{
			Kind:       secretKind,
			APIVersion: secretApiVersion,
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
	} else if !IsValidUUID(secret.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}
	if secret.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(secret.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, secret.Name, strings.Join(err, ","))))
	}
	for k := range secret.Data {
		if err := api_validation.IsConfigMapKey(k); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, k, strings.Join(err, ",")))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateSecretFromFile(secret *api_core.Secret) []error {
	errs := []error{}

	if secret.Kind != secretKind {
		errs = append(errs, fmt.Errorf(invalidResourceKind, secret.Kind, secretKind))
	}

	if secret.APIVersion != "" && secret.APIVersion != secretApiVersion {
		errs = append(errs, fmt.Errorf(invalidApiVersion, secret.APIVersion, secretApiVersion))
	}

	if secret.GetLabels()[ownerLabel] == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Label: Owner"))
	} else if !IsValidUUID(secret.GetLabels()[ownerLabel]) {
		errs = append(errs, errors.New(invalidOwner))
	}

	if secret.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Label(secret.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, secret.Name, strings.Join(err, ","))))
	}

	for k := range secret.Data {
		if err := api_validation.IsConfigMapKey(k); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, k, strings.Join(err, ",")))
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
