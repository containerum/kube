package model

import (
	v1 "k8s.io/api/apps/v1"
)

type Deployment struct {
	Name            string            `json:"name"`
	Owner           string            `json:"owner_id,omitempty"`
	Replicas        int               `json:"replicas"`
	Containers      []Container       `json:"containers"`
	ImagePullSecret map[string]string `json:"image_pull_secret,omitempty"`
	Status          DeploymentStatus  `json:"status,omitempty"`
	Hostname        string            `json:"hostname,omitempty"`
}

type DeploymentStatus struct {
	Created             int64 `json:"created_at"`
	Updated             int64 `json:"updated_at"`
	Replicas            int   `json:"replicas"`
	ReadyReplicas       int   `json:"ready_replicas"`
	AvailableReplicas   int   `json:"available_replicas"`
	UnavailableReplicas int   `json:"unavailable_replicas"`
	UpdatedReplicas     int   `json:"updated_replicas"`
}

func ParseDeploymentList(deploys interface{}) []Deployment {
	objects := deploys.(*v1.DeploymentList)
	var deployments []Deployment
	for _, deployment := range objects.Items {
		deployment := ParseDeployment(&deployment)
		deployments = append(deployments, deployment)
	}
	return deployments
}

func ParseDeployment(deployment interface{}) Deployment {
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
	return Deployment{
		Name:     obj.GetName(),
		Owner:    owner,
		Replicas: replicas,
		Status: DeploymentStatus{
			Created:             obj.ObjectMeta.CreationTimestamp.Unix(),
			Updated:             updated,
			Replicas:            int(obj.Status.Replicas),
			ReadyReplicas:       int(obj.Status.ReadyReplicas),
			AvailableReplicas:   int(obj.Status.AvailableReplicas),
			UpdatedReplicas:     int(obj.Status.UpdatedReplicas),
			UnavailableReplicas: int(obj.Status.UnavailableReplicas),
		},
		Containers: containers,
		Hostname:   obj.Spec.Template.Spec.Hostname,
	}
}
