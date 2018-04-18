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
	configMapParam = "configmap"
)

func GetConfigMapList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get config maps list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	cmList, err := kube.GetConfigMapList(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMapList(cmList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"ConfigMap":       configMap,
	}).Debug("Get config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	cm, err := kube.GetConfigMap(namespace, configMap)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMap(cm, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cmReq model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cmReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespace(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		cmReq.Owner = ctx.MustGet(m.UserID).(string)
	}

	cm, errs := cmReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	cmAfter, err := kube.CreateConfigMap(cm)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeConfigMap(cmAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"ConfigMap":       configMap,
	}).Debug("Create config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cmReq model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cmReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespace(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	oldCm, err := kube.GetConfigMap(namespace, ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	cmReq.Name = configMap
	cmReq.Owner = oldCm.GetObjectMeta().GetLabels()[ownerQuery]

	newCm, errs := cmReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	cmAfter, err := kube.UpdateConfigMap(newCm)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMap(cmAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"ConfigMap":       configMap,
	}).Debug("Delete config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteConfigMap(namespace, configMap)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetSelectedConfigMaps(ctx *gin.Context) {
	log.Debug("Get selected config maps Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ret := make(map[string]model.ConfigMapsList, 0)

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {
			cmList, err := kube.GetConfigMapList(n.ID)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			cm, err := model.ParseKubeConfigMapList(cmList, role == m.RoleUser)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			ret[n.Label] = *cm
		}
	}

	ctx.JSON(http.StatusOK, ret)
}
