package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
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
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Owner":           ctx.Query(ownerQuery),
	}).Debug("Get deployment list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	deploy, err := kube.GetDeploymentList(namespace, ctx.Query(ownerQuery))
	if err != nil {
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)

	ret, err := model.ParseKubeDeploymentList(deploy, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetDeployment(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Deployment":      deployment,
	}).Debug("Get deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	deploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deploy, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateDeployment(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq model.DeploymentWithOwner
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		deployReq.Owner = ctx.MustGet(m.UserID).(string)
	}

	deploy, errs := deployReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	deployAfter, err := kube.CreateDeployment(deploy)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}
	ctx.JSON(http.StatusCreated, ret)
}

func UpdateDeployment(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Deployment":      deployment,
	}).Debug("Update deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq model.DeploymentWithOwner
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	oldDeploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	deployReq.Name = deployment
	deployReq.Owner = oldDeploy.GetObjectMeta().GetLabels()[ownerQuery]

	deploy, errs := deployReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(deploy)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)

	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func UpdateDeploymentReplicas(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Deployment":      deployment,
	}).Debug("Update deployment replicas Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var replicas kube_types.UpdateReplicas
	if err := ctx.ShouldBindWith(&replicas, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	deploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	newRepl := int32(replicas.Replicas)
	deploy.Spec.Replicas = &newRepl

	deployAfter, err := kube.UpdateDeployment(deploy)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func UpdateDeploymentImage(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Deployment":      deployment,
	}).Debug("Update deployment container image Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var newImage kube_types.UpdateImage
	if err := ctx.ShouldBindWith(&newImage, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	deploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	deployUpd, err := model.UpdateImage(deploy, newImage.Container, newImage.Image)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableUpdateResource().AddDetailsErr(err), ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(deployUpd)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteDeployment(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Deployment":      deployment,
	}).Debug("Delete deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
