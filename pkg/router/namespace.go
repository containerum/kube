package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin/binding"
)

const (
	ownerQuery     = "owner"
	namespaceParam = "namespace"
)

func getNamespaceList(ctx *gin.Context) {
	log.WithField("Owner", ctx.Query(ownerQuery)).Debug("Get namespace list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	quotas, err := kube.GetNamespaceQuotaList(ctx.Query(ownerQuery))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseResourceQuotaList(quotas)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func getNamespace(ctx *gin.Context) {
	log.WithField("Namespace", ctx.Param(namespaceParam)).Debug("Get namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	quota, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseResourceQuota(quota)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func сreateNamespace(ctx *gin.Context) {
	log.Debug("Create namespace Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ns kube_types.Namespace
	if err := ctx.ShouldBindWith(&ns, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	nsAfter, err := kubecli.CreateNamespace(model.MakeNamespace(ns))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := model.MakeResourceQuota(ns.Resources.Hard.CPU, ns.Resources.Hard.Memory, nsAfter.Labels, nsAfter.Name)
	if err != nil {
		ctx.Error(err)

		if err := kubecli.DeleteNamespace(nsAfter.Name); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		}

		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quotaAfter, err := kubecli.CreateNamespaceQuota(ns.Name, quota)
	if err != nil {
		ctx.Error(err)

		if err := kubecli.DeleteNamespace(nsAfter.Name); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		}

		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quotaAfter.Labels = nsAfter.Labels

	if ns.GlusterIP != nil {
		_, err = kubecli.CreateService(model.MakeGlustercFSService(ns.Name))
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(model.ParseErorrs(err))
			return
		}

		_, err = kubecli.CreateEndpoint(model.MakeEndpoint(ns.Name, *ns.GlusterIP))
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(model.ParseErorrs(err))
			return
		}
	}

	ret, err := model.ParseResourceQuota(quotaAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, ret)
}

func deleteNamespace(ctx *gin.Context) {
	log.WithField("Namespace", ctx.Param(namespaceParam)).Debug("Delete namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteNamespace(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func updateNamespace(ctx *gin.Context) {
	log.WithField("Namespace", ctx.Param(namespaceParam)).Debug("Update namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var res kube_types.UpdateNamespace
	if err := ctx.ShouldBindWith(&res, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quotaOld, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := model.MakeResourceQuota(res.Resources.Hard.CPU, res.Resources.Hard.Memory, quotaOld.Labels, quotaOld.Name)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota.Labels = quotaOld.Labels
	quota.SetNamespace(ctx.Param(namespaceParam))
	quota.SetName("quota")
	quotaAfter, err := kube.UpdateNamespaceQuota(ctx.Param(namespaceParam), quota)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseResourceQuota(quotaAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}
