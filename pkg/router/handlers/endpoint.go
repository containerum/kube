package handlers

import (
	"net/http"

	json_types "git.containerum.net/ch/json-types/kube-api"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
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
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ret, err := model.ParseEndpointList(endpoints)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
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
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableGetResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseEndpoint(endpoint)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResource().Gonic(ctx)
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
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	newEndpoint, errs := model.MakeEndpoint(ctx.Param(namespaceParam), endpoint, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	endpointAfter, err := kubecli.CreateEndpoint(newEndpoint)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
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
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	endpoint.Name = ctx.Param(endpointParam)

	newEndpoint, errs := model.MakeEndpoint(ctx.Param(namespaceParam), endpoint, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	endpointAfter, err := kubecli.UpdateEndpoint(newEndpoint)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
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
		model.ParseResourceError(err, cherry.ErrUnableDeleteResource()).Gonic(ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
