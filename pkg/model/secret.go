package model

import (
	json_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
)

func MakeSecret(nsName string, secret json_types.Secret) *api_core.Secret {
	newData := make(map[string][]byte)
	for k, v := range secret.Data {
		newData[k] = []byte(v)
	}

	newSecret := api_core.Secret{}
	newSecret.Kind = "Secret"
	newSecret.APIVersion = "v1"
	newSecret.Data = newData
	newSecret.Type = "Opaque"
	newSecret.ObjectMeta.Name = secret.Name
	newSecret.ObjectMeta.Namespace = nsName

	return &newSecret
}
