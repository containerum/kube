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
	configMapParam = "configmap"
)

func GetConfigMapList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get config maps list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	cm, err := kube.GetConfigMapList(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ret, err := model.ParseConfigMapList(cm)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetConfigMap(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"ConfigMap":       ctx.Param(configMapParam),
	}).Debug("Get config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	cm, err := kube.GetConfigMap(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(configMapParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableGetResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseConfigMap(cm)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResource().Gonic(ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateConfigMap(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cm model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cm, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	newCm, errs := model.MakeConfigMap(ctx.Param(namespaceParam), cm, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	cmAfter, err := kube.CreateConfigMap(newCm)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseConfigMap(cmAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateConfigMap(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"ConfigMap": ctx.Param(configMapParam),
	}).Debug("Create config map Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cm model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cm, binding.JSON); err != nil {
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

	cm.Name = ctx.Param(configMapParam)

	newCm, errs := model.MakeConfigMap(ctx.Param(namespaceParam), cm, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	cmAfter, err := kubecli.UpdateConfigMap(newCm)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseConfigMap(cmAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteConfigMap(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"ConfigMap": ctx.Param(configMapParam),
	}).Debug("Delete config map Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteConfigMap(ctx.Param(namespaceParam), ctx.Param(configMapParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableDeleteResource()).Gonic(ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
