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
	secretParam = "secret"
)

func getSecretList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get secret list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	secrets, err := kube.GetSecretList(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseSecretList(secrets)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func getSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Secret":          ctx.Param(secretParam),
	}).Debug("Get secret Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	secret, err := kube.GetSecret(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(secretParam))
	if err != nil {
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseSecret(secret)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func createSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secret kube_types.Secret
	if err := ctx.ShouldBindWith(&secret, binding.JSON); err != nil {
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

	newSecret := model.MakeSecret(ctx.Param(namespaceParam), secret, quota.Labels)

	secretAfter, err := kubecli.CreateSecret(newSecret)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseSecret(secretAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusCreated, ret)
}

func updateSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Secret":    ctx.Param(secretParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secret kube_types.Secret
	if err := ctx.ShouldBindWith(&secret, binding.JSON); err != nil {
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

	if ctx.Param(secretParam) != secret.Name {
		log.Errorf(invalidUpdateSecretName, ctx.Param(secretParam), secret.Name)
		ctx.Error(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateSecretName, ctx.Param(secretParam), secret.Name), http.StatusBadRequest))
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateSecretName, ctx.Param(secretParam), secret.Name), http.StatusBadRequest)))
		return
	}

	newSecret := model.MakeSecret(ctx.Param(namespaceParam), secret, quota.Labels)

	secretAfter, err := kubecli.UpdateSecret(newSecret)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseSecret(secretAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func deleteSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Secret":    ctx.Param(secretParam),
	}).Debug("Delete secret Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteSecret(ctx.Param(namespaceParam), ctx.Param(secretParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}
