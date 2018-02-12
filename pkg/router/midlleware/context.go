package middleware

import (
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"github.com/gin-gonic/gin"
)

const (
	UserNamespaces = "user-namespaces"
	UserVolumes    = "user-volumes"
	UserRole       = "user-role"

	KubeClient = "kubernetes-client"

	NamespaceKey = "namespace"
	ServiceKey   = "service"
)

func RegisterKubeClient(kube *kubernetes.Kube) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(KubeClient, kube)
	}
}

func SetNamespace() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(NamespaceKey, c.Param(namespaceParam))
	}
}
