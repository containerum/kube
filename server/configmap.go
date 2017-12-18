package server

import (
	"fmt"

	"git.containerum.net/ch/kube-api/utils"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
)

func ListConfigMaps(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	cmList, err := kubecli.CoreV1().ConfigMaps(nsname).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.ConfigMap.List error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": "cannot list configmaps: " + err.Error(),
		})
		return
	}

	redactConfigMapListForUser(cmList)
	c.JSON(200, cmList)
}

func CreateConfigMap(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	cm, ok := c.MustGet("requestObject").(*v1.ConfigMap)
	if !ok || cm == nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "bad request",
		})
		return
	}
	if nsname != cm.ObjectMeta.Namespace {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in endpoints",
		})
		return
	}
	clientConfigMapInsertions(cm)

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	cmAfter, err := kubecli.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(cm)
	if err != nil {
		utils.Log(c).Warnf("kubecli.ConfigMap.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create configmap: %v", err),
		})
		return
	}

	redactConfigMapForUser(cmAfter)

	c.JSON(201, cmAfter)
}

func GetConfigMap(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	cm, err := kubecli.CoreV1().ConfigMaps(nsname).Get(objname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.ConfigMap.Get error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get configmap %s: %v", objname, err),
		})
		return
	}
	redactConfigMapForUser(cm)
	c.JSON(200, cm)
}

func DeleteConfigMap(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	err := kubecli.CoreV1().ConfigMaps(nsname).Delete(objname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.ConfigMap.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot delete endpoints %s: %v", objname, err),
		})
		return
	}
	c.Status(204)
}

func redactConfigMapForUser(cm *v1.ConfigMap) {
	cm.TypeMeta.Kind = "ConfigMap"
	cm.TypeMeta.APIVersion = "v1"
}

func redactConfigMapListForUser(cmlist *v1.ConfigMapList) {
	for i := range cmlist.Items {
		redactConfigMapForUser(&cmlist.Items[i])
	}
}

func clientConfigMapInsertions(cm *v1.ConfigMap) {
}
