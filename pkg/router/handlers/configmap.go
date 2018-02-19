package handlers

import (
	"net/http"

	"fmt"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
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
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseConfigMapList(cm)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
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
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseConfigMap(cm)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateConfigMap(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create config map Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cm model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cm, binding.JSON); err != nil {
		log.WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
		}).Warning(kubernetes.ErrUnableCreateConfigMap)
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

	newCm, err := model.MakeConfigMap(ctx.Param(namespaceParam), cm, quota.Labels)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	cmAfter, err := kubecli.CreateConfigMap(newCm)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseConfigMap(cmAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
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
		log.WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
			"ConfigMap": ctx.Param(configMapParam),
		}).Warning(kubernetes.ErrUnableUpdateConfigMap)
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

	if ctx.Param(configMapParam) != cm.Name {
		log.WithFields(log.Fields{
			"Namespace": ctx.Param(namespaceParam),
			"ConfigMap": ctx.Param(configMapParam),
		}).Warning(kubernetes.ErrUnableUpdateConfigMap)
		ctx.Error(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateConfigMapName, ctx.Param(configMapParam), cm.Name), http.StatusBadRequest))
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateConfigMapName, ctx.Param(configMapParam), cm.Name), http.StatusBadRequest)))
		return
	}

	newCm, err := model.MakeConfigMap(ctx.Param(namespaceParam), cm, quota.Labels)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	cmAfter, err := kubecli.UpdateConfigMap(newCm)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseConfigMap(cmAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
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
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}
