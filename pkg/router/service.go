package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	middleware "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	serviceParam = "service"
)

func getServiceList(ctx *gin.Context) {
	owner := ctx.Query(ownerQuery)
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Owner":     owner,
		"Namespace": namespace,
	}).Debug("Get service list call")
	kube := ctx.MustGet(middleware.KubeClient).(*kubernetes.Kube)
	nativeServices, err := kube.GetServiceList(namespace, owner)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	services, err := model.ParseServiceList(nativeServices)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, services)
}

func getService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	serviceName := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Get service call")
	kube := ctx.MustGet(middleware.KubeClient).(*kubernetes.Kube)
	nativeService, err := kube.GetService(namespace, serviceName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	service, err := model.ServiceFromNativeKubeService(nativeService)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, service)
}
