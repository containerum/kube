package server

import (
	"fmt"
	"strconv"

	"git.containerum.net/ch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNamespaces(c *gin.Context) {
	kubecli := c.MustGet(KubeClientKey).(*kubernetes.Clientset)

	nsList, err := kubecli.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Namespaces.List error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot list namespaces: %v", err),
		})
		return
	}

	c.Status(200)
	c.Set(ResponseObjectKey, nsList)
}

// CreateNamespace uses query parameters 'cpu' and 'memory' to determine namespace quota.
func CreateNamespace(c *gin.Context) {
	ns, ok := c.MustGet(RequestObjectKey).(*v1.Namespace)
	if !ok {
		utils.Log(c).Warnf("request is not a Namespace")
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "not a Namespace",
		})
		return
	}
	ns.Spec = v1.NamespaceSpec{}

	quotacpu, err := strconv.Atoi(c.Query("cpu"))
	if err != nil {
		utils.Log(c).Warnf("invalid cpu format: %v", err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "invalid format for cpu quota",
		})
		return
	}
	quotamem, err := strconv.Atoi(c.Query("memory"))
	if err != nil {
		utils.Log(c).Warnf("invalid memory format: %v", err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "invalid format for memory quota",
		})
		return
	}

	kubecli := c.MustGet(KubeClientKey).(*kubernetes.Clientset)
	nsAfter, err := kubecli.CoreV1().Namespaces().Create(ns)
	if err != nil {
		utils.Log(c).Errorf("kubecli.Namespaces.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create namespace %s: %v", ns.ObjectMeta.Name, err),
		})
		return
	}
	quota := makeResourceQuota(quotacpu, quotamem)
	quota.SetNamespace(ns.ObjectMeta.Name)
	quota.SetName("quota")
	_, err = kubecli.CoreV1().ResourceQuotas(nsAfter.ObjectMeta.Name).Create(quota)
	if err != nil {
		utils.Log(c).Errorf("kubecli.ResourceQuota.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create namespace quota: %v", err),
		})
		return
	}

	c.Status(201)
	c.Set(ResponseObjectKey, nsAfter)
}

func GetNamespace(c *gin.Context) {
	nsname := c.MustGet(NamespaceKey).(string)
	kubecli := c.MustGet(KubeClientKey).(*kubernetes.Clientset)

	ns, err := kubecli.CoreV1().Namespaces().Get(nsname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Namespaces.Get error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get namespace %s: %v", nsname, err),
		})
		return
	}

	c.Status(200)
	c.Set(ResponseObjectKey, ns)
}

func GetNamespaceQuota(c *gin.Context) {
	nsname := c.MustGet(NamespaceKey).(string)
	kubecli := c.MustGet(KubeClientKey).(*kubernetes.Clientset)
	quota, err := kubecli.CoreV1().ResourceQuotas(nsname).Get("quota", meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.ResourceQuota.Get error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get quota \"quota\": %v", err),
		})
		return
	}

	c.Status(200)
	c.Set(ResponseObjectKey, quota)
}

func DeleteNamespace(c *gin.Context) {
	nsname := c.MustGet(NamespaceKey).(string)
	kubecli := c.MustGet(KubeClientKey).(*kubernetes.Clientset)

	err := kubecli.CoreV1().Namespaces().Delete(nsname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Namespaces.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot delete namespace %s: %v", nsname, err),
		})
		return
	}
	c.Status(204)
}

func makeResourceQuota(cpu, memory int) *v1.ResourceQuota {
	cpuq := api_resource.NewQuantity(int64(cpu), api_resource.DecimalSI)
	memoryq := api_resource.NewQuantity(int64(memory/1024/1024), api_resource.BinarySI)
	quota := &v1.ResourceQuota{
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceRequestsCPU:    *cpuq,
				v1.ResourceLimitsCPU:      *cpuq,
				v1.ResourceRequestsMemory: *memoryq,
				v1.ResourceLimitsMemory:   *memoryq,
			},
		},
	}
	return quota
}
