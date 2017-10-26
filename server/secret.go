package server

import (
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
)

func ListSecrets(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	secretList, err := kubecli.CoreV1().Secrets(nsname).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Errorf("kubecli.Secrets.List error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": "cannot list secrets: " + err.Error(),
		})
		return
	}

	redactSecretListForUser(secretList)
	c.Status(200)
	c.Set("responseObject", secretList)
}

func CreateSecret(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	secret, ok := c.MustGet("requestObject").(*v1.Secret)
	if !ok || secret == nil {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "bad request",
		})
		return
	}
	if nsname != secret.ObjectMeta.Namespace {
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in secret",
		})
		return
	}
	clientSecretInsertions(secret)

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	secretAfter, err := kubecli.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(secret)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Secrets.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create secret: %v", err),
		})
		return
	}

	redactSecretForUser(secretAfter)

	c.Status(201)
	c.Set("responseObject", secretAfter)
}

func GetSecret(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	objname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	secret, err := kubecli.CoreV1().Secrets(nsname).Get(objname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Secrets.Get error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get secret %s: %v", objname, err),
		})
		return
	}
	redactSecretForUser(secret)
	c.Status(200)
	c.Set("responseObject", secret)
}

func DeleteSecret(c *gin.Context) {
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

func redactSecretForUser(secret *v1.Secret) {
	secret.TypeMeta.Kind = "Secret"
	secret.TypeMeta.APIVersion = "v1"
}

func redactSecretListForUser(secretList *v1.SecretList) {
	for i := range secretList.Items {
		redactSecretForUser(&secretList.Items[i])
	}
}

func clientSecretInsertions(secret *v1.Secret) {
}
