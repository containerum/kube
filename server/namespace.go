package server

import (
	"fmt"
	"strconv"

	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
)

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

func CreateNamespace(c *gin.Context) {
	ns_ := c.MustGet("requestObject")
	ns, ok := ns_.(*v1.Namespace)
	if !ok || ns == nil {
		utils.Log(c).Warnf("request is not a Namespace")
		c.JSON(400, "not a Namespace")
		return
	}

	quotacpu, err := strconv.Atoi(c.Query("cpu"))
	if err != nil {
		utils.Log(c).Warnf("invalid cpu format: %v", err)
		c.JSON(400, "invalid format for cpu quota")
		return
	}
	quotamem, err := strconv.Atoi(c.Query("memory"))
	if err != nil {
		utils.Log(c).Warnf("invalid memory format: %v", err)
		c.JSON(400, "invalid format for memory quota")
	}

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	nsAfter, err := kubecli.CoreV1().Namespaces().Create(ns)
	if err != nil {
		utils.Log(c).Errorf("kubecli.Namespaces.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot create namespace: %v", err),
		})
		return
	}
	quota := makeResourceQuota(quotacpu, quotamem)
	quota.SetNamespace(ns.ObjectMeta.Name)
	quota.SetName("quota")

	_, err = kubecli.CoreV1().ResourceQuotas(nsAfter.ObjectMeta.Name).Create(quota)
	if err != nil {
		utils.Log(c).Errorf("kubecli.ResourceQuota.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot create namespace quota: %v", err),
		})
		return
	}
}

func GetNamespace(c *gin.Context) {
}

func ListNamespaces(c *gin.Context) {
}

func UpdateNamespace(c *gin.Context) {
}

func DeleteNamespace(c *gin.Context) {
	c.AbortWithStatusJSON(500, map[string]string{
		"error": "not implemented",
	})
}
