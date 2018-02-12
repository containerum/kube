package router

import (
	"net/http"

	"fmt"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	serviceParam = "service"
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

func getService(ctx *gin.Context) {
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
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}

func updateService(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Service":   ctx.Param(serviceParam),
	}).Debug("Update service Call")
	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	var svc kube_types.Service
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	if ctx.Param(serviceParam) != svc.Name {
		log.Errorf(invalidUpdateSecretName, ctx.Param(serviceParam), svc.Name)
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateServiceName, ctx.Param(serviceParam), svc.Name), http.StatusBadRequest)))
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

	for k, v := range quota.Labels {
		newSvc.Labels[k] = v
	}

	updatedService, err := kubecli.UpdateService(newSvc)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusAccepted, model.ParseService(updatedService))
}
