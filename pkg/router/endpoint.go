package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	endpointParam = "endpoints"
)

func getEndpointList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get endpoints list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	endpoints, err := kube.GetEndpointList(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, endpoints)
}

func getEndpoint(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, endpoint)
}

func deleteEndpoint(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Endpoint":  ctx.Param(endpointParam),
	}).Debug("Delete endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteEndpoint(ctx.Param(namespaceParam), ctx.Param(endpointParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}
