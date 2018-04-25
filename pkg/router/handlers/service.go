package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	serviceParam = "service"
)

// swagger:operation GET /namespaces/{namespace}/services Service GetServiceList
// Get services list.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-service-list
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
//    description: services list
//    schema:
//      $ref: '#/definitions/ServicesList'
//  default:
//    description: error
func GetServiceList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get service list call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	svcList, err := kube.GetServiceList(namespace)
	if err != nil {
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeServiceList(svcList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/services/{service} Service GetService
// Get services list.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-service
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
//  - name: service
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: secrets list
//    schema:
//      $ref: '#/definitions/ServiceWithOwner'
//  default:
//    description: error
func GetService(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	service := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Service":         service,
	}).Debug("Get service call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	svc, err := kube.GetService(namespace, service)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeService(svc, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/services Service CreateService
// Create service.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#create-service
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
//      $ref: '#/definitions/ServiceWithOwner'
// responses:
//  '201':
//    description: service created
//    schema:
//      $ref: '#/definitions/ServiceWithOwner'
//  default:
//    description: error
func CreateService(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create service Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithOwner
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableCreateResource(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	newSvc, errs := svc.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	svcAfter, err := kube.CreateService(newSvc)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeService(svcAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/services/{service} Service UpdateService
// Update service.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#update-service
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
//  - name: service
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/ServiceWithOwner'
// responses:
//  '202':
//    description: service updated
//    schema:
//      $ref: '#/definitions/ServiceWithOwner'
//  default:
//    description: error
func UpdateService(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	service := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Service":         service,
	}).Debug("Update service Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithOwner
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableUpdateResource(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	svc.Name = ctx.Param(serviceParam)

	oldSvc, err := kube.GetService(namespace, service)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableUpdateResource()), ctx)
		return
	}

	svc.Name = ctx.Param(serviceParam)
	svc.Owner = oldSvc.GetObjectMeta().GetLabels()[ownerQuery]

	newSvc, errs := svc.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	newSvc.ResourceVersion = oldSvc.ResourceVersion
	newSvc.Spec.ClusterIP = oldSvc.Spec.ClusterIP

	updatedService, err := kube.UpdateService(newSvc)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeService(updatedService, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/services/{service} Service DeleteService
// Delete service.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#delete-service
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
//  - name: service
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: service deleted
//  default:
//    description: error
func DeleteService(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	service := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Service":         service,
	}).Debug("Delete service call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteService(namespace, service)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
