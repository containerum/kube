package router

import (
	"fmt"
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"

	api_core "k8s.io/api/core/v1"
)

func getServiceList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get service list call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeServices, err := kube.GetServiceList(namespace)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	}
	services, err := model.ParseServiceList(nativeServices.(*api_core.ServiceList))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, services)
}

func createService(c *gin.Context) {
	log.WithField("Service", c.Param(m.ServiceKey)).Debug("Create service Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	nsname := c.Param(namespaceParam)

	var svc *api_core.Service
	if err := c.ShouldBindWith(&svc, binding.JSON); err != nil {
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

	c.JSON(http.StatusAccepted, svcAfter)
}

func getService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	serviceName := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Get service call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeService, err := kube.GetService(namespace, serviceName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	}
	service, err := model.ServiceFromNativeKubeService(nativeService)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, service)
}

func deleteService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	serviceName := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Delete service call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteService(namespace, serviceName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusOK)
}

func updateService(ctx *gin.Context) {
	serviceName := ctx.Param(serviceParam)
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Update service Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	var service api_core.Service
	if err := ctx.ShouldBindWith(&service, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	if service.ObjectMeta.Namespace != namespace {
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			fmt.Sprintf(namespaceNotMatchError, service.ObjectMeta.Namespace, namespace))
	}
	updatedService, err := kube.UpdateService(&service)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusAccepted, updatedService)
}
