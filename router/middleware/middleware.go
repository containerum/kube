package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"

	"git.containerum.net/ch/kube-api/access"
	"git.containerum.net/ch/kube-api/server"
	"git.containerum.net/ch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	//"github.com/sirupsen/logrus"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetNamespace(c *gin.Context) {
	ns := c.Param("namespace")
	if ns == "" {
		c.AbortWithStatusJSON(400, map[string]string{
			"error":   `missing namespace name`,
			"errcode": "BAD_INPUT",
		})
	}
	c.Set("namespace", ns)
}

func SetObjectName(c *gin.Context) {
	objname := c.Param("objname")
	if objname == "" {
		c.AbortWithStatusJSON(400, map[string]string{
			"error":   `missing object name`,
			"errcode": "BAD_INPUT",
		})
	}
	c.Set("objectName", objname)
}

func SetRandomKubeClient(c *gin.Context) {
	if len(server.KubeClients) == 0 {
		c.AbortWithStatusJSON(500, map[string]string{
			"error":   "no available kubernetes apiserver clients",
			"errcode": "INTERNAL",
		})
		return
	}

	n := rand.Intn(len(server.KubeClients))
	utils.Log(c).Debugf("picked server.KubeClients[%d]", n)
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
		c.AbortWithStatusJSON(400, map[string]string{
			"error":   "bad request",
			"errcode": "BAD_INPUT",
		})
		return
	}

	err = json.Unmarshal(jsn, &kind)
	if err != nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error":   "json error: " + err.Error(),
			"errcode": "BAD_INPUT",
		})
		return
	}

	var objmeta *meta_v1.ObjectMeta

	switch kind.Kind {
	case "ConfigMap":
		var obj *v1.ConfigMap
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	case "Deployment":
		var obj *v1beta1.Deployment
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	case "Endpoints":
		var obj *v1.Endpoints
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	case "Namespace":
		var obj *v1.Namespace
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	case "Secret":
		var obj *v1.Secret
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	case "Service":
		var obj *v1.Service
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	case "Ingress":
		var obj *ext_v1beta1.Ingress
		err = json.Unmarshal(jsn, &obj)
		objmeta = &obj.ObjectMeta
		c.Set("requestObject", obj)

	default:
		c.Set("requestObject", json.RawMessage(jsn))
	}

	if _, ok := c.Get("namespace"); !ok {
		if kind.Kind == "Namespace" {
			c.Set("namespace", objmeta.Name)
		} else {
			c.Set("namespace", objmeta.Namespace)
		}
	}

	if _, ok := c.Get("objectName"); !ok {
		if kind.Kind != "Namespace" {
			c.Set("objectName", objmeta.Name)
		}
	}
}

func SetRequestID(c *gin.Context) {
	reqid := uuid.NewV4().String()
	c.Set("request-id", reqid)
	c.Header("X-Request-ID", reqid)
}

// CheckHTTP411 middleware rejects all submissions with unknown length.
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
	jsonDeleteInMetadata(m, "selfLink", fmt.Sprintf("%p", m))
	jsonDeleteInMetadata(m, "uid", fmt.Sprintf("%p", m))
	//jsonDeleteInMetadata(m, "resourceVersion", fmt.Sprintf("%p", m))

	var newobj interface{}
	t := reflect.TypeOf(obj)
	tt := t.Elem() //t should always be a pointer, so this should work w/o problems
	v := reflect.New(tt)
	newobj = v.Interface()

	jsn, _ = json.Marshal(m)
	json.Unmarshal(jsn, newobj)
	c.Set("responseObject", newobj)
}

func jsonDeleteInMetadata(obj interface{}, fieldName string, trace string) {
	//logrus.Infof("in %s", trace)
	switch objTyped := obj.(type) {
	case map[string]interface{}:
		//logrus.Infof("type object")
		for k, v := range objTyped {
			if k == "metadata" {
				vmap := v.(map[string]interface{})
				for Mk := range vmap {
					//logrus.Infof("considering %s from %s -> metadata", fieldName, trace)
					if Mk == fieldName {
						//logrus.Infof("deleting %s from %s -> metadata", fieldName, trace)
						delete(vmap, Mk)
					}
				}
				objTyped[k] = vmap
			} else {
				jsonDeleteInMetadata(v, fieldName, trace+" -> "+k)
			}
		}
	case []interface{}:
		//logrus.Infof("type array")
		for i := range objTyped {
			jsonDeleteInMetadata(objTyped[i], fieldName, fmt.Sprintf("%s -> %d", trace, i))
		}
	default:
		//logrus.Infof("skip")
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

func SwapInputOutput(c *gin.Context) {
	in, ok1 := c.Get("requestObject")
	out, ok2 := c.Get("responseObject")
	if ok1 {
		c.Set("responseObject", in)
	}
	if ok2 {
		c.Set("requestObject", out)
	}
}

func ParseUserData(c *gin.Context) {
	hheaders := &access.HTTPHeaders{}

	if hdrdat := c.Request.Header.Get("x-user-namespace"); len(hdrdat) > 0 {
		unb64, err := base64.StdEncoding.DecodeString(hdrdat)
		if err != nil {
			utils.Log(c).Warnf("invalid base64 in header x-user-namespace: %v", err)
			c.AbortWithStatusJSON(400, map[string]string{
				"error":   "invalid base64 in header x-user-namespace",
				"errcode": "BAD_INPUT",
			})
			return
		}

		err = json.Unmarshal(unb64, &hheaders.Namespace)
		if err != nil {
			utils.Log(c).Warnf("cannot unmarshal json in header x-user-namespace: %v (%[1]T)", err)
			c.AbortWithStatusJSON(400, map[string]string{
				"error":   "invalid json in header x-user-namespace",
				"errcode": "BAD_INPUT",
			})
			return
		}
	}

	if hdrdat := c.Request.Header.Get("x-user-volume"); len(hdrdat) > 0 {
		unb64, err := base64.StdEncoding.DecodeString(c.Request.Header.Get("x-user-volume"))
		if err != nil {
			utils.Log(c).Warnf("invalid base64 in header x-user-volume: %v", err)
			c.AbortWithStatusJSON(400, map[string]string{
				"error":   "invalid base64 in header x-user-volume",
				"errcode": "BAD_INPUT",
			})
			return
		}

		err = json.Unmarshal(unb64, &hheaders.Volume)
		if err != nil {
			utils.Log(c).Warnf("cannot unmarshal json in header x-user-volume: %v", err)
			c.AbortWithStatusJSON(400, map[string]string{
				"error":   "invalid json in header x-user-volume",
				"errcode": "BAD_INPUT",
			})
			return
		}
	}

	c.Set("userData", hheaders)
}
