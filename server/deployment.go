package server

import (
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// middleware deps:
// 	SetNamespace
// 	Set(…)KubeClient
func ListDeployments(c *gin.Context) {
	ns := c.MustGet("namespace").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	deplList, err := kubecli.AppsV1beta1().Deployments(ns).List(meta_v1.ListOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.List error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot list deployments: %v", err),
		})
		return
	}

	redactDeploymentListForUser(deplList)

	c.Status(200)
	c.Set("responseObject", deplList)
}

// middleware deps:
// 	SetNamespace
// 	Set(…)KubeClient
// 	ParseJSON
func CreateDeployment(c *gin.Context) {
	nsname := c.MustGet("namespace").(string)
	depl, ok := c.MustGet("requestObject").(*v1beta1.Deployment)
	if !ok || depl == nil {
		c.AbortWithStatusJSON(400, map[string]string{"error": "bad request"})
		return
	}
	if nsname != depl.ObjectMeta.Namespace {
		utils.Log(c).Warnf("namespace in URI (%s) does not match namespace in deployment (%s)", nsname, depl.ObjectMeta.Namespace)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in deployment",
		})
		return
	}

	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	clientDeploymentInsertions(depl)
	deplAfter, err := kubecli.AppsV1beta1().Deployments(depl.ObjectMeta.Namespace).Create(depl)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Create error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot create deployment: %v", err),
		})
		return
	}

	redactDeploymentForUser(deplAfter)

	c.Status(201)
	c.Set("responseObject", deplAfter)
}

func GetDeployment(c *gin.Context) {
	ns := c.MustGet("namespace").(string)
	deplname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)

	depl, err := kubecli.AppsV1beta1().Deployments(ns).Get(deplname, meta_v1.GetOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Get error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot get deployment %s: %v", deplname, err),
		})
		return
	}
	redactDeploymentForUser(depl)
	c.Status(200)
	c.Set("responseObject", depl)
}

func DeleteDeployment(c *gin.Context) {
	ns := c.MustGet("namespace").(string)
	deplname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)
	err := kubecli.AppsV1beta1().Deployments(ns).Delete(deplname, &meta_v1.DeleteOptions{})
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Delete error: %[1]T %[1]v", err)
		c.AbortWithStatusJSON(503, map[string]string{
			"error": fmt.Sprintf("cannot delete deployments: %v", err),
		})
	}
	c.Status(204)
}

func redactDeploymentForUser(depl *v1beta1.Deployment) {
	depl.ObjectMeta.SelfLink = ""
	depl.ObjectMeta.UID = ""
	for i := range depl.Spec.Template.Spec.Volumes {
		depl.Spec.Template.Spec.Volumes[i].VolumeSource = v1.VolumeSource{}
	}
	depl.Spec.Template.Spec.NodeSelector = nil
}

func redactDeploymentListForUser(deplList *v1beta1.DeploymentList) {
	for i := range deplList.Items {
		redactDeploymentForUser(&deplList.Items[i])
	}
	deplList.ListMeta.SelfLink = ""
	deplList.ListMeta.ResourceVersion = ""
}

func clientDeploymentInsertions(depl *v1beta1.Deployment) {
	depl.Spec.Template.Spec.NodeSelector = map[string]string{
		"role": "slave",
	}
	for i := range depl.Spec.Template.Spec.Containers {
		depl.Spec.Template.Spec.Containers[i].Resources.Limits = depl.Spec.Template.Spec.Containers[i].Resources.Requests
	}
}
