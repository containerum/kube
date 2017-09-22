package middleware

import (
	"encoding/json"
	"math/rand"

	"bitbucket.org/exonch/kube-api/server"
	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
)

func SetNamespace(c *gin.Context) {
	ns := c.Param("namespace")
	if ns == "" {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": `missing namespace name`,
		})
	}
	c.Set("namespace", ns)
}

func SetObjectName(c *gin.Context) {
	objname := c.Param("objname")
	if objname == "" {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": `missing object name`,
		})
	}
	c.Set("objectName", objname)
}

func SetRandomKubeClient(c *gin.Context) {
	if len(server.KubeClients) == 0 {
		c.AbortWithStatusJSON(500, map[string]string{
			"error": "no available kubernetes apiserver clients",
		})
		return
	}

	n := rand.Intn(len(server.KubeClients))
	utils.Log(c).Infof("picked server.KubeClients[%d]", n)
	utils.AddLogField(c, "kubeclient-address", server.KubeClients[n].Tag)
	server.KubeClients[n].UseCount++
	c.Set("kubeclient", server.KubeClients[n])
}

// ParseJSON parses a JSON payload into a kubernetes struct of appropriate
// type and sets it into the gin context under the "requestObject" key.
func ParseJSON(c *gin.Context) {
	var kind struct {
		Kind string `json:"kind"`
	}

	jsn, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(999, map[string]string{"error": "bad request"})
		return
	}

	err = json.Unmarshal(jsn, &kind)
	if err != nil {
		return
	}
	switch kind.Kind {
	case "Namespace":
		var obj *v1.Namespace
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Deployment":
		var obj *v1beta2.Deployment
		err = json.Unmarshal(jsn, &obj)
		obj.Status = v1beta2.DeploymentStatus{}
		c.Set("requestObject", obj)
	case "Service":
		var obj *v1.Service
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Endpoints":
		var obj *v1.Endpoints
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	}
}

func SetRequestID(c *gin.Context) {
	reqid := uuid.NewV4().String()
	c.Set("request-id", reqid)
	c.Header("X-Request-ID", reqid)
}
