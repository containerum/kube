package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetSecretList returns secrets list
func (k *Kube) GetSecretList(nsName string) (*api_core.SecretList, error) {
	secrets, err := k.CoreV1().Secrets(nsName).List(api_meta.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": nsName,
		}).Error(err)
		return nil, err
	}
	return secrets, nil
}

//GetSecret returns secret
func (k *Kube) GetSecret(nsName string, secretName string) (*api_core.Secret, error) {
	secret, err := k.CoreV1().Secrets(nsName).Get(secretName, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": nsName,
			"Secret":    secretName,
		}).Error(err)
		return nil, err
	}
	return secret, nil
}

//CreateSecret creates secret
func (k *Kube) CreateSecret(secret *api_core.Secret) (*api_core.Secret, error) {
	newSecret, err := k.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": secret.Namespace,
			"Secret":    secret.Name,
		}).Error(err)
		return nil, err
	}
	return newSecret, nil
}

//UpdateSecret updates secret
func (k *Kube) UpdateSecret(secret *api_core.Secret) (*api_core.Secret, error) {
	newSecret, err := k.CoreV1().Secrets(secret.Namespace).Update(secret)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": secret.Namespace,
			"Secret":    secret.Name,
		}).Error(err)
		return nil, err
	}
	return newSecret, nil
}

//DeleteSecret deletes secret
func (k *Kube) DeleteSecret(nsName string, secretName string) error {
	err := k.CoreV1().Secrets(nsName).Delete(secretName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": nsName,
			"Secret":    secretName,
		}).Error(err)
		return err
	}
	return nil
}
