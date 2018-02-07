package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	json_types "git.containerum.net/ch/kube-client/pkg/model"

	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	//	api_core "k8s.io/api/core/v1"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, ParseErorrs(err))
		return
	}
	ret := model.ParseResourceQuotaList(quotas)
	c.JSON(http.StatusOK, ret)
}

func getNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Get namespace Call")

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	quota, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, ParseErorrs(err))
		return
	}
	ret := model.ParseResourceQuota(quota)
	c.JSON(http.StatusOK, ret)
}

func сreateNamespace(c *gin.Context) {
	log.Debug("Create namespace Call")

	var ns json_types.Namespace
	if err := c.ShouldBindWith(&ns, binding.JSON); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ParseErorrs(err))
		return
	}

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	cpuq, err := api_resource.ParseQuantity(c.Query("cpu"))
	if err != nil {
		log.Errorln(invalidCPUFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ParseErorrs(errors.New(fmt.Sprintf(invalidCPUFormat, err.Error()))))
		return
	}
	memoryq, err := api_resource.ParseQuantity(c.Query("memory"))
	if err != nil {
		log.Errorln(invalidMemoryFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ParseErorrs(errors.New(fmt.Sprintf(invalidMemoryFormat, err.Error()))))
		return
	}

	nsAfter, err := kubecli.CreateNamespace(&ns)
	if err != nil {
		log.Errorf(namespaceCreationError, ns.Name, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ParseErorrs(errors.New(fmt.Sprintf(namespaceCreationError, ns.Name, err.Error()))))
		return
	}

	quota := model.MakeResourceQuota(cpuq, memoryq)
	quota.Labels = nsAfter.Labels
	quotaAfter, err := kubecli.CreateNamespaceQuota(ns.Name, quota)
	if err != nil {
		log.Errorf(namespaceQuotaCreationError, ns.Name, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ParseErorrs(errors.New(fmt.Sprintf(namespaceQuotaCreationError, ns.Name, err.Error()))))
		return
	}

	quotaAfter.Labels = nsAfter.Labels

	ret := model.ParseResourceQuota(quotaAfter)

	c.Set(m.ResponseObjectKey, ret)
	c.JSON(http.StatusCreated, ret)
}

func deleteNamespace(c *gin.Context) {
	log.WithField("Namespace", c.Param(namespaceParam)).Debug("Delete namespace Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteNamespace(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ParseErorrs(err))
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
		c.AbortWithStatusJSON(http.StatusBadRequest, ParseErorrs(errors.New(fmt.Sprintf(invalidCPUFormat, ParseErorrs(err)))))
		return
	}
	memoryq, err := api_resource.ParseQuantity(c.Query("memory"))
	if err != nil {
		log.Errorln(invalidMemoryFormat, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ParseErorrs(errors.New(fmt.Sprintf(invalidMemoryFormat, ParseErorrs(err)))))
		return
	}

	quotaOld, err := kube.GetNamespaceQuota(c.Param(namespaceParam))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, ParseErorrs(err))
		return
	}

	quota := model.MakeResourceQuota(cpuq, memoryq)
	quota.Labels = quotaOld.Labels
	quotaAfter, err := kube.UpdateNamespaceQuota(c.Param(namespaceParam), quota)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ParseErorrs(err))
		return
	}

	ns := model.ParseResourceQuota(quotaAfter)

	c.Set(m.ResponseObjectKey, ns)
	c.JSON(http.StatusAccepted, ns)
}
