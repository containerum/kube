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
)

const (
	ownerQuery     = "owner"
	namespaceParam = "namespace"
)

func getNamespaceList(c *gin.Context) {
	log.WithField("Owner", c.Query(ownerQuery)).Debug("Get namespace list Call")

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	quotas, err := kube.GetNamespaceQuotaList(c.Query(ownerQuery))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, model.ParseResourceQuotaList(quotas))
}

func getNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Get namespace Call")

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	quota, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, model.ParseResourceQuota(quota))
}

func сreateNamespace(c *gin.Context) {
	log.Debug("Create namespace Call")

	var ns kube_types.Namespace
	if err := c.ShouldBindWith(&ns, binding.JSON); err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	nsAfter, err := kubecli.CreateNamespace(model.MakeNamespace(ns))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := model.MakeResourceQuota(ns.Resources.Hard.CPU, ns.Resources.Hard.Memory)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota.Labels = nsAfter.Labels
	quota.SetNamespace(ns.Name)
	quota.SetName("quota")
	quotaAfter, err := kubecli.CreateNamespaceQuota(ns.Name, quota)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
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
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	c.Status(http.StatusAccepted)
}

func updateNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Update namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	var res kube_types.UpdateNamespace
	if err := c.ShouldBindWith(&res, binding.JSON); err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quotaOld, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := model.MakeResourceQuota(res.Resources.Hard.CPU, res.Resources.Hard.Memory)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota.Labels = quotaOld.Labels
	quota.SetNamespace(c.Param(namespaceParam))
	quota.SetName("quota")
	quotaAfter, err := kube.UpdateNamespaceQuota(c.Param(namespaceParam), quota)
	if err != nil {
		c.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	c.JSON(http.StatusAccepted, model.ParseResourceQuota(quotaAfter))
}
