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
	ingressParam = "ingress"
)

func GetIngressList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get ingress list")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingressList, err := kube.GetIngressList(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeIngressList(ingressList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Ingress":         ingr,
	}).Debug("Get ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingress, err := kube.GetIngress(namespace, ingr)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeIngressList(ingress, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingressReq model.IngressWithOwner
	if err := ctx.ShouldBindWith(&ingressReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		ingressReq.Owner = ctx.MustGet(m.UserID).(string)
	}

	newIngress, errs := ingressReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kube.CreateIngress(newIngress)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeIngress(ingressAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Ingress":         ingr,
	}).Debug("Update ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingressReq model.IngressWithOwner
	if err := ctx.ShouldBindWith(&ingressReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	oldIngress, err := kube.GetIngress(namespace, ingr)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	ingressReq.Name = ingr
	ingressReq.Owner = oldIngress.GetObjectMeta().GetLabels()[ownerQuery]

	newIngress, errs := ingressReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kube.UpdateIngress(newIngress)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeIngress(ingressAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Ingress":         ingr,
	}).Debug("Delete ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteIngress(namespace, ingr)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetSelectedIngresses(ctx *gin.Context) {
	log.Debug("Get selected ingresses Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingresses := make(map[string]model.IngressesList, 0)

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {

			ingressList, err := kube.GetIngressList(n.ID)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			ingressesList, err := model.ParseKubeIngressList(ingressList, role == m.RoleUser)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			ingresses[n.Label] = *ingressesList
		}
	}

	ctx.JSON(http.StatusOK, ingresses)
}
