package router

import (
	"fmt"
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	api_core "k8s.io/api/core/v1"
)

func getServiceList(c *gin.Context) {
	log.WithField("Namespace", c.Query(namespaceParam)).Debug("Get services list Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	svc, err := kubecli.GetServiceList(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, svc)
}

func createService(c *gin.Context) {
	log.WithField("Service", c.Param(m.ServiceKey)).Debug("Create namespace Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	nsname := c.Param(namespaceParam)

	var svc *api_core.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	if nsname != svc.ObjectMeta.Namespace {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf(namespaceNotMatchError, svc.ObjectMeta.Name, nsname))
		return
	}

	svcAfter, err := kubecli.CreateService(svc)
	if err != nil {
		log.Errorf(serviceCreationError, svc.ObjectMeta.Name, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf(serviceCreationError, svc.ObjectMeta.Name, err.Error()))
		return
	}

	c.Status(http.StatusAccepted)
	c.Set(m.ResponseObjectKey, svcAfter)
}
