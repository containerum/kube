package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
	core_v1 "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

func CreateDeployment(c *gin.Context) {
	depl, ok := c.MustGet("requestObject").(*apps_v1beta2.Deployment)
	if !ok || depl == nil {
		c.AbortWithStatusJSON(400, map[string]string{"error": "bad request"})
		return
	}

	kubecli = GetKubeClient()
	deplAfter, err := kubecli.AppsV1beta2().Deployments(depl.ObjectMeta.Namespace).Create(depl)
	if err != nil {
		logWithContext(c).Warnf("kubecli.Deployments.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot create deployment: %v", err),
		})
		return
	}

	filterDeploymentForUser(deplAfter)

	c.JSON(201, deplAfter)
}

func filterDeploymentForUser(depl *apps_v1beta2.Deployment) {
}
