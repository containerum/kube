package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) GetSecretList(nsName string) (*api_core.SecretList, error) {
	secrets, err := k.CoreV1().Secrets(nsName).List(api_meta.ListOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": nsName,
		}).Error(ErrUnableGetSecretList)
		return nil, err
	}
	return secrets, nil
}

func (k *Kube) GetSecret(nsName string, secretName string) (*api_core.Secret, error) {
	secret, err := k.CoreV1().Secrets(nsName).Get(secretName, api_meta.GetOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": nsName,
			"Secret":    secretName,
		}).Error(ErrUnableGetSecret)
		return nil, err
	}
	return secret, nil
}

func (k *Kube) CreateSecret(secret *api_core.Secret) (*api_core.Secret, error) {
	newSecret, err := k.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": secret.Namespace,
			"Secret":    secret.Name,
		}).Error(ErrUnableCreateSecret)
		return nil, err
	}
	return newSecret, nil
}

func (k *Kube) UpdateSecret(secret *api_core.Secret) (*api_core.Secret, error) {
	newSecret, err := k.CoreV1().Secrets(secret.Namespace).Update(secret)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": secret.Namespace,
			"Secret":    secret.Name,
		}).Error(ErrUnableUpdateSecret)
		return nil, err
	}
	return newSecret, nil
}

func (k *Kube) DeleteSecret(nsName string, secretName string) error {
	err := k.CoreV1().Secrets(nsName).Delete(secretName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": nsName,
			"Secret":    secretName,
		}).Error(ErrUnableDeleteSecret)
		return err
	}
	return nil
}
