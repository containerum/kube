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
	api_extensions "k8s.io/api/extensions/v1beta1"
)

const (
	ingressParam = "ingress"
)

func GetIngressList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get ingress list")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingressList, err := kubecli.GetIngressList(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseIngressList(ingressList, role == "user")
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Ingress":         ctx.Param(ingressParam),
	}).Debug("Get ingress Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingress, err := kubecli.GetIngress(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(ingressParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseIngress(ingress, role == "user")
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace_Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Create ingress Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress model.IngressWithOwner
	if err := ctx.ShouldBindWith(&ingress, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		ingress.Owner = ctx.MustGet(m.UserID).(string)
	}

	newIngress, errs := model.MakeIngress(ctx.MustGet(m.NamespaceKey).(string), ingress, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kubecli.CreateIngress(newIngress)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseIngress(ingressAfter, role == "user")
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace_Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Ingress":         ctx.Param(ingressParam),
	}).Debug("Update ingress Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress model.IngressWithOwner
	if err := ctx.ShouldBindWith(&ingress, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		ingress.Owner = ctx.MustGet(m.UserID).(string)
	}

	newIngress, errs := model.MakeIngress(ctx.MustGet(m.NamespaceKey).(string), ingress, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kubecli.UpdateIngress(newIngress)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	ret, err := model.ParseIngress(ingressAfter, role == "user")
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteIngress(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace_Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Delete ingress Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kubecli.DeleteIngress(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(ingressParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

func GetSelectedIngresses(ctx *gin.Context) {
	log.Debug("Get selected ingresses Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingresses := make(map[string]model.IngressesList, 0)

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {

			ingressList, err := kubecli.GetIngressList(n.ID)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			il, err := model.ParseIngressList(ingressList, true)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			ingresses[n.Label] = *il
		}
	}

	ctx.JSON(http.StatusOK, ingresses)
}

func CreateIngressFromFile(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace_Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Create ingress from file Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress api_extensions.Ingress
	if err := ctx.ShouldBindWith(&ingress, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	errs := model.ValidateIngressFromFile(&ingress)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		ingress.Labels["owner"] = ctx.MustGet(m.UserID).(string)
		ingress.Namespace = ctx.MustGet(m.NamespaceKey).(string)
	} else {
		ingress.Namespace = ctx.Param(namespaceParam)
	}

	_, err := kubecli.GetNamespaceQuota(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	ingressAfter, err := kubecli.CreateIngress(&ingress)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailsErr(err), ctx)
		return
	}

	ret, err := model.ParseIngress(ingressAfter, role == "user")
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}
