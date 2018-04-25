package middleware

import (
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"github.com/gin-gonic/gin"
)

const (
	UserNamespaces = "user-namespaces"
	UserVolumes    = "user-volumes"
	UserRole       = "user-role"
	UserID         = "user-id"

	KubeClient = "kubernetes-client"

	NamespaceKey      = "namespace"
	NamespaceLabelKey = "namespace-key"
)

func RegisterKubeClient(kube *kubernetes.Kube) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(KubeClient, kube)
	}
}

func SetNamespace() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(NamespaceKey, c.Param(namespaceParam))
		c.Set(NamespaceLabelKey, "")
	}
}
