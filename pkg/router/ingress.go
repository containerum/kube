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
	ingressParam = "ingress"
)

func getIngressList(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingressList, err := kubecli.GetIngressList(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	c.JSON(http.StatusOK, model.ParseIngressList(ingressList))
}

func getIngress(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingress, err := kubecli.GetIngress(c.Param(namespaceParam), c.Param(ingressParam))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	c.JSON(http.StatusOK, model.ParseIngress(ingress))
}

func createIngress(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress kube_types.Ingress
	if err := c.ShouldBindWith(&ingress, binding.JSON); err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	newIngress := model.MakeIngress(ingress)

	ingressAfter, err := kubecli.CreateIngress(c.Param(namespaceParam), newIngress)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	c.JSON(http.StatusCreated, model.ParseIngress(ingressAfter))
}

func updateIngress(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Ingress":   c.Param(ingressParam),
	}).Debug("Create secret Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingress kube_types.Ingress
	if err := c.ShouldBindWith(&ingress, binding.JSON); err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	if c.Param(ingressParam) != ingress.Name {
		log.Errorf(invalidUpdateIngressName, c.Param(ingressParam), ingress.Name)
		c.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateIngressName, c.Param(ingressParam), ingress.Name), http.StatusBadRequest)))
		return
	}

	newIngress := model.MakeIngress(ingress)

	ingressAfter, err := kubecli.UpdateIngress(c.Param(namespaceParam), newIngress)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	c.JSON(http.StatusAccepted, model.ParseIngress(ingressAfter))
}

func deleteIngress(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kubecli.DeleteIngress(c.Param(namespaceParam), c.Param(ingressParam))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	c.Status(http.StatusAccepted)
}
