package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

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

// SubstitutionsFromHeadersFor takes an object from the context by the key objctxkey,
// if it is there, does the substitutions as specified in HTTP request headers
// and puts it back into the context by the same key. Parameter after controls whether
// the substitutions are performed after all other handlers in this chain are run.
// In such case, the substitutions are reversed
func SubstitutionsFromHeadersFor(objctxkey string, after bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reps map[string][]userReplacement // http header name -> array of userReplacement

		reps = make(map[string][]userReplacement)
		reps["x-user-namespace"] = nil
		reps["x-user-volume"] = nil

		// find, decode & validate data
		if c.Request.Header.Get("x-user-hide-data") == "" {
			return
		}
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

		if after {
			c.Next()

			// reverse the from-to parts
			for _, repArray := range reps {
				for i := range repArray {
					repArray[i].From, repArray[i].To = repArray[i].To, repArray[i].From
				}
			}
		}

		if obj, ok := c.Get(objctxkey); ok {
			switch objtyped := obj.(type) {
			case *v1.ConfigMap:
				subConfigMap(objtyped, reps)
			case *v1.ConfigMapList:
				for i := range objtyped.Items {
					subConfigMap(&objtyped.Items[i], reps)
				}

			case *v1beta1.Deployment:
				subDeployment(objtyped, reps)
			case *v1beta1.DeploymentList:
				for i := range objtyped.Items {
					subDeployment(&objtyped.Items[i], reps)
				}

			case *v1.Endpoints:
				subEndpoints(objtyped, reps)
			case *v1.EndpointsList:
				for i := range objtyped.Items {
					subEndpoints(&objtyped.Items[i], reps)
				}

			case *v1.Namespace:
				subNamespace(objtyped, reps)
			case *v1.NamespaceList:
				for i := range objtyped.Items {
					subNamespace(&objtyped.Items[i], reps)
				}

			case *v1.Secret:
				subSecret(objtyped, reps)
			case *v1.SecretList:
				for i := range objtyped.Items {
					subSecret(&objtyped.Items[i], reps)
				}

			case *v1.Service:
				subService(objtyped, reps)
			case *v1.ServiceList:
				for i := range objtyped.Items {
					subService(&objtyped.Items[i], reps)
				}

			case string:
				switch objtyped {
				case "namespace":
					for _, re := range reps["x-user-namespace"] {
						if !re.zero() && re.From == objtyped {
							c.Set("namespace", re.To)
						}
					}

				default:
					utils.Log(c).
						WithField("handler", "SubstitutionsFromHeadersFor").
						WithField("handler-objctxkey", objctxkey).
						WithField("handler-after", after).
						Infof("refusing to handle type %T", obj)
				}

			default:
				utils.Log(c).
					WithField("handler", "SubstitutionsFromHeadersFor").
					WithField("handler-objctxkey", objctxkey).
					WithField("handler-after", after).
					Infof("refusing to handle type %T", obj)
			}
			c.Set(objctxkey, obj)
		}
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
			// FIXME: must also substitute Deployment.Spec.Template.Spec.Containers[*].VolumeMounts[*].Name
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

func subConfigMap(objtyped *v1.ConfigMap, reps map[string][]userReplacement) {
	for _, re := range reps["x-user-namespace"] {
		if !re.zero() && re.From == objtyped.ObjectMeta.Namespace {
			objtyped.ObjectMeta.Namespace = re.To
		}
	}
}

func subSecret(objtyped *v1.Secret, reps map[string][]userReplacement) {
	for _, re := range reps["x-user-namespace"] {
		if !re.zero() && re.From == objtyped.ObjectMeta.Namespace {
			objtyped.ObjectMeta.Namespace = re.To
		}
	}
}
