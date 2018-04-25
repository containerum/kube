package handlers

import (
	"net/http"

	"git.containerum.net/ch/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	ingressParam = "ingress"
)

// swagger:operation GET /namespaces/{namespace}/ingresses Ingress GetIngressList
// Get ingresses list.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-ingress-list
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: ingresses list
//    schema:
//      $ref: '#/definitions/IngressesList'
//  configmap:
//    description: error
func GetIngressList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get ingress list")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingressList, err := kube.GetIngressList(namespace)
	if err != nil {
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeIngressList(ingressList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/ingresses/{ingress} Ingress GetIngress
// Get ingresses list.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-ingress
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: ingress
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: ingresses
//    schema:
//      $ref: '#/definitions/IngressWithOwner'
//  configmap:
//    description: error
func GetIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Ingress":         ingr,
	}).Debug("Get ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingress, err := kube.GetIngress(namespace, ingr)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeIngressList(ingress, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/ingresses Ingress CreateIngress
// Create ingress.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#create-ingress
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/IngressWithOwner'
// responses:
//  '201':
//    description: ingress created
//    schema:
//      $ref: '#/definitions/IngressWithOwner'
//  default:
//    description: error
func CreateIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingressReq model.IngressWithOwner
	if err := ctx.ShouldBindWith(&ingressReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		ingressReq.Owner = ctx.MustGet(m.UserID).(string)
	}

	newIngress, errs := ingressReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kube.CreateIngress(newIngress)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeIngress(ingressAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/ingresses/{ingress} Ingress UpdateIngress
// Update ingress.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#update-ingress
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: ingress
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/IngressWithOwner'
// responses:
//  '201':
//    description: ingress updated
//    schema:
//      $ref: '#/definitions/IngressWithOwner'
//  default:
//    description: error
func UpdateIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Ingress":         ingr,
	}).Debug("Update ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingressReq model.IngressWithOwner
	if err := ctx.ShouldBindWith(&ingressReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	oldIngress, err := kube.GetIngress(namespace, ingr)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	ingressReq.Name = ingr
	ingressReq.Owner = oldIngress.GetObjectMeta().GetLabels()[ownerQuery]

	newIngress, errs := ingressReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kube.UpdateIngress(newIngress)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeIngress(ingressAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/ingresses/{ingress} Ingress DeleteIngress
// Delete ingress.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#delete-ingress
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: ingress
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: ingress deleted
//  default:
//    description: error
func DeleteIngress(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Ingress":         ingr,
	}).Debug("Delete ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteIngress(namespace, ingr)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation GET /ingresses Ingress GetSelectedIngresses
// Get ingresses from all user namespaces.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
// responses:
//  '200':
//    description: ingresses list from all users namespaces
//    schema:
//      $ref: '#/definitions/SelectedIngressesList'
//  default:
//    description: error
func GetSelectedIngresses(ctx *gin.Context) {
	log.Debug("Get selected ingresses Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingresses := make(model.SelectedIngressesList, 0)

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {

			ingressList, err := kube.GetIngressList(n.ID)
			if err != nil {
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			ingressesList, err := model.ParseKubeIngressList(ingressList, role == m.RoleUser)
			if err != nil {
				ctx.Error(err)
				gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
				return
			}

			ingresses[n.Label] = *ingressesList
		}
	}

	ctx.JSON(http.StatusOK, ingresses)
}
