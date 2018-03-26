package kubernetes

import (
	log "github.com/sirupsen/logrus"
	api_apps "k8s.io/api/apps/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetDeploymentList returns deployments list
func (k *Kube) GetDeploymentList(ns string, owner string) (*api_apps.DeploymentList, error) {
	deployments, err := k.AppsV1().Deployments(ns).List(api_meta.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Owner":     owner,
		}).Error(ErrUnableGetDeploymentList)
		return nil, err
	}
	return deployments, nil
}

//GetDeployment returns deployment
func (k *Kube) GetDeployment(ns string, deploy string) (*api_apps.Deployment, error) {
	deployment, err := k.AppsV1().Deployments(ns).Get(deploy, api_meta.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace":  ns,
			"Deployment": deploy,
		}).Error(ErrUnableGetDeployment)
		return nil, err
	}
	return deployment, nil
}

//CreateDeployment creates deployment
func (k *Kube) CreateDeployment(depl *api_apps.Deployment) (*api_apps.Deployment, error) {
	deployment, err := k.AppsV1().Deployments(depl.Namespace).Create(depl)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace":  depl.Namespace,
			"Deployment": depl.Name,
		}).Error(ErrUnableCreateDeployment)
		return nil, err
	}
	return deployment, nil
}

//DeleteDeployment deletes deployment
func (k *Kube) DeleteDeployment(ns string, deployName string) error {
	err := k.AppsV1().Deployments(ns).Delete(deployName, &api_meta.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace":  ns,
			"Deployment": deployName,
		}).Error(ErrUnableDeleteDeployment)
		return err
	}
	return nil
}

//UpdateDeployment updates deployment
func (k *Kube) UpdateDeployment(depl *api_apps.Deployment) (*api_apps.Deployment, error) {
	deployment, err := k.AppsV1().Deployments(depl.Namespace).Update(depl)
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace":  depl.Namespace,
			"Deployment": depl.Name,
		}).Error(ErrUnableUpdateDeployment)
		return nil, err
	}
	return deployment, nil
}
