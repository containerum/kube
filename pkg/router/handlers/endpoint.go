package handlers

import (
	"net/http"

	"fmt"

	json_types "git.containerum.net/ch/json-types/kube-api"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	endpointParam = "endpoint"
)

func GetEndpointList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get endpoints list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	endpoints, err := kube.GetEndpointList(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseEndpointList(endpoints)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetEndpoint(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Endpoint":        ctx.Param(endpointParam),
	}).Debug("Get endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	endpoint, err := kube.GetEndpoint(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(endpointParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseEndpoint(endpoint)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateEndpoint(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create endpoint Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var endpoint json_types.Endpoint
	if err := ctx.ShouldBindWith(&endpoint, binding.JSON); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
		}).Warning(kubernetes.ErrUnableCreateEndpoint)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	endpointAfter, err := kubecli.CreateEndpoint(model.MakeEndpoint(ctx.Param(namespaceParam), endpoint, quota.Labels))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateEndpoint(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Endpoint":  ctx.Param(endpointParam),
	}).Debug("Create endpoint Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var endpoint json_types.Endpoint
	if err := ctx.ShouldBindWith(&endpoint, binding.JSON); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
			"Endpoint":  ctx.Param(endpointParam),
		}).Warning(kubernetes.ErrUnableUpdateEndpoint)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	if ctx.Param(endpointParam) != endpoint.Name {
		log.WithError(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateEndpointName, ctx.Param(endpointParam), endpoint.Name), http.StatusBadRequest)).WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
			"Endpoint":  ctx.Param(endpointParam),
		}).Warning(kubernetes.ErrUnableUpdateEndpoint)
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateEndpointName, ctx.Param(endpointParam), endpoint.Name), http.StatusBadRequest)))
		return
	}

	endpointAfter, err := kubecli.UpdateEndpoint(model.MakeEndpoint(ctx.Param(namespaceParam), endpoint, quota.Labels))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, ret)
}

func DeleteEndpoint(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Endpoint":  ctx.Param(endpointParam),
	}).Debug("Delete endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteEndpoint(ctx.Param(namespaceParam), ctx.Param(endpointParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}
