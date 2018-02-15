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
	ingressParam = "ingress"
)

func getIngressList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingressList, err := kubecli.GetIngressList(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseIngressList(ingressList)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func getIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingress, err := kubecli.GetIngress(ctx.Param(namespaceParam), ctx.Param(ingressParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseIngress(ingress)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func createIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress kube_types.Ingress
	if err := ctx.ShouldBindWith(&ingress, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ingressAfter, err := kubecli.CreateIngress(model.MakeIngress(ctx.Param(namespaceParam), ingress, quota.Labels))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseIngress(ingressAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, ret)
}

func updateIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Ingress":   ctx.Param(ingressParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress kube_types.Ingress
	if err := ctx.ShouldBindWith(&ingress, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	if ctx.Param(ingressParam) != ingress.Name {
		log.Errorf(invalidUpdateIngressName, ctx.Param(ingressParam), ingress.Name)
		ctx.Error(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateIngressName, ctx.Param(ingressParam), ingress.Name), http.StatusBadRequest))
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateIngressName, ctx.Param(ingressParam), ingress.Name), http.StatusBadRequest)))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ingressAfter, err := kubecli.UpdateIngress(model.MakeIngress(ctx.Param(namespaceParam), ingress, quota.Labels))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseIngress(ingressAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func deleteIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kubecli.DeleteIngress(ctx.Param(namespaceParam), ctx.Param(ingressParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}
