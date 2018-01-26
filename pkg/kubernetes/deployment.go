package kubernetes

import (
	"errors"

	log "github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nsName = "hosting"
)

var (
	ErrUnableGetDeploymentList = errors.New("Unable to get deployment list")
	//ErrUnableGetDeployment     = errors.New("Unable to get deployment")
)

func (k *Kube) GetDeploymentList(ns string) (interface{}, error) {

	deployments, err := k.AppsV1().Deployments(ns).List(meta_v1.ListOptions{})
	if err != nil {
		log.WithError(err).WithField("Namespace", ns).Error(ErrUnableGetDeploymentList)
	}
	return deployments, nil
}
