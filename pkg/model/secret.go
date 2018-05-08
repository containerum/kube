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

// SecretsList -- model for secrets list
//
// swagger:model
type SecretsList struct {
	Secrets []SecretWithOwner `json:"secrets"`
}

// SecretWithOwner -- model for secret with owner
//
// swagger:model
type SecretWithOwner struct {
	// swagger: allOf
	kube_types.Secret
	// required: true
	Owner  string `json:"owner,omitempty"`
	Hidden bool   `json:"hidden,omitempty"`
}

const (
	secretKind       = "Secret"
	secretAPIVersion = "v1"
)

// ParseKubeSecretList parses kubernetes v1.SecretList to more convenient []Secret struct.
func ParseKubeSecretList(secreti interface{}, parseforuser bool) (*SecretsList, error) {
	nativeSecrets := secreti.(*api_core.SecretList)
	if nativeSecrets == nil {
		return nil, ErrUnableConvertSecretList
	}

	secrets := make([]SecretWithOwner, 0)
	for _, secret := range nativeSecrets.Items {
		newSecret, err := ParseKubeSecret(&secret, parseforuser)
		if err != nil {
			return nil, err
		}
		if !newSecret.Hidden || !parseforuser {
			secrets = append(secrets, *newSecret)
		}
	}
	return &SecretsList{secrets}, nil
}

// ParseKubeSecret parses kubernetes v1.Secret to more convenient Secret struct.
func ParseKubeSecret(secreti interface{}, parseforuser bool) (*SecretWithOwner, error) {
	secret := secreti.(*api_core.Secret)
	if secret == nil {
		return nil, ErrUnableConvertSecret
	}

	newData := make(map[string]string)
	for k, v := range secret.Data {
		newData[k] = string(v)
	}

	owner := secret.GetObjectMeta().GetLabels()[ownerLabel]
	createdAt := secret.CreationTimestamp.UTC().Format(time.RFC3339)

	newSecret := SecretWithOwner{
		Secret: kube_types.Secret{
			Name:      secret.GetName(),
			CreatedAt: &createdAt,
			Data:      newData,
		},
		Owner: owner,
	}

	newSecret.ParseForUser()

	return &newSecret, nil

}

// ToKube creates kubernetes v1.Secret from Secret struct and namespace labels
func (secret *SecretWithOwner) ToKube(nsName string, labels map[string]string) (*api_core.Secret, []error) {
	err := secret.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError()}
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

func (secret *SecretWithOwner) Validate() []error {
	errs := []error{}
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
func (secret *SecretWithOwner) ParseForUser() {
	if secret.Owner == "" {
		secret.Hidden = true
		return
	}
	secret.Owner = ""
}
