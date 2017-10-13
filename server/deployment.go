package server

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GET /api/v1/namespaces/:namespace/deployments
//
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

// POST /api/v1/namespaces/:namespace/deployments
//
// middleware deps:
// 	SetNamespace
// 	Set(…)KubeClient
// 	ParseJSON
func CreateDeployment(c *gin.Context) {
	var err error
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
	incomingDeploymentMod(depl)

	err = deploymentSanityCheck(depl)
	if err != nil {
		utils.Log(c).Warnf("deploymentSanityCheck: %v", err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": fmt.Sprintf("bad input: %v", err),
		})
		return
	}

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

// GET /api/v1/namespaces/:namespace/deployments/:objname
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

// DELETE /api/v1/namespaces/:namespace/deployments/:objname
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
		return
	}

	c.Status(204)
}

// PUT /api/v1/namespaces/:namespace/deployments/:objname
func UpdateDeployment(c *gin.Context) {
	var err error
	ns := c.MustGet("namespace").(string)
	deplname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)

	depl, ok := c.MustGet("requestObject").(*v1beta1.Deployment)
	if !ok {
		utils.Log(c).Warnf("invalid input: type %T value %[1]v", depl)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "invalid input",
		})
		return
	}
	if ns != depl.ObjectMeta.Namespace {
		utils.Log(c).Warnf("namespace in URI (%s) does not match namespace in deployment (%s)", ns, depl.ObjectMeta.Namespace)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "namespace in URI does not match namespace in deployment",
		})
		return
	}

	incomingDeploymentMod(depl)
	err = deploymentSanityCheck(depl)
	if err != nil {
		utils.Log(c).Warnf("deployment sanity check fail: %v", err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": fmt.Sprintf("invalid input: %v", err),
		})
		return
	}

	deplAfter, err := kubecli.AppsV1beta1().Deployments(ns).Update(depl)
	if err != nil {
		utils.Log(c).Warnf("cannot update deployment", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot update deployment %s: %v", deplname, err),
		})
		return
	}

	redactDeploymentForUser(deplAfter)
	c.Status(200)
	c.Set("responseObject", deplAfter)
}

// PATCH /api/v1/namespaces/:namespace/deployments/:objname
func PatchDeployment(c *gin.Context) {
	c.AbortWithStatusJSON(501, map[string]string{
		"error": "not implemented",
	})
}

// Only invoke after GetDeployment.
//
// Example
//
//   {"name":"ngx","image":"nginx:1.10.3"}
//
// PATCH /api/v1/namespaces/:namespace/deployments/:objname/image
func ChangeDeploymentImage(c *gin.Context) {
	var chimg struct {
		Name  string // selector for container name
		Image string // new image name
	}
	var err error
	ns := c.MustGet("namespace").(string)
	deplname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)

	depl := c.MustGet("responseObject").(*v1beta1.Deployment)
	jsn := c.MustGet("requestObject").(json.RawMessage)
	json.Unmarshal(jsn, &chimg)

	for i := range depl.Spec.Template.Spec.Containers {
		if depl.Spec.Template.Spec.Containers[i].Name == chimg.Name {
			utils.Log(c).Infof("changing image in deployment %q container %q from %q to %q",
				deplname, chimg.Name, depl.Spec.Template.Spec.Containers[i].Image, chimg.Image)
			depl.Spec.Template.Spec.Containers[i].Image = chimg.Image
		}
	}

	err = deploymentSanityCheck(depl)
	if err != nil {
		utils.Log(c).Warnf("deploymentSanityCheck: %v", err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": fmt.Sprintf("bad input: %v", err),
		})
		return
	}

	deplAfter, err := kubecli.AppsV1beta1().Deployments(ns).Update(depl)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Update error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot update deployment %s: %v", deplname, err),
		})
		return
	}
	redactDeploymentForUser(deplAfter)
	c.Status(200)
	c.Set("responseObject", deplAfter)
}

// Only invoke after GetDeployment.
//
// Example
//
//   {"replicas": 3}
//
// PUT /api/v1/namespaces/:namespace/deployments/:objname/replicas
func ChangeDeploymentReplicas(c *gin.Context) {
	var replicas struct {
		Replicas *int32 `json:"replicas,omitempty"`
	}
	var err error
	ns := c.MustGet("namespace").(string)
	deplname := c.MustGet("objectName").(string)
	kubecli := c.MustGet("kubeclient").(*kubernetes.Clientset)

	depl := c.MustGet("responseObject").(*v1beta1.Deployment)
	jsn := c.MustGet("requestObject").(json.RawMessage)
	json.Unmarshal(jsn, &replicas)

	if replicas.Replicas == nil || *replicas.Replicas < 0 || *replicas.Replicas >= 20 {
		utils.Log(c).WithField("function", "ChangeDeploymentReplicas").
			Warnf("invalid replicas: %v %+[1]v", replicas.Replicas)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": "invalid input",
		})
		return
	}
	depl.Spec.Replicas = replicas.Replicas

	err = deploymentSanityCheck(depl)
	if err != nil {
		utils.Log(c).Warnf("deploymentSanityCheck: %v", err)
		c.AbortWithStatusJSON(400, map[string]string{
			"error": fmt.Sprintf("bad input: %v", err),
		})
		return
	}

	deplAfter, err := kubecli.AppsV1beta1().Deployments(ns).Update(depl)
	if err != nil {
		utils.Log(c).Warnf("kubecli.Deployments.Update error: %T %[1]v", err)
		c.AbortWithStatusJSON(utils.KubeErrorHTTPStatus(err), map[string]string{
			"error": fmt.Sprintf("cannot update deployment %s: %v", deplname, err),
		})
		return
	}
	redactDeploymentForUser(deplAfter)
	c.Status(200)
	c.Set("responseObject", deplAfter)
}

func redactDeploymentForUser(depl *v1beta1.Deployment) {
	depl.TypeMeta.Kind = "Deployment"
	depl.TypeMeta.APIVersion = "apps/v1beta1"
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

func incomingDeploymentMod(depl *v1beta1.Deployment) {
	depl.TypeMeta.APIVersion = "apps/v1beta1"
	depl.Status = v1beta1.DeploymentStatus{}
	depl.Spec.Template.Spec.NodeSelector = map[string]string{
		"role": "slave",
	}
	for i := range depl.Spec.Template.Spec.Containers {
		depl.Spec.Template.Spec.Containers[i].Resources.Limits = depl.Spec.Template.Spec.Containers[i].Resources.Requests
	}
}

// deploymentSanityCheck checks, e.g. that replicas is within [1; 20].
func deploymentSanityCheck(depl *v1beta1.Deployment) (err error) {
	if depl.Spec.Replicas != nil && (*depl.Spec.Replicas < 1 || *depl.Spec.Replicas > 20) {
		return fmt.Errorf("invalid replicas (1 <= %d <= 20)", *depl.Spec.Replicas)
	}

	return
}
