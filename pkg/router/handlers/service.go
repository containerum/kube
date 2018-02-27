package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
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
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}
	ret, err := model.ParseServiceList(nativeServices)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
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
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableGetResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseService(nativeService)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResource().Gonic(ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateService(ctx *gin.Context) {
	log.WithField("Service", ctx.Param(m.ServiceKey)).Debug("Create service Call")
	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithOwner
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrUnableCreateResource().Gonic(ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	newSvc, errs := model.MakeService(ctx.Param(namespaceParam), svc, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	svcAfter, err := kubecli.CreateService(newSvc)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseService(svcAfter)
	if err != nil {
		ctx.Error(err)
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
		ctx.Error(err)
		cherry.ErrUnableUpdateResource().Gonic(ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	svc.Name = ctx.Param(serviceParam)

	oldSvc, err := kubecli.GetService(ctx.Param(namespaceParam), ctx.Param(serviceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	newSvc, errs := model.MakeService(ctx.Param(namespaceParam), svc, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	newSvc.ResourceVersion = oldSvc.ResourceVersion
	newSvc.Spec.ClusterIP = oldSvc.Spec.ClusterIP

	updatedService, err := kubecli.UpdateService(newSvc)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ret, err := model.ParseService(updatedService)
	if err != nil {
		ctx.Error(err)
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
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableDeleteResource()).Gonic(ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
