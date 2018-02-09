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

func getServiceList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get service list call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeServices, err := kube.GetServiceList(namespace)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	ret, err := model.ParseServiceList(nativeServices.(*api_core.ServiceList))
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

func createService(ctx *gin.Context) {
	log.WithField("Service", ctx.Param(m.ServiceKey)).Debug("Create service Call")

	var svc json_types.Service
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newSvc, err := model.MakeService(ctx.Param(namespaceParam), &svc)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	svcAfter, err := kubecli.CreateService(newSvc)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, model.ParseService(svcAfter))
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
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	ctx.JSON(http.StatusOK, model.ParseService(nativeService))
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
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}

func updateService(ctx *gin.Context) {
	serviceName := ctx.Param(serviceParam)
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Update service Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	var svc json_types.Service
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newSvc, err := model.MakeService(ctx.Param(namespaceParam), &svc)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	updatedService, err := kube.UpdateService(newSvc)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusAccepted, model.ParseService(updatedService))
}
