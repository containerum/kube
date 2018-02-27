package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	deploymentParam = "deployment"
)

func GetDeploymentList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Owner":           ctx.Query(ownerQuery),
	}).Debug("Get deployment list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployments, err := kube.GetDeploymentList(ctx.MustGet(m.NamespaceKey).(string), ctx.Query(ownerQuery))
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ret, err := model.ParseDeploymentList(deployments)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Deployment":      ctx.Param(deploymentParam),
	}).Debug("Get deployment Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableGetResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseDeployment(deployment)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResource().Gonic(ctx)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

func CreateDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq model.DeploymentWithOwner
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	deployment, errs := model.MakeDeployment(ctx.Param(namespaceParam), deployReq, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	deployAfter, err := kube.CreateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
	}
	ctx.JSON(http.StatusCreated, ret)
}

func UpdateDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Update deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq model.DeploymentWithOwner
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	oldDeploy, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	deployReq.Name = ctx.Param(deploymentParam)
	deployReq.Owner = oldDeploy.GetObjectMeta().GetLabels()["owner"]

	deployment, errs := model.MakeDeployment(ctx.Param(namespaceParam), deployReq, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func UpdateDeploymentReplicas(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Update deployment replicas Call")
	var replicas kube_types.UpdateReplicas
	if err := ctx.ShouldBindWith(&replicas, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))

	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	newRepl := int32(replicas.Replicas)
	deployment.Spec.Replicas = &newRepl

	deployAfter, err := kube.UpdateDeployment(ctx.Param(namespaceParam), deployment)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func UpdateDeploymentImage(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Update deployment container image Call")
	var image kube_types.UpdateImage
	if err := ctx.ShouldBindWith(&image, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	deployment, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	deploymentUpd, err := model.UpdateImage(deployment, image.Container, image.Image)
	if err != nil {
		cherry.ErrUnableUpdateResource().AddDetailsErr(err).Gonic(ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(ctx.Param(namespaceParam), deploymentUpd)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseDeployment(deployAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteDeployment(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":  ctx.Param(namespaceParam),
		"Deployment": ctx.Param(deploymentParam),
	}).Debug("Delete deployment Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableDeleteResource()).Gonic(ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
