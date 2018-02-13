package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
)

func ParseSecretList(secreti interface{}) []kube_types.Secret {
	secrets := secreti.(*api_core.SecretList)

	var newSecrets []kube_types.Secret
	for _, secret := range secrets.Items {
		newSecret := ParseSecret(&secret)
		newSecrets = append(newSecrets, *newSecret)
	}
	return newSecrets
}

func ParseSecret(secreti interface{}) *kube_types.Secret {
	secret := secreti.(*api_core.Secret)

	newData := make(map[string]string)
	for k, v := range secret.Data {
		newData[k] = string(v)
	}

	createdAt := secret.CreationTimestamp.Unix()

	newSecret := kube_types.Secret{}
	newSecret.Name = secret.GetName()
	newSecret.CreatedAt = &createdAt
	newSecret.Data = newData

	return &newSecret
}

func MakeSecret(nsName string, secret kube_types.Secret) *api_core.Secret {
	newData := make(map[string][]byte)
	for k, v := range secret.Data {
		newData[k] = []byte(v)
	}

	newSecret := api_core.Secret{}
	newSecret.Kind = "Secret"
	newSecret.APIVersion = "v1"
	newSecret.Data = newData
	newSecret.Type = "Opaque"
	newSecret.SetName(secret.Name)
	newSecret.SetNamespace(nsName)

	return &newSecret
}
