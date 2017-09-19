package http

import (
	"github.com/gin-gonic/gin"
	core_v1 "k8s.io/api/core/v1"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
)

func NamespaceCreate(c *cmdContext) {
	c.log = c.log.WithField("handler", "NamespaceCreate")
	c.log.Infof("NamespaceCreate start")
	defer c.log.Info("NamespaceCreate end")

	ns_, kind, err := parseJSON(c.body)
	if err != nil {
		c.log.Warnf("parseJSON error: %[1]T: %[1]v", err)
		c.JSON(500, map[string]string{"error": "E1"})
		return
	}
	ns := ns_.(*core_v1.Namespace)

	kubecli := c.server.GetKubeClient()
	nsAfter, err := kubecli.CoreV1().Namespaces().Create(ns)
	if err != nil {
		c.log.Errorf("kubecli.Namespaces.Create error: %[1]T %[1]v", err)
		c.JSON(503, map[string]string{"error": "E2"})
		return
	}
}
