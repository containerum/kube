package middleware

import (
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	api_apps "k8s.io/api/apps/v1beta1"
	api_core "k8s.io/api/core/v1"
	api_ext "k8s.io/api/extensions/v1beta1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UserNamespaces = "user-namespaces"
	UserVolumes    = "user-volumes"
	UserRole       = "user-role"

	KubeClient = "kubernetes-client"

	RequestObjectKey  = "requestObject"
	ResponseObjectKey = "responseObject"
	NamespaceKey      = "namespace"
	ObjectNameKey     = "objectName"
	KubeClientKey     = "kubeclient"
)

func RegisterKubeClient(kube *kubernetes.Kube) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(KubeClient, kube)
	}
}

// ParseJSON parses a JSON payload into a kubernetes struct of appropriate
// type and sets it into the gin context under the server.RequestObjectKey key.
func ParseJSON(c *gin.Context) {
	var kind struct {
		Kind string `json:"kind"`
	}

	jsn, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error":   "bad request",
			"errcode": "BAD_INPUT",
		})
		return
	}

	err = jsoniter.Unmarshal(jsn, &kind)
	if err != nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error":   "json error: " + err.Error(),
			"errcode": "BAD_INPUT",
		})
		return
	}

	var objmeta *api_meta.ObjectMeta

	switch kind.Kind {
	case "ConfigMap":
		var obj *api_core.ConfigMap
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	case "Deployment":
		var obj *api_apps.Deployment
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	case "Endpoints":
		var obj *api_core.Endpoints
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	case "Namespace":
		var obj *api_core.Namespace
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	case "Secret":
		var obj *api_core.Secret
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	case "Service":
		var obj *api_core.Service
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	case "Ingress":
		var obj *api_ext.Ingress
		err = jsoniter.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set(RequestObjectKey, obj)

	default:
		c.Set(RequestObjectKey, jsoniter.RawMessage(jsn))
	}

	if _, ok := c.Get(NamespaceKey); !ok {
		if kind.Kind == "Namespace" {
			c.Set(NamespaceKey, objmeta.Name)
		} else {
			c.Set(NamespaceKey, objmeta.Namespace)
		}
	}

	if _, ok := c.Get(ObjectNameKey); !ok {
		if kind.Kind != "Namespace" {
			c.Set(ObjectNameKey, objmeta.Name)
		}
	}
}
