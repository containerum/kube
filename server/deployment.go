package server

import (
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"k8s.io/api/apps/v1beta2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// middleware deps:
// 	SetNamespace
// 	Set(…)KubeClient
// 	ParseJSON
func CreateDeployment(c *gin.Context) {
	depl, ok := c.MustGet("requestObject").(*v1beta2.Deployment)
	if !ok || depl == nil {
		c.AbortWithStatusJSON(400, map[string]string{"error": "bad request"})
		return
	}

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	deplAfter, err := kubecli.AppsV1beta2().Deployments(depl.ObjectMeta.Namespace).Create(depl)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot create deployment: %v", err),
		})
		return
	}

	redactDeploymentForUser(deplAfter)

	c.JSON(201, deplAfter)
}

// middleware deps:
// 	SetNamespace
// 	Set(…)KubeClient
func ListDeployments(c *gin.Context) {
	ns := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	deplList, err := kubecli.AppsV1beta2().Deployments(ns).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.List error: %v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot list deployments: %v", err),
		})
	}

	for i := range deplList.Items {
		redactDeploymentForUser(&deplList.Items[i])
	}

	c.JSON(200, deplList)
}

// middleware deps:
// 	SetNamespace
// 	SetObjectName
// 	Set(…)KubeClient
func DeleteDeployment(c *gin.Context) {
	ns := c.MustGet("namespace").(string)
	deplname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	err := kubecli.AppsV1beta2().Deployments(ns).Delete(deplname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot delete deployments: %v", err),
		})
	}
}

func redactDeploymentForUser(depl *v1beta2.Deployment) {
	depl.Spec.Template.Spec.NodeSelector = nil
}
