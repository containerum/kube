package server

import (
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
)

func ListIngresses(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	ingressList, err := kubecli.ExtensionsV1beta1().Ingresses(nsname).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Ingresses.List error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": "cannot list ingresses: " + err.Error(),
		})
		return
	}

	redactIngressListForUser(ingressList)
	c.Status(200)
	c.Set("responseObject", ingressList)
}

func CreateIngress(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	ingress, ok := c.MustGet("requestObject").(*v1beta1.Ingress)
	if !ok || ingress == nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "bad request",
		})
		return
	}
	if nsname != ingress.ObjectMeta.Namespace {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in ingress",
		})
		return
	}
	clientIngressInsertions(ingress)

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	ingressAfter, err := kubecli.ExtensionsV1beta1().Ingresses(ingress.ObjectMeta.Namespace).Create(ingress)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Ingresses.Create error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create ingress: %v", err),
		})
		return
	}

	redactIngressForUser(ingressAfter)

	c.Status(201)
	c.Set("responseObject", ingressAfter)
}

func DeleteIngress(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	err := kubecli.CoreV1().Secrets(nsname).Delete(objname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Secrets.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot delete endpoints %s: %v", objname, err),
		})
		return
	}
	c.Status(204)
}

func redactIngressForUser(ing *v1beta1.Ingress) {
}

func redactIngressListForUser(ingList *v1beta1.IngressList) {
}

func clientIngressInsertions(ing *v1beta1.Ingress) {
}
