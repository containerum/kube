package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	json_types "git.containerum.net/ch/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"

	api_core "k8s.io/api/core/v1"
)

func getServiceList(c *gin.Context) {
	namespace := c.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": c.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get service list call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeServices, err := kube.GetServiceList(namespace)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	ret, err := model.ParseServiceList(nativeServices.(*api_core.ServiceList))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, ret)
}

func createService(c *gin.Context) {
	log.WithField("Service", c.Param(m.ServiceKey)).Debug("Create service Call")
	var svc json_types.Service
	if err := c.ShouldBindWith(&svc, binding.JSON); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newSvc, err := model.MakeService(c.Param(namespaceParam), &svc)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	svcAfter, err := kubecli.CreateService(newSvc)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusCreated, model.ParseService(svcAfter))
}

func getService(c *gin.Context) {
	namespace := c.MustGet(m.NamespaceKey).(string)
	serviceName := c.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace Param": c.Param(namespaceParam),
		"Namespace":       namespace,
		"Service":         serviceName,
	}).Debug("Get service call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeService, err := kube.GetService(namespace, serviceName)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, model.ParseService(nativeService))
}

func deleteService(c *gin.Context) {
	namespace := c.Param(namespaceParam)
	serviceName := c.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Delete service call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteService(namespace, serviceName)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.Status(http.StatusAccepted)
}

func updateService(c *gin.Context) {
	serviceName := c.Param(serviceParam)
	namespace := c.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Update service Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	var svc json_types.Service
	if err := c.ShouldBindWith(&svc, binding.JSON); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newSvc, err := model.MakeService(c.Param(namespaceParam), &svc)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	updatedService, err := kube.UpdateService(newSvc)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, model.ParseService(updatedService))
}
