package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	json_types "git.containerum.net/ch/kube-client/pkg/model"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	deploymentParam = "deployment"
)

func getDeploymentList(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Owner":     c.Query(ownerQuery),
	}).Debug("Get deployment list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployments, err := kube.GetDeploymentList(c.Param(namespaceParam), c.Query(ownerQuery))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusOK, model.ParseDeploymentList(deployments))
}

func getDeployment(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  c.Param(namespaceParam),
		"Deployment": c.Param(deploymentParam),
	}).Debug("Get deployment Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(c.Param(namespaceParam), c.Param(deploymentParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.JSON(http.StatusOK, model.ParseDeployment(deployment))
}

func createDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create deployment Call")

	var depl json_types.Deployment
	if err := ctx.ShouldBindWith(&depl, binding.JSON); err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	contaiers, err := model.MakeContainers(*depl.Containers)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newDepl, err := model.MakeDeployment(ctx.Param(namespaceParam), &depl, contaiers)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	deplAfter, err := kubecli.CreateDeployment(ctx.Param(namespaceParam), newDepl)
	if err != nil {
		ctx.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	ctx.JSON(http.StatusCreated, model.ParseDeployment(deplAfter))
}

func deleteDeployment(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  c.Param(namespaceParam),
		"Deployment": c.Param(deploymentParam),
	}).Debug("Delete deployment Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteDeployment(c.Param(namespaceParam), c.Param(deploymentParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	c.Status(http.StatusAccepted)
}

func updateDeployment(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  c.Param(namespaceParam),
		"Deployment": c.Param(deploymentParam),
	}).Debug("Update deployment Call")

	var depl json_types.Deployment
	if err := c.ShouldBindWith(&depl, binding.JSON); err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	if c.Param(deploymentParam) != depl.Name {
		log.Errorf(invalidUpdateDeploymentName, c.Param(deploymentParam), depl.Name)
		c.AbortWithStatusJSON(ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateDeploymentName, c.Param(deploymentParam), depl.Name), http.StatusBadRequest)))
		return
	}

	kubecli := c.MustGet(m.KubeClient).(*kubernetes.Kube)

	contaiers, err := model.MakeContainers(*depl.Containers)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	newDepl, err := model.MakeDeployment(c.Param(namespaceParam), &depl, contaiers)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	deplAfter, err := kubecli.UpdateDeployment(c.Param(namespaceParam), newDepl)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusAccepted, model.ParseDeployment(deplAfter))
}

func updateDeploymentReplicas(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  c.Param(namespaceParam),
		"Deployment": c.Param(deploymentParam),
	}).Debug("Update deployment replicas Call")
	var replicas json_types.UpdateReplicas
	if err := c.ShouldBindWith(&replicas, binding.JSON); err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(c.Param(namespaceParam), c.Param(deploymentParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}
	newRepl := int32(replicas.Replicas)
	deployment.Spec.Replicas = &newRepl

	deplAfter, err := kube.UpdateDeployment(c.Param(namespaceParam), deployment)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusAccepted, model.ParseDeployment(deplAfter))
}

func updateDeploymentImage(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  c.Param(namespaceParam),
		"Deployment": c.Param(deploymentParam),
	}).Debug("Update deployment container image Call")
	var image json_types.UpdateImage
	if err := c.ShouldBindWith(&image, binding.JSON); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(c.Param(namespaceParam), c.Param(deploymentParam))
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	updated := false

	for i, v := range deployment.Spec.Template.Spec.Containers {
		if v.Name == image.ContainerName {
			deployment.Spec.Template.Spec.Containers[i].Image = image.Image
			updated = true
			break
		}
	}

	if updated == false {
		c.AbortWithStatusJSON(ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(containerNotFoundError, c.Param(namespaceParam), c.Param(deploymentParam)), http.StatusNotFound)))
		return
	}

	deplAfter, err := kube.UpdateDeployment(c.Param(namespaceParam), deployment)
	if err != nil {
		c.AbortWithStatusJSON(ParseErorrs(err))
		return
	}

	c.JSON(http.StatusAccepted, model.ParseDeployment(deplAfter))
}
