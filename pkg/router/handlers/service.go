package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	api_core "k8s.io/api/core/v1"

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
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseServiceList(nativeServices, role == "user")
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
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
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseService(nativeService, role == "user")
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateService(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create service Call")
	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithOwner
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableCreateResource(), ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	newSvc, errs := model.MakeService(ctx.Param(namespaceParam), svc, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	svcAfter, err := kubecli.CreateService(newSvc)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseService(svcAfter, role == "user")
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
		gonic.Gonic(cherry.ErrUnableUpdateResource(), ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	svc.Name = ctx.Param(serviceParam)

	oldSvc, err := kubecli.GetService(ctx.Param(namespaceParam), ctx.Param(serviceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	newSvc, errs := model.MakeService(ctx.Param(namespaceParam), svc, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
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

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseService(updatedService, role == "user")
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
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

func CreateServiceFromFile(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Create service Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc api_core.Service
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableCreateResource(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		svc.Labels["owner"] = ctx.MustGet(m.UserID).(string)
		svc.Namespace = ctx.MustGet(m.NamespaceKey).(string)
	} else {
		svc.Namespace = ctx.Param(namespaceParam)
	}

	errs := model.ValidateServiceFromFile(&svc)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	_, err := kubecli.GetNamespaceQuota(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	svcAfter, err := kubecli.CreateService(&svc)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailsErr(err), ctx)
		return
	}

	ret, err := model.ParseService(svcAfter, role == "user")
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}
