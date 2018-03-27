package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin/binding"
)

const (
	ownerQuery     = "owner"
	namespaceParam = "namespace"

	userIDHeader = "X-User-Id"
)

func GetNamespaceList(ctx *gin.Context) {
	owner := ctx.Query(ownerQuery)

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		owner = ctx.MustGet(userIDHeader).(string)
	}

	log.WithField("Owner", owner).Debug("Get namespace list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	quotas, err := kube.GetNamespaceQuotaList(owner)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ret, err := model.ParseResourceQuotaList(quotas)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	if role == "user" {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		ret = model.ParseNamespaceListForUser(*nsList, ret.Namespaces)
	}
	ctx.JSON(http.StatusOK, ret)
}

func GetNamespace(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	quota, err := kube.GetNamespaceQuota(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	ret, err := model.ParseResourceQuota(quota)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		ret = model.ParseNamespaceForUser(*nsList, ret)
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateNamespace(ctx *gin.Context) {
	log.Debug("Create namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ns model.NamespaceWithOwner
	if err := ctx.ShouldBindWith(&ns, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	newNs, errs := model.MakeNamespace(ns)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	newQuota, errs := model.MakeResourceQuota(ns.Name, newNs.Labels, ns.Resources.Hard)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	_, err := kube.CreateNamespace(newNs)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	quotaCreated, err := kube.CreateNamespaceQuota(ns.Name, newQuota)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseResourceQuota(quotaCreated)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateNamespace(ctx *gin.Context) {
	log.WithField("Namespace", ctx.Param(namespaceParam)).Debug("Update namespace Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var res model.NamespaceWithOwner
	if err := ctx.ShouldBindWith(&res, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quotaOld, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	quota, errs := model.MakeResourceQuota(quotaOld.Namespace, quotaOld.Labels, res.Resources.Hard)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	quotaAfter, err := kubecli.UpdateNamespaceQuota(ctx.Param(namespaceParam), quota)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	ret, err := model.ParseResourceQuota(quotaAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)

}

func DeleteNamespace(ctx *gin.Context) {
	log.WithField("Namespace", ctx.Param(namespaceParam)).Debug("Delete namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteNamespace(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
