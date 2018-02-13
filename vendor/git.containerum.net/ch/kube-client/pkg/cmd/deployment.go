package cmd

import (
	"git.containerum.net/ch/kube-client/pkg/model"
)

const (
	kubeAPIdeploymentPath  = "/namespaces/{namespace}/deployments/{deployment}"
	kubeAPIdeploymentsPath = "/namespaces/{namespace}/deployments"

	resourceDeploymentPath = "/namespace/{namespace}/deployment/{deployment}"
	resourceImagePath      = resourceDeploymentPath + "/image"
	resourceReplicasPath   = resourceDeploymentPath + "/replicas"
)

// GetDeployment -- consumes a namespace and a deployment names,
// returns a Deployment data OR uninitialized struct AND an error
func (client *Client) GetDeployment(namespace, deployment string) (model.Deployment, error) {
	resp, err := client.Request.
		SetPathParams(map[string]string{
			"namespace":  namespace,
			"deployment": deployment,
		}).SetResult(model.Deployment{}).
		Get(client.serverURL + kubeAPIdeploymentPath)
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
		Get(client.serverURL + kubeAPIdeploymentsPath)
	if err != nil {
		return nil, err
	}
	return *resp.Result().(*[]model.Deployment), nil
}

// DeleteDeployment -- consumes a namespace, a deployment,
// an user role and an ID
func (client *Client) DeleteDeployment(namespace, deployment string) error {
	_, err := client.Request.
		SetPathParams(map[string]string{
			"namespace":  namespace,
			"deployment": deployment,
		}).
		Delete(client.resourceServiceAddr + resourceDeploymentPath)
	return err
}

// CreateDeployment -- consumes a namespace, an user ID and a Role,
// returns nil if OK
func (client *Client) CreateDeployment(namespace string, deployment model.Deployment) error {
	_, err := client.Request.
		SetPathParams(map[string]string{
			"namespace": namespace,
		}).SetBody(deployment).
		Post(client.resourceServiceAddr + "/namespace/{namespace}/deployment")
	return err
}

func (client *Client) SetContainerImage(namespace, deployment string, updateImage model.UpdateImage) error {
	_, err := client.Request.
		SetPathParams(map[string]string{
			"namespace":  namespace,
			"deployment": deployment,
		}).SetBody(updateImage).
		Put(client.resourceServiceAddr + resourceImagePath)
	return err
}

func (client *Client) ReplaceDeployment(namespace string, deployment model.Deployment) error {
	_, err := client.Request.
		SetPathParams(map[string]string{
			"namespace":  namespace,
			"deployment": deployment.Name,
		}).SetBody(deployment).
		Put(client.resourceServiceAddr + resourceDeploymentPath)
	return err
}

func (client *Client) SetReplicas(namespace, deployment string, replicas int) error {
	_, err := client.Request.SetPathParams(map[string]string{
		"namespace":  namespace,
		"deployment": deployment,
	}).SetBody(model.UpdateReplicas{replicas}).
		Put(client.resourceServiceAddr + resourceReplicasPath)
	return err
}
