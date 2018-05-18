package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin/binding"
)

const (
	ownerQuery     = "owner"
	namespaceParam = "namespace"
)

// swagger:operation GET /namespaces Namespace GetNamespaceList
// Get namespaces list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: owner
//    in: query
//    type: string
//    required: false
// responses:
//  '200':
//    description: ingresses list
//    schema:
//      $ref: '#/definitions/NamespacesList'
//  default:
//    $ref: '#/responses/error'
func GetNamespaceList(ctx *gin.Context) {
	owner := ctx.Query(ownerQuery)

	role := ctx.MustGet(m.UserRole).(string)
	//TODO: Show only namespaces with owner = X-User-Id
	/*if role == m.RoleUser {
		owner = ctx.MustGet(m.UserID).(string)
	}*/

	log.WithField("Owner", owner).Debug("Get namespace list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	quotas, err := kube.GetNamespaceQuotaList(owner)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ret, err := model.ParseKubeResourceQuotaList(quotas, role == m.RoleAdmin)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		ret = model.ParseNamespaceListForUser(*nsList, ret.Namespaces)
	}
	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace} Namespace GetNamespace
// Get namespace.
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
//    description: namespace
//    schema:
//      $ref: '#/definitions/NamespaceWithOwner'
//  default:
//    $ref: '#/responses/error'
func GetNamespace(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeResourceQuota(quota, role == m.RoleAdmin)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResource(), ctx)
		return
	}

	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		ret.ParseForUser(*nsList)
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces Namespace CreateNamespace
// Create namespace.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/NamespaceWithOwner'
// responses:
//  '201':
//    description: namespace created
//    schema:
//      $ref: '#/definitions/NamespaceWithOwner'
//  default:
//    $ref: '#/responses/error'
func CreateNamespace(ctx *gin.Context) {
	log.Debug("Create namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ns model.NamespaceWithOwner
	if err := ctx.ShouldBindWith(&ns, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	newNs, errs := ns.ToKube()
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	newQuota, errs := model.MakeResourceQuota(ns.Name, newNs.Labels, ns.Resources.Hard)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	_, err := kube.CreateNamespace(newNs)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	quotaCreated, err := kube.CreateNamespaceQuota(ns.Name, newQuota)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeResourceQuota(quotaCreated, role == m.RoleAdmin)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace} Namespace UpdateNamespace
// Update namespace.
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
//      $ref: '#/definitions/NamespaceWithOwner'
// responses:
//  '201':
//    description: namespace updated
//    schema:
//      $ref: '#/definitions/NamespaceWithOwner'
//  default:
//    $ref: '#/responses/error'
func UpdateNamespace(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Update namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var res model.NamespaceWithOwner
	if err := ctx.ShouldBindWith(&res, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	quotaOld, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	quota, errs := model.MakeResourceQuota(quotaOld.Namespace, quotaOld.Labels, res.Resources.Hard)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	quotaAfter, err := kube.UpdateNamespaceQuota(namespace, quota)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeResourceQuota(quotaAfter, role == m.RoleAdmin)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)

}

// swagger:operation DELETE /namespaces/{namespace} Namespace DeleteNamespace
// Delete namespace.
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
// responses:
//  '202':
//    description: namespace deleted
//  default:
//    $ref: '#/responses/error'
func DeleteNamespace(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Delete namespace Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
