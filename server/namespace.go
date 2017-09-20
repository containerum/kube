package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
	core_v1 "k8s.io/api/core/v1"
	api_resource "k8s.io/apimachinery/pkg/api/resource"
)

func makeResourceQuota(cpu, memory int) *v1.ResourceQuota {
	cpuq := api_resource.NewQuantity(cpu, api_resource.DecimalSI)
	memoryq := api_resource.NewQuantity(memory/1024/1024, api_resource.BinarySI)
	quota := &core_v1.ResourceQuota{
		Spec: core_v1.ResourceQuotaSpec{
			Hard: core_v1.ResourceList{
				core_v1.ResourceRequestsCPU:    cpuq,
				core_v1.ResourceLimitsCPU:      cpuq,
				core_v1.ResourceRequestsMemory: memoryq,
				core_v1.ResourceLimitsMemory:   memoryq,
			},
		},
	}
	return quota
}

func CreateNamespace(c *gin.Context) {
	ns_ := c.MustGet("requestObject")
	ns, ok := ns_.(*core_v1.Namespace)
	if !ok || ns == nil {
		logWithContext(c).Warnf("request is not a Namespace")
		c.JSON(400, "not a Namespace")
		return
	}

	quotacpu, err := strconv.Atoi(c.Query("cpu"))
	if err != nil {
		logWithContext(c).Warnf("invalid cpu format: %v", err)
		c.JSON(400, "invalid format for cpu quota")
		return
	}
	quotamem, err := strconv.Atoi(c.Query("memory"))
	if err != nil {
		logWithContext(c).Warnf("invalid memory format: %v", err)
		c.JSON(400, "invalid format for memory quota")
	}

	kubecli := c.server.GetKubeClient()
	nsAfter, err := kubecli.CoreV1().Namespaces().Create(ns)
	if err != nil {
		c.log.Errorf("kubecli.Namespaces.Create error: %[1]T %[1]v", err)
		c.ErrorJSON(503, "E2")
		return
	}
	quota := makeResourceQuota(quotacpu, qoutamem)
	quota.SetNamespace(ns.ObjectMeta.Name)
	quota.SetName("quota")

	quotaAfter, err := kubecli.CoreV1().ResourceQuota(nsAfter.ObjectMeta.Name).
		Create(quota)
	if err != nil {
	}
}

func GetNamespace(c *gin.Context) {
}

func ListNamespaces(c *gin.Context) {
}

func UpdateNamespace(c *gin.Context) {
}

func RemoveNamespace(c *gin.Context) {
}
