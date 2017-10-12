package server

import (
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
)

func ListEndpoints(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	eptList, err := kubecli.CoreV1().Endpoints(nsname).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Endpoints.List error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": "cannot list endpoints: " + err.Error(),
		})
		return
	}

	redactEndpointsListForUser(eptList)
	c.JSON(200, eptList)
}

func CreateEndpoints(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	ept, ok := c.MustGet("requestObject").(*v1.Endpoints)
	if !ok || ept == nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "bad request",
		})
		return
	}
	if nsname != ept.ObjectMeta.Namespace {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in endpoints",
		})
		return
	}
	clientEndpointsInsertions(ept)

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	eptAfter, err := kubecli.CoreV1().Endpoints(ept.ObjectMeta.Namespace).Create(ept)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Endpoints.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create endpoints: %v", err),
		})
		return
	}

	redactEndpointsForUser(eptAfter)

	c.JSON(201, eptAfter)
}

func GetEndpoints(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	ept, err := kubecli.CoreV1().Endpoints(nsname).Get(objname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Endpoints.Get error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get endpoints %s: %v", objname, err),
		})
		return
	}
	redactEndpointsForUser(ept)
	c.JSON(200, ept)
}

func DeleteEndpoints(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	err := kubecli.CoreV1().Endpoints(nsname).Delete(objname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Endpoints.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot delete endpoints %s: %v", objname, err),
		})
		return
	}
	c.Status(204)
}

func redactEndpointsForUser(ept *v1.Endpoints) {
	ept.TypeMeta.Kind = "Endpoints"
	ept.TypeMeta.APIVersion = "v1"
}

func redactEndpointsListForUser(eptl *v1.EndpointsList) {
	for i := range eptl.Items {
		redactEndpointsForUser(&eptl.Items[i])
	}
}

func clientEndpointsInsertions(ept *v1.Endpoints) {
}
