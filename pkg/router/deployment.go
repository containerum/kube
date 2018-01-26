package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func getDeploymentList(c *gin.Context) {
	log.WithField("Namespace", c.Query(namespaceParam)).Debug("Get deployment list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployments, err := kube.GetDeploymentList(c.Query("namespace"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	deployList := model.ParseDeploymentList(deployments)
	c.JSON(http.StatusOK, deployList)
}
