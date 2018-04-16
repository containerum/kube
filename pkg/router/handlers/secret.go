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
	api_core "k8s.io/api/core/v1"
)

const (
	secretParam = "secret"
)

func GetSecretList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get secret list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	secrets, err := kube.GetSecretList(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecretList(secrets, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetSecret(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	sct := ctx.Param(secretParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Secret":          sct,
	}).Debug("Get secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	secret, err := kube.GetSecret(namespace, sct)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecret(secret, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateSecret(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secretReq model.SecretWithOwner
	if err := ctx.ShouldBindWith(&secretReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		secretReq.Owner = ctx.MustGet(m.UserID).(string)
	}

	newSecret, errs := secretReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	secretAfter, err := kube.CreateSecret(newSecret)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeSecret(secretAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateSecret(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	sct := ctx.Param(secretParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Secret":          sct,
	}).Debug("Create secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secretReq model.SecretWithOwner
	if err := ctx.ShouldBindWith(&secretReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	oldSecret, err := kube.GetIngress(namespace, sct)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	secretReq.Name = sct
	secretReq.Owner = oldSecret.GetObjectMeta().GetLabels()[ownerQuery]

	newSecret, errs := secretReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	secretAfter, err := kube.UpdateSecret(newSecret)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecret(secretAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteSecret(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	sct := ctx.Param(secretParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Secret":          sct,
	}).Debug("Delete secret Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteSecret(namespace, sct)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

func CreateSecretFromFile(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secret api_core.Secret
	if err := ctx.ShouldBindWith(&secret, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		if secret.Labels == nil {
			secret.Labels = map[string]string{}
		}
		secret.Labels["owner"] = ctx.MustGet(m.UserID).(string)

		secret.Namespace = namespace
	} else {
		secret.Namespace = ctx.Param(namespaceParam)
	}

	errs := model.ValidateSecretFromFile(&secret)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	_, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	secretAfter, err := kube.CreateSecret(&secret)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailsErr(err), ctx)
		return
	}

	ret, err := model.ParseKubeSecret(secretAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}
