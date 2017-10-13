package server

import (
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
)

func ListServices(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	svcList, err := kubecli.CoreV1().Services(nsname).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Services.List error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": "cannot list services: " + err.Error(),
		})
		return
	}

	redactServiceListForUser(svcList)
	c.Status(200)
	c.Set("responseObject", svcList)
}

func CreateService(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	svc, ok := c.MustGet("requestObject").(*v1.Service)
	if !ok || svc == nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "bad request",
		})
		return
	}
	if nsname != svc.ObjectMeta.Namespace {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in deployment",
		})
		return
	}
	clientServiceInsertions(svc)

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	svcAfter, err := kubecli.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Services.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create service: %v", err),
		})
		return
	}

	redactServiceForUser(svcAfter)

	c.Status(201)
	c.Set("responseObject", svcAfter)
}

func GetService(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	svc, err := kubecli.CoreV1().Services(nsname).Get(objname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Services.Get error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get service %s: %v", objname, err),
		})
		return
	}
	redactServiceForUser(svc)
	c.Status(200)
	c.Set("responseObject", svc)
}

func ReplaceService(c *gin.Context) {
	var err error
	nsname := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	svc := c.MustGet("requestObject").(*v1.Service)

	if nsname != svc.ObjectMeta.Namespace {
		err = fmt.Errorf("namespace name mismatch (url %q, service %q)",
			nsname, svc.ObjectMeta.Namespace)
	}
	if err != nil {
		utils.Log(c).Warnf("service %s: %v", svc.ObjectMeta.Name, err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": err.Error(),
		})
		return
	}

	clientServiceInsertions(svc)

	svcAfter, err := kubecli.CoreV1().Services(nsname).Update(svc)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Services.Update %s: %v", svc.ObjectMeta.Name, err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": fmt.Sprintf("cannot replace service: %v", err),
		})
		return
	}

	redactServiceForUser(svcAfter)
	c.Status(200)
	c.Set("responseObject", svcAfter)
}

func DeleteService(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	err := kubecli.CoreV1().Services(nsname).Delete(objname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Services.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot delete service: %v", err),
		})
		return
	}
	c.Status(204)
}

func redactServiceForUser(svc *v1.Service) {
	svc.TypeMeta.Kind = "Service"
	svc.TypeMeta.APIVersion = "v1"
}

func redactServiceListForUser(svcl *v1.ServiceList) {
	for i := range svcl.Items {
		redactServiceForUser(&svcl.Items[i])
	}
}

func clientServiceInsertions(svc *v1.Service) {
}
