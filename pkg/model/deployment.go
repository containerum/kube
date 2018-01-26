package model

import v1 "k8s.io/api/apps/v1"

type Deployment struct {
	Name string `json:"name"`
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

	return Deployment{
		Name: obj.GetName(),
	}
}
