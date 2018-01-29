package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	deploymentParam = "deployment"
)

func getDeploymentList(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Owner":     c.Query(ownerQuery),
	}).Debug("Get deployment list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployments, err := kube.GetDeploymentList(c.Param(namespaceParam), c.Query(ownerQuery))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	deployList := model.ParseDeploymentList(deployments)
	c.JSON(http.StatusOK, deployList)
}

func getDeployment(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  c.Param(namespaceParam),
		"Deployment": c.Param(deploymentParam),
	}).Debug("Get deployment Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(c.Param(namespaceParam), c.Param(deploymentParam))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	deploy := model.ParseDeployment(deployment)
	c.JSON(http.StatusOK, deploy)
}
