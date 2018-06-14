package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/containerum/cherry/adaptors/gonic"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	deploymentParam = "deployment"
	solutionParam   = "solution"
)

// swagger:operation GET /namespaces/{namespace}/deployments Deployment GetDeploymentList
// Get deployments list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: owner
//    in: query
//    type: string
//    required: false
// responses:
//  '200':
//    description: deployments list
//    schema:
//      $ref: '#/definitions/DeploymentsList'
//  default:
//    $ref: '#/responses/error'
func GetDeploymentList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Owner":     ctx.Query(ownerQuery),
	}).Debug("Get deployment list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	deploy, err := kube.GetDeploymentList(namespace, ctx.Query(ownerQuery))
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)

	ret, err := model.ParseKubeDeploymentList(deploy, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/solutions/{solution}/deployments Deployment GetDeploymentSolutionList
// Get solution deployments list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: solution
//    in: path
//    type: string
//    required: false
// responses:
//  '200':
//    description: deployments list
//    schema:
//      $ref: '#/definitions/DeploymentsList'
//  default:
//    $ref: '#/responses/error'
func GetDeploymentSolutionList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	solution := ctx.Param(solutionParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Solution":  solution,
	}).Debug("Get deployment list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	deploy, err := kube.GetDeploymentSolutionList(namespace, solution)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)

	ret, err := model.ParseKubeDeploymentList(deploy, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/deployments/{deployment} Deployment GetDeployment
// Get deployment.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: deployment
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: deployment
//    schema:
//      $ref: '#/definitions/Deployment'
//  default:
//    $ref: '#/responses/error'
func GetDeployment(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace":  namespace,
		"Deployment": deployment,
	}).Debug("Get deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	deploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deploy, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/deployments Deployment CreateDeployment
// Create deployment.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Deployment'
// responses:
//  '201':
//    description: deployment created
//    schema:
//      $ref: '#/definitions/Deployment'
//  default:
//    $ref: '#/responses/error'
func CreateDeployment(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq model.DeploymentKubeAPI
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	deploy, errs := deployReq.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}
	for _, v := range deploy.Spec.Template.Spec.Volumes {
		if v.PersistentVolumeClaim != nil {
			if pvc, err := kube.GetPersistentVolumeClaim(namespace, v.PersistentVolumeClaim.ClaimName); err != nil {
				gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
				return
			} else {
				if pvc.Status.Phase != "Bound" {
					gonic.Gonic(kubeErrors.ErrVolumeNotReady().AddDetailF("Volume status: %v", pvc.Status.Phase), ctx)
					return
				}
			}
		}
	}
	deployAfter, err := kube.CreateDeployment(deploy)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}
	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/deployments/{deployment} Deployment UpdateDeployment
// Update deployment.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: deployment
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Deployment'
// responses:
//  '202':
//    description: deployment updated
//    schema:
//      $ref: '#/definitions/Deployment'
//  default:
//    $ref: '#/responses/error'
func UpdateDeployment(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace":  namespace,
		"Deployment": deployment,
	}).Debug("Update deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var deployReq model.DeploymentKubeAPI
	if err := ctx.ShouldBindWith(&deployReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldDeploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	deployReq.Name = deployment
	deployReq.Owner = oldDeploy.GetObjectMeta().GetLabels()[ownerQuery]
	if oldDeploy.GetObjectMeta().GetLabels()["solution"] != "" {
		deployReq.SolutionID = oldDeploy.GetObjectMeta().GetLabels()["solution"]
	}

	deploy, errs := deployReq.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	//Ensure that immutable selectors wouldn't change
	deploy.Spec.Selector = oldDeploy.Spec.Selector
	deploy.Spec.Template.Labels = oldDeploy.Spec.Template.Labels

	deployAfter, err := kube.UpdateDeployment(deploy)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)

	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation PUT /namespaces/{namespace}/deployments/{deployment}/replicas Deployment UpdateDeploymentReplicas
// Update deployments replicas count.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: deployment
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UpdateReplicas'
// responses:
//  '202':
//    description: deployment updated
//    schema:
//      $ref: '#/definitions/Deployment'
//  default:
//    $ref: '#/responses/error'
func UpdateDeploymentReplicas(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace":  namespace,
		"Deployment": deployment,
	}).Debug("Update deployment replicas Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var replicas kube_types.UpdateReplicas
	if err := ctx.ShouldBindWith(&replicas, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	deploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	newRepl := int32(replicas.Replicas)
	deploy.Spec.Replicas = &newRepl

	deployAfter, err := kube.UpdateDeployment(deploy)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation PUT /namespaces/{namespace}/deployments/{deployment}/image Deployment UpdateDeploymentImage
// Update image in deployments container.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: deployment
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/UpdateImage'
// responses:
//  '202':
//    description: deployment updated
//    schema:
//      $ref: '#/definitions/Deployment'
//  default:
//    $ref: '#/responses/error'
func UpdateDeploymentImage(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace":  namespace,
		"Deployment": deployment,
	}).Debug("Update deployment container image Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var newImage kube_types.UpdateImage
	if err := ctx.ShouldBindWith(&newImage, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	deploy, err := kube.GetDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	deployUpd, err := model.UpdateImage(deploy, newImage.Container, newImage.Image)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableUpdateResource().AddDetailsErr(err), ctx)
		return
	}

	deployAfter, err := kube.UpdateDeployment(deployUpd)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeDeployment(deployAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/deployments/{deployment} Deployment DeleteDeployment
// Delete deployment.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: deployment
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: deployment deleted
//  default:
//    $ref: '#/responses/error'
func DeleteDeployment(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	deployment := ctx.Param(deploymentParam)
	log.WithFields(log.Fields{
		"Namespace":  namespace,
		"Deployment": deployment,
	}).Debug("Delete deployment Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteDeployment(namespace, deployment)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation DELETE /namespaces/{namespace}/solutiosn/{solution}deployments Deployment DeleteDeploymentsSolution
// Delete solution deployments.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: solution
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: deployments deleted
//  default:
//    $ref: '#/responses/error'
func DeleteDeploymentsSolution(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	solution := ctx.Param(solutionParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Solution":  solution,
	}).Debug("Delete solution deployments Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteDeploymentSolution(namespace, solution)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
