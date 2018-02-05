package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	api_core "k8s.io/api/core/v1"
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
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
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	ns := model.ParseResourceQuota(quota)
	c.JSON(http.StatusOK, ns)
}

func сreateNamespace(c *gin.Context) {
	log.Debug("Create namespace Call")

	var ns *api_core.Namespace
	if err := c.ShouldBindJSON(&ns); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	ns.Spec = api_core.NamespaceSpec{}

	cpuq, err := api_resource.ParseQuantity(c.Query("cpu"))
	if err != nil {
		log.Errorln(invalidCPUFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf(invalidCPUFormat, err.Error()))
		return
	}
	memoryq, err := api_resource.ParseQuantity(c.Query("memory"))
	if err != nil {
		log.Errorln(invalidMemoryFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf(invalidMemoryFormat, err.Error()))
		return
	}

	nsAfter, err := kubecli.CreateNamespace(ns)
	if err != nil {
		log.Errorf(namespaceCreationError, ns.ObjectMeta.Name, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf(namespaceCreationError, ns.ObjectMeta.Name, err.Error()))
		return
	}

	quota := kubernetes.MakeResourceQuota(cpuq, memoryq)
	err = kubecli.CreateNamespaceQuota(ns.Name, quota)
	if err != nil {
		log.Errorf(namespaceQuotaCreationError, ns.ObjectMeta.Name, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf(namespaceQuotaCreationError, ns.ObjectMeta.Name, err.Error()))
		return
	}

	c.Set(m.ResponseObjectKey, nsAfter)
	c.JSON(http.StatusCreated, nsAfter)
}

func deleteNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Delete namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteNamespace(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func updateNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Update namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	cpuq, err := api_resource.ParseQuantity(c.Query("cpu"))
	if err != nil {
		log.Errorln(invalidCPUFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf(invalidCPUFormat, err.Error()))
		return
	}
	memoryq, err := api_resource.ParseQuantity(c.Query("memory"))
	if err != nil {
		log.Errorln(invalidMemoryFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf(invalidMemoryFormat, err.Error()))
		return
	}

	quota := kubernetes.MakeResourceQuota(cpuq, memoryq)

	quotaAfter, err := kube.UpdateNamespaceQuota(c.Param(namespaceParam), quota)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ns := model.ParseResourceQuota(quotaAfter)

	c.Set(m.ResponseObjectKey, ns)
	c.JSON(http.StatusAccepted, ns)
}
