package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin/binding"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	ownerQuery     = "owner"
	namespaceParam = "namespace"
	serviceParam   = "service"
)

func getNamespaceList(c *gin.Context) {
	log.WithField("Owner", c.Query(ownerQuery)).Debug("Get namespace list Call")

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	quotas, err := kube.GetNamespaceQuotaList(c.Query(ownerQuery))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, model.ParseResourceQuotaList(quotas))
}

func getNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Get namespace Call")

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	quota, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, model.ParseResourceQuota(quota))
}

func сreateNamespace(c *gin.Context) {
	log.Debug("Create namespace Call")

	var ns kube_types.Namespace
	if err := c.ShouldBindWith(&ns, binding.JSON); err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	cpuq, err := api_resource.ParseQuantity(ns.Resources.Hard.CPU)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(model.NewErrorWithCode(invalidCPUFormat, http.StatusBadRequest)))
		return
	}
	memoryq, err := api_resource.ParseQuantity(ns.Resources.Hard.Memory)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(model.NewErrorWithCode(invalidMemoryFormat, http.StatusBadRequest)))
		return
	}

	nsAfter, err := kubecli.CreateNamespace(model.MakeNamespace(ns))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	quota := model.MakeResourceQuota(cpuq, memoryq)
	quota.Labels = nsAfter.Labels
	quota.SetNamespace(ns.Name)
	quota.SetName("quota")
	quotaAfter, err := kubecli.CreateNamespaceQuota(ns.Name, quota)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	quotaAfter.Labels = nsAfter.Labels

	c.JSON(http.StatusCreated, model.ParseResourceQuota(quotaAfter))
}

func deleteNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Delete namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteNamespace(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.Status(http.StatusAccepted)
}

func updateNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Update namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	var res kube_types.UpdateNamespace
	if err := c.ShouldBindWith(&res, binding.JSON); err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	cpuq, err := api_resource.ParseQuantity(res.Resources.Hard.CPU)
	if err != nil {
		log.Errorln(invalidCPUFormat, err)
		c.AbortWithStatusJSON(ParseErorrs(model.NewErrorWithCode(invalidCPUFormat, http.StatusBadRequest)))
		return
	}
	memoryq, err := api_resource.ParseQuantity(res.Resources.Hard.Memory)
	if err != nil {
		log.Errorln(invalidMemoryFormat, err)
		c.AbortWithStatusJSON(ParseErorrs(model.NewErrorWithCode(invalidMemoryFormat, http.StatusBadRequest)))
		return
	}

	quotaOld, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	quota := model.MakeResourceQuota(cpuq, memoryq)
	quota.Labels = quotaOld.Labels
	quota.SetNamespace(c.Param(namespaceParam))
	quota.SetName("quota")
	quotaAfter, err := kube.UpdateNamespaceQuota(c.Param(namespaceParam), quota)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusAccepted, model.ParseResourceQuota(quotaAfter))
}
