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

	RequestObjectKey  = "requestObject"
	ResponseObjectKey = "responseObject"
	NamespaceKey      = "namespace"
	ServiceKey        = "service"
	ObjectNameKey     = "objectName"
	KubeClientKey     = "kubeclient"
)

func RegisterKubeClient(kube *kubernetes.Kube) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(KubeClient, kube)
	}
}
