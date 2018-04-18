package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	endpointParam = "endpoint"
)

func GetEndpointList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get endpoints list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	endpoints, err := kube.GetEndpointList(namespace)
	if err != nil {
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ret, err := model.ParseKubeEndpointList(endpoints)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetEndpoint(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ep := ctx.Param(endpointParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Endpoint":        ep,
	}).Debug("Get endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	endpoint, err := kube.GetEndpoint(namespace, ep)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	ret, err := model.ParseKubeEndpoint(endpoint)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateEndpoint(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var endpointReq model.Endpoint
	if err := ctx.ShouldBindWith(&endpointReq, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		owner := ctx.MustGet(m.UserID).(string)
		endpointReq.Owner = &owner
	}

	newEndpoint, errs := endpointReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	endpointAfter, err := kube.CreateEndpoint(newEndpoint)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateEndpoint(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ep := ctx.Param(endpointParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Endpoint":        ep,
	}).Debug("Create endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var endpointReq model.Endpoint
	if err := ctx.ShouldBindWith(&endpointReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	oldEndpoint, err := kube.GetEndpoint(namespace, ep)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	endpointReq.Name = ep
	owner := oldEndpoint.GetObjectMeta().GetLabels()[ownerQuery]
	endpointReq.Owner = &owner

	newEndpoint, errs := endpointReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	endpointAfter, err := kube.UpdateEndpoint(newEndpoint)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteEndpoint(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ep := ctx.Param(endpointParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Endpoint":        ep,
	}).Debug("Delete endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteEndpoint(namespace, ep)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
