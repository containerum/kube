package server

import (
	"fmt"

	"git.containerum.net/ch/kube-api/utils"

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

func GetIngress(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)

	ingress, err := kubecli.ExtensionsV1beta1().Ingresses(nsname).Get(objname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Ingresses.Get error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get ingress %s: %v", objname, err),
		})
		return
	}
	redactIngressForUser(ingress)
	c.Status(200)
	c.Set("responseObject", ingress)
}

func DeleteIngress(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)

	err := kubecli.ExtensionsV1beta1().Ingresses(nsname).Delete(objname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Ingresses.Delete error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot delete ingress %s: %v", objname, err),
		})
		return
	}
	c.Status(204)
}

func redactIngressForUser(ing *v1beta1.Ingress) {
	ing.TypeMeta.Kind = "Ingress"
	ing.TypeMeta.APIVersion = "extensions/v1beta1"
}

func redactIngressListForUser(ingList *v1beta1.IngressList) {
	for i := range ingList.Items {
		redactIngressForUser(&ingList.Items[i])
	}
}

func clientIngressInsertions(ing *v1beta1.Ingress) {
}
