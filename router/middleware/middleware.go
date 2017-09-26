package middleware

import (
	"encoding/json"
	"math/rand"
	"reflect"

	"bitbucket.org/exonch/kube-api/server"
	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"k8s.io/api/apps/v1beta1"
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
	c.Set("kubeclient", server.KubeClients[n].Client)
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
		var obj *v1beta1.Deployment
		err = json.Unmarshal(jsn, &obj)
		obj.Status = v1beta1.DeploymentStatus{}
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

func CheckHTTP411(c *gin.Context) {
	if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
		return
	}
	if c.Request.ContentLength < 0 {
		c.AbortWithStatus(411)
	}
}

// NOTE: обработчик выполняется на обратном ходе рекурсии.
func RedactResponseMetadata(c *gin.Context) {
	c.Next() // NOTE
	obj, ok := c.Get("responseObject")
	if !ok {
		return
	}
	jsn, _ := json.Marshal(obj)
	var m map[string]interface{}
	json.Unmarshal(jsn, &m)
	jsonDeleteInMetadata(m, "selfLink")
	jsonDeleteInMetadata(m, "uid")
	jsonDeleteInMetadata(m, "resourceVersion")

	var newobj interface{}
	t := reflect.TypeOf(obj)
	tt := t.Elem()
	v := reflect.New(tt)
	newobj = v.Interface()

	jsn, _ = json.Marshal(m)
	json.Unmarshal(jsn, newobj)
	c.Set("responseObject", newobj)
}

func jsonDeleteInMetadata(m map[string]interface{}, fieldName string) {
	for k, v := range m {
		if k == "metadata" {
			vmap := v.(map[string]interface{}) //"metadata" is always a JSON object
			for k2 := range vmap {
				if k2 == fieldName {
					delete(vmap, k2)
				}
			}
			m[k] = vmap
		} else if vmap, ok := v.(map[string]interface{}); ok {
			jsonDeleteInMetadata(vmap, fieldName)
			m[k] = vmap
		} else if varray, ok := v.([]interface{}); ok {
			for i := range varray {
				if vimap, ok := varray[i].(map[string]interface{}); ok {
					jsonDeleteInMetadata(vimap, fieldName)
				}
			}
			m[k] = varray
		}
	}
}

// NOTE: обработчик выполняется на обратном ходе рекурсии.
func WriteResponseObject(c *gin.Context) {
	c.Next() // NOTE
	obj, ok := c.Get("responseObject")
	if !ok {
		return
	}
	jsn, _ := json.Marshal(obj)
	c.Writer.Write(jsn)
}
