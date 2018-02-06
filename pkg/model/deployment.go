package model

import (
	"git.containerum.net/ch/kube-client/pkg/model"
	v1 "k8s.io/api/apps/v1"
)

func ParseDeploymentList(deploys interface{}) []model.Deployment {
	objects := deploys.(*v1.DeploymentList)
	var deployments []model.Deployment
	for _, deployment := range objects.Items {
		deployment := ParseDeployment(&deployment)
		deployments = append(deployments, deployment)
	}
	return deployments
}

func ParseDeployment(deployment interface{}) model.Deployment {
	obj := deployment.(*v1.Deployment)
	// var containers []Container
	owner := obj.GetLabels()[ownerLabel]
	replicas := 0
	containers := getContainers(obj.Spec.Template.Spec.Containers)
	updated := obj.ObjectMeta.CreationTimestamp.Unix()
	if r := obj.Spec.Replicas; r != nil {
		replicas = int(*r)
	}
	for _, c := range obj.Status.Conditions {
		if t := c.LastUpdateTime.Unix(); t > updated {
			updated = t
		}
	}
	return model.Deployment{
		Name:     obj.GetName(),
		Owner:    &owner,
		Replicas: replicas,
		Status: &model.DeploymentStatus{
			Created:             obj.ObjectMeta.CreationTimestamp.Unix(),
			Updated:             updated,
			Replicas:            int(obj.Status.Replicas),
			ReadyReplicas:       int(obj.Status.ReadyReplicas),
			AvailableReplicas:   int(obj.Status.AvailableReplicas),
			UpdatedReplicas:     int(obj.Status.UpdatedReplicas),
			UnavailableReplicas: int(obj.Status.UnavailableReplicas),
		},
		Containers: containers,
		Hostname:   &obj.Spec.Template.Spec.Hostname,
	}
}
