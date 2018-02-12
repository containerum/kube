package cmd

import (
	"git.containerum.net/ch/kube-client/pkg/model"
)

const (
	deploymentPath  = "/namespaces/{namespace}/deployments/{deployment}"
	deploymentsPath = "/namespaces/{namespace}/deployments"
)

// GetDeployment -- consumes a namespace and a deployment names,
// returns a Deployment data OR uninitialized struct AND an error
func (client *Client) GetDeployment(namespace, deployment string) (model.Deployment, error) {
	resp, err := client.Request.
		SetPathParams(map[string]string{
			"namespace":  namespace,
			"deployment": deployment,
		}).SetResult(model.Deployment{}).
		Get(client.serverURL + deploymentPath)
	if err != nil {
		return model.Deployment{}, err
	}
	return *resp.Result().(*model.Deployment), nil
}

// GetDeploymentList -- consumes a namespace and a deployment names,
// returns a list of Deployments OR nil slice AND an error
func (client *Client) GetDeploymentList(namespace string) ([]model.Deployment, error) {
	resp, err := client.Request.
		SetPathParams(map[string]string{
			"namespace": namespace,
		}).SetResult([]model.Deployment{}).
		Get(client.serverURL + deploymentsPath)
	if err != nil {
		return nil, err
	}
	return *resp.Result().(*[]model.Deployment), nil
}
