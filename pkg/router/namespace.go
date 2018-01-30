package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	nsList := model.ParseResourceQuotaList(quotas)
	c.JSON(http.StatusOK, nsList)
}

func getNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Get namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	quota, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	ns := model.ParseResourceQuota(quota)
	c.JSON(http.StatusOK, ns)
}
