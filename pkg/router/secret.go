package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	json_types "git.containerum.net/ch/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	secretParam = "secret"
)

func getSecretList(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": c.Param(namespaceParam),
		"Namespace":       c.MustGet(m.NamespaceKey).(string),
	}).Debug("Get secret list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	secrets, err := kube.GetSecretList(c.MustGet(m.NamespaceKey).(string))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, secrets)
}

func getSecret(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": c.Param(namespaceParam),
		"Namespace":       c.MustGet(m.NamespaceKey).(string),
		"Secret":          c.Param(secretParam),
	}).Debug("Get secret Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	secrets, err := kube.GetSecret(c.MustGet(m.NamespaceKey).(string), c.Param(secretParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, secrets)
}

func createSecret(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secret *json_types.Secret
	if err := c.ShouldBindWith(&secret, binding.JSON); err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newSecret := model.MakeSecret(c.Param(namespaceParam), *secret)

	secretAfter, err := kubecli.CreateSecret(newSecret)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusCreated, secretAfter)
}

func deleteSecret(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Secret":    c.Param(secretParam),
	}).Debug("Delete secret Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteSecret(c.Param(namespaceParam), c.Param(secretParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.Status(http.StatusAccepted)
}
