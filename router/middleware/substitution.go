package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// userReplacement represents the relation between a user-visible
// resource name (label), and our internal name for that resource (id).
type userReplacement struct {
	From string `json:"label"`
	To   string `json:"id"`
}

func (re userReplacement) zero() bool {
	return re.To == "" || re.From == ""
}

// NOTE: обработчик выполняется на прямом и обратном ходе рекурсии.
func SubstitutionsFromHeaders(c *gin.Context) {
	var reps map[string][]userReplacement // http header name -> array of userReplacement

	reps = make(map[string][]userReplacement)
	reps["x-user-namespace"] = nil
	reps["x-user-volume"] = nil

	// find, decode & validate data
	for hdr, repArray := range reps {
		var b64 string
		if b64 = c.Request.Header.Get(hdr); b64 == "" {
			continue
		}
		jsn, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			logrus.Errorf("cannot b64-decode header %q: %v", hdr, err)
			c.AbortWithStatus(400)
			c.Set("responseObject", map[string]string{
				"error": fmt.Sprintf("cannot decode %q: %v", hdr, err),
			})
			return
		}
		err = json.Unmarshal(jsn, &repArray)
		if err != nil {
			logrus.Errorf("cannot parse json from header %q: %v", hdr, err)
			c.AbortWithStatus(400)
			c.Set("responseObject", map[string]string{
				"error": fmt.Sprintf("cannot parse %q: %v", hdr, err),
			})
			return
		}
		reps[hdr] = repArray
	}

	// >>> sub INCOMING data >>>
	if obj, ok := c.Get("requestObject"); ok {
		switch objtyped := obj.(type) {
		case *v1.Namespace:
			subNamespace(objtyped, reps)
		case *v1.NamespaceList:
			for i := range objtyped.Items {
				subNamespace(&objtyped.Items[i], reps)
			}
		case *v1beta1.Deployment:
			subDeployment(objtyped, reps)
		case *v1beta1.DeploymentList:
			for i := range objtyped.Items {
				subDeployment(&objtyped.Items[i], reps)
			}
		case *v1.Service:
			subService(objtyped, reps)
		case *v1.ServiceList:
			for i := range objtyped.Items {
				subService(&objtyped.Items[i], reps)
			}
		case *v1.Endpoints:
			subEndpoints(objtyped, reps)
		case *v1.EndpointsList:
			for i := range objtyped.Items {
				subEndpoints(&objtyped.Items[i], reps)
			}
		default:
			logrus.Panicf("refusing to handle incoming type %T", obj)
		}
		//c.Set("requestObject", obj) // no need as the types are pointers
	}

	// reverse the from-to parts
	for _, repArray := range reps {
		for i := range repArray {
			repArray[i].From, repArray[i].To = repArray[i].To, repArray[i].From
		}
	}

	// invoke further handerls in the chain
	c.Next()

	// <<< sub OUTGOING data <<<
	if obj, ok := c.Get("responseObject"); ok {
		switch objtyped := obj.(type) {
		case *v1.Namespace:
			subNamespace(objtyped, reps)
		case *v1.NamespaceList:
			for i := range objtyped.Items {
				subNamespace(&objtyped.Items[i], reps)
			}
		case *v1beta1.Deployment:
			subDeployment(objtyped, reps)
		case *v1beta1.DeploymentList:
			for i := range objtyped.Items {
				subDeployment(&objtyped.Items[i], reps)
			}
		case *v1.Service:
			subService(objtyped, reps)
		case *v1.ServiceList:
			for i := range objtyped.Items {
				subService(&objtyped.Items[i], reps)
			}
		case *v1.Endpoints:
			subEndpoints(objtyped, reps)
		case *v1.EndpointsList:
			for i := range objtyped.Items {
				subEndpoints(&objtyped.Items[i], reps)
			}
		case map[string]string:
		default:
			logrus.Panicf("refusing to handle type %T", obj)
		}
		//c.Set("responseObject", obj) // no need as the types are pointers
	}
}

func subNamespace(objtyped *v1.Namespace, reps map[string][]userReplacement) {
	for _, re := range reps["x-user-namespace"] {
		if !re.zero() && re.From == objtyped.ObjectMeta.Name {
			objtyped.ObjectMeta.Name = re.To
		}
	}
}

func subDeployment(objtyped *v1beta1.Deployment, reps map[string][]userReplacement) {
	for _, re := range reps["x-user-namespace"] {
		if !re.zero() && re.From == objtyped.ObjectMeta.Namespace {
			objtyped.ObjectMeta.Namespace = re.To
		}
	}
	for _, re := range reps["x-user-volume"] {
		if !re.zero() {
			for i := range objtyped.Spec.Template.Spec.Volumes {
				vol := &objtyped.Spec.Template.Spec.Volumes[i]
				if re.From == vol.Name {
					vol.Name = re.To
				}
			}
		}
	}
}

func subService(objtyped *v1.Service, reps map[string][]userReplacement) {
	for _, re := range reps["x-user-namespace"] {
		if !re.zero() && re.From == objtyped.ObjectMeta.Namespace {
			objtyped.ObjectMeta.Namespace = re.To
		}
	}
}

func subEndpoints(objtyped *v1.Endpoints, reps map[string][]userReplacement) {
	for _, re := range reps["x-user-namespace"] {
		if !re.zero() && re.From == objtyped.ObjectMeta.Namespace {
			objtyped.ObjectMeta.Namespace = re.To
		}
	}
}
