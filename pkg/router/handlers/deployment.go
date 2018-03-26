package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_apps "k8s.io/api/apps/v1"

	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
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
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseDeploymentList(deployments, role == "user")
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
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
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseDeployment(deployment, role == "user")
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
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
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	deployment, errs := model.MakeDeployment(ctx.Param(namespaceParam), deployReq, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	deployAfter, err := kube.CreateDeployment(deployment)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseDeployment(deployAfter, role == "user")
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
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	oldDeploy, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	deployReq.Name = ctx.Param(deploymentParam)
	deployReq.Owner = oldDeploy.GetObjectMeta().GetLabels()["owner"]

	deployment, errs := model.MakeDeployment(ctx.Param(namespaceParam), deployReq, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(deployment)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseDeployment(deployAfter, role == "user")
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
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	deployment, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	newRepl := int32(replicas.Replicas)
	deployment.Spec.Replicas = &newRepl

	deployAfter, err := kube.UpdateDeployment(deployment)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseDeployment(deployAfter, role == "user")
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
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	deployment, err := kube.GetDeployment(ctx.Param(namespaceParam), ctx.Param(deploymentParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	deploymentUpd, err := model.UpdateImage(deployment, image.Container, image.Image)
	if err != nil {
		gonic.Gonic(cherry.ErrUnableUpdateResource().AddDetailsErr(err), ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(deploymentUpd)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseDeployment(deployAfter, role == "user")
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
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

func CreateDeploymentFromFile(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Create deployment from file Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deploy api_apps.Deployment
	if err := ctx.ShouldBindWith(&deploy, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == "user" {
		deploy.Namespace = ctx.MustGet(m.NamespaceKey).(string)
	} else {
		deploy.Namespace = ctx.Param(namespaceParam)
	}

	_, err := kube.GetNamespaceQuota(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailF(noNamespace, ctx.Param(namespaceParam)), ctx)
		return
	}

	deployAfter, err := kube.CreateDeployment(&deploy)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableCreateResource()).AddDetailsErr(err), ctx)
		return
	}

	ret, err := model.ParseDeployment(deployAfter, role == "user")
	if err != nil {
		ctx.Error(err)
	}
	ctx.JSON(http.StatusCreated, ret)
}
