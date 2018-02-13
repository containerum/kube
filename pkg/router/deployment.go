package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	deploymentParam = "deployment"
)

func getDeploymentList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Owner":           ctx.Query(ownerQuery),
	}).Debug("Get deployment list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployments, err := kube.GetDeploymentList(ctx.MustGet(m.NamespaceKey).(string), ctx.Query(ownerQuery))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseDeploymentList(deployments)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func getDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Deployment":      ctx.Param(deploymentParam),
	}).Debug("Get deployment Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseDeployment(deployment)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

func createDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create deployment Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq kube_types.Deployment
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	deployment, err := model.MakeDeployment(ctx.Param(namespaceParam), &deployReq, quota.Labels)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	deployAfter, err := kubecli.CreateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.JSON(http.StatusCreated, ret)
}

func deleteDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Delete deployment Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}

func updateDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Update deployment Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq kube_types.Deployment
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	if ctx.Param(deploymentParam) != deployReq.Name {
		log.Errorf(invalidUpdateDeploymentName, ctx.Param(deploymentParam), deployReq.Name)
		ctx.Error(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateDeploymentName, ctx.Param(deploymentParam), deployReq.Name), http.StatusBadRequest))
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(invalidUpdateDeploymentName, ctx.Param(deploymentParam), deployReq.Name), http.StatusBadRequest)))
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	deployment, err := model.MakeDeployment(ctx.Param(namespaceParam), &deployReq, quota.Labels)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	deployAfter, err := kubecli.UpdateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func updateDeploymentReplicas(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Update deployment replicas Call")
	var replicas kube_types.UpdateReplicas
	if err := ctx.ShouldBindWith(&replicas, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}
	newRepl := int32(replicas.Replicas)
	deployment.Spec.Replicas = &newRepl

	deployAfter, err := kube.UpdateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func updateDeploymentImage(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Update deployment container image Call")
	var image kube_types.UpdateImage
	if err := ctx.ShouldBindWith(&image, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	updated := false
	for i, v := range deployment.Spec.Template.Spec.Containers {
		if v.Name == image.Container {
			deployment.Spec.Template.Spec.Containers[i].Image = image.Image
			updated = true
			break
		}
	}
	if updated == false {
		ctx.Error(model.NewErrorWithCode(fmt.Sprintf(containerNotFoundError, ctx.Param(namespaceParam), ctx.Param(deploymentParam)), http.StatusNotFound))
		ctx.AbortWithStatusJSON(model.ParseErorrs(model.NewErrorWithCode(fmt.Sprintf(containerNotFoundError, ctx.Param(namespaceParam), ctx.Param(deploymentParam)), http.StatusNotFound)))
		return
	}

	deployAfter, err := kube.UpdateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(model.ParseErorrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, ret)
}
