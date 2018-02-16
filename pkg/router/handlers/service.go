package handlers

import (
	"net/http"

	"fmt"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	serviceParam = "service"
)

func GetServiceList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get service list call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeServices, err := kube.GetServiceList(namespace)
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ret, err := model.ParseServiceList(nativeServices)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

func GetService(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	serviceName := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Service":         serviceName,
	}).Debug("Get service call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	nativeService, err := kube.GetService(namespace, serviceName)
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseService(nativeService)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateService(ctx *gin.Context) {
	log.WithField("Service", ctx.Param(m.ServiceKey)).Debug("Create service Call")
	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithOwner
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
		}).Warning(kubernetes.ErrUnableCreateService)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	svcAfter, err := kubecli.CreateService(model.MakeService(ctx.Param(namespaceParam), svc, quota.Labels))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseService(svcAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateService(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Service":   ctx.Param(serviceParam),
	}).Debug("Update service Call")
	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	var svc model.ServiceWithOwner
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
			"Service":   ctx.Param(serviceParam),
		}).Warning(kubernetes.ErrUnableUpdateService)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	if ctx.Param(serviceParam) != svc.Name {
		log.WithError(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateServiceName, ctx.Param(serviceParam), svc.Name), http.StatusBadRequest)).WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
			"Service":   ctx.Param(serviceParam),
		}).Warning(kubernetes.ErrUnableUpdateService)
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateServiceName, ctx.Param(serviceParam), svc.Name), http.StatusBadRequest)))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	updatedService, err := kubecli.UpdateService(model.MakeService(ctx.Param(namespaceParam), svc, quota.Labels))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ret, err := model.ParseService(updatedService)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	serviceName := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   serviceName,
	}).Debug("Delete service call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteService(namespace, serviceName)
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}
