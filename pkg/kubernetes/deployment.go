package kubernetes

import (
	"errors"

	log "github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrUnableGetDeploymentList = errors.New("Unable to get deployment list")
	//ErrUnableGetDeployment     = errors.New("Unable to get deployment")
)

func (k *Kube) GetDeploymentList(ns string, owner string) (interface{}, error) {
	deployments, err := k.AppsV1().Deployments(ns).List(meta_v1.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
			"Owner":     owner,
		}).Error(ErrUnableGetDeploymentList)
	}
	return deployments, nil
}
