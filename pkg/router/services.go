package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
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
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ret, err := model.ParseServiceList(nativeServices)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, ret)
}

func createService(ctx *gin.Context) {
	log.WithField("Service", ctx.Param(m.ServiceKey)).Debug("Create service Call")
	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc kube_types.Service
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	newSvc, err := model.MakeService(ctx.Param(namespaceParam), &svc)
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	newSvc.Labels = quota.Labels

	svcAfter, err := kubecli.CreateService(newSvc)
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, model.ParseService(svcAfter))
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
		c.AbortWithStatusJSON(model.ParseErorrs(err))
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
		c.AbortWithStatusJSON(model.ParseErorrs(err))
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
	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	var svc kube_types.Service
	if err := c.ShouldBindWith(&svc, binding.JSON); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	newSvc, err := model.MakeService(c.Param(namespaceParam), &svc)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	for k, v := range quota.Labels {
		newSvc.Labels[k] = v
	}

	updatedService, err := kubecli.UpdateService(newSvc)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, model.ParseService(updatedService))
}
