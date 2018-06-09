package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	serviceParam = "service"
)

// swagger:operation GET /namespaces/{namespace}/services Service GetServiceList
// Get services list.
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
// responses:
//  '200':
//    description: services list
//    schema:
//      $ref: '#/definitions/ServicesList'
//  default:
//    $ref: '#/responses/error'
func GetServiceList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get service list call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	svcList, err := kube.GetServiceList(namespace)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeServiceList(svcList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/solutions/{solution}/services Service GetServiceSolutionList
// Get solution services list.
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
//    required: true
// responses:
//  '200':
//    description: services list
//    schema:
//      $ref: '#/definitions/ServicesList'
//  default:
//    $ref: '#/responses/error'
func GetServiceSolutionList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	solution := ctx.Param(solutionParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Solution":  solution,
	}).Debug("Get service list call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	svcList, err := kube.GetServiceSolutionList(namespace, solution)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeServiceList(svcList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/services/{service} Service GetService
// Get services list.
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
//  - name: service
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: service
//    schema:
//      $ref: '#/definitions/ServiceWithParam'
//  default:
//    $ref: '#/responses/error'
func GetService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	service := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   service,
	}).Debug("Get service call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	svc, err := kube.GetService(namespace, service)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeService(svc, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/services Service CreateService
// Create service.
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
//      $ref: '#/definitions/ServiceWithParam'
// responses:
//  '201':
//    description: service created
//    schema:
//      $ref: '#/definitions/ServiceWithParam'
//  default:
//    $ref: '#/responses/error'
func CreateService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create service Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithParam
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableCreateResource(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	newSvc, errs := svc.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	svcAfter, err := kube.CreateService(newSvc)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
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
//  - name: service
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/ServiceWithParam'
// responses:
//  '202':
//    description: service updated
//    schema:
//      $ref: '#/definitions/ServiceWithParam'
//  default:
//    $ref: '#/responses/error'
func UpdateService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	service := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   service,
	}).Debug("Update service Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var svc model.ServiceWithParam
	if err := ctx.ShouldBindWith(&svc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableUpdateResource(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldSvc, err := kube.GetService(namespace, service)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	svc.Name = ctx.Param(serviceParam)
	svc.Owner = oldSvc.GetObjectMeta().GetLabels()[ownerQuery]
	if oldSvc.GetObjectMeta().GetLabels()["solution"] != "" {
		svc.SolutionID = oldSvc.GetObjectMeta().GetLabels()["solution"]
	}

	newSvc, errs := svc.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	newSvc.ResourceVersion = oldSvc.ResourceVersion
	newSvc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
	newSvc.Labels = oldSvc.Labels
	newSvc.Spec.Selector = oldSvc.Spec.Selector

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
//  - name: service
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: service deleted
//  default:
//    $ref: '#/responses/error'
func DeleteService(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	service := ctx.Param(serviceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Service":   service,
	}).Debug("Delete service call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteService(namespace, service)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

// swagger:operation DELETE /namespaces/{namespace}/solutions/{solution}/services Service DeleteServicesSolution
// Delete solution services.
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
//    required: true
// responses:
//  '202':
//    description: services deleted
//  default:
//    $ref: '#/responses/error'
func DeleteServicesSolution(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	solution := ctx.Param(solutionParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Solution":  solution,
	}).Debug("Delete solution services call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	list, err := kube.GetServiceSolutionList(namespace, solution)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	for _, s := range list.Items {
		err = kube.DeleteService(namespace, s.Name)
		if err != nil {
			log.WithError(err)
		}
	}

	ctx.Status(http.StatusAccepted)
}
