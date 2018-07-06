package model

import (
	"fmt"
	"strings"

	"time"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

// SecretWithParamList -- model for secrets list
//
// swagger:model
type SecretWithParamList struct {
	Secrets []SecretWithParam `json:"secrets"`
}

// SecretWithParam -- model for secret with owner
//
// swagger:model
type SecretWithParam struct {
	// swagger: allOf
	*kube_types.Secret
	//hide secret from users
	Hidden bool `json:"hidden,omitempty"`
}

const (
	secretKind       = "Secret"
	secretAPIVersion = "v1"
)

// ParseKubeSecretList parses kubernetes v1.SecretList to more convenient []Secret struct.
func ParseKubeSecretList(secreti interface{}, parseforuser bool) (*SecretWithParamList, error) {
	nativeSecrets := secreti.(*api_core.SecretList)
	if nativeSecrets == nil {
		return nil, ErrUnableConvertSecretList
	}

	secrets := make([]SecretWithParam, 0)
	for _, secret := range nativeSecrets.Items {
		newSecret, err := ParseKubeSecret(&secret, parseforuser)
		if err != nil {
			return nil, err
		}
		if !newSecret.Hidden || !parseforuser {
			secrets = append(secrets, *newSecret)
		}
	}
	return &SecretWithParamList{secrets}, nil
}

// ParseKubeSecret parses kubernetes v1.Secret to more convenient Secret struct.
func ParseKubeSecret(secreti interface{}, parseforuser bool) (*SecretWithParam, error) {
	secret := secreti.(*api_core.Secret)
	if secret == nil {
		return nil, ErrUnableConvertSecret
	}

	newData := make(map[string]string)
	for k, v := range secret.Data {
		newData[k] = string(v)
	}

	newSecret := SecretWithParam{
		Secret: &kube_types.Secret{
			Name:      secret.GetName(),
			CreatedAt: secret.CreationTimestamp.UTC().Format(time.RFC3339),
			Data:      newData,
			Owner:     secret.GetObjectMeta().GetLabels()[ownerLabel],
		},
	}

	newSecret.ParseForUser()

	return &newSecret, nil

}

// ToKube creates kubernetes v1.Secret from Secret struct and namespace labels
func (secret *SecretWithParam) ToKube(nsName string, labels map[string]string) (*api_core.Secret, []error) {
	err := secret.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError().AddDetails("invalid project labels")}
	}

	newSecret := api_core.Secret{
		TypeMeta: api_meta.TypeMeta{
			Kind:       secretKind,
			APIVersion: secretAPIVersion,
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

func (secret *SecretWithParam) Validate() []error {
	var errs []error
	if secret.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1123Label(secret.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, secret.Name, strings.Join(err, ",")))
	}
	if len(secret.Data) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "data"))
	} else {
		for k := range secret.Data {
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

// ParseForUser removes information not interesting for users
func (secret *SecretWithParam) ParseForUser() {
	if secret.Owner == "" {
		secret.Hidden = true
		return
	}
	secret.Mask()
}
