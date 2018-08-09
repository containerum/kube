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
	endpointParam = "endpoint"
)

// swagger:operation GET /namespaces/{namespace}/endpoints Endpoint GetEndpointList
// Get endpoints list.
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
// responses:
//  '200':
//    description: endpoints list
//    schema:
//      $ref: '#/definitions/EndpointsList'
//  default:
//    $ref: '#/responses/error'
func GetEndpointList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get endpoints list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	endpoints, err := kube.GetEndpointList(namespace)
	if err != nil {
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ret, err := model.ParseKubeEndpointList(endpoints)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/endpoints/{endpoint} Endpoint GetEndpoint
// Get endpoint.
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
//  - name: endpoint
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: endpoint
//    schema:
//      $ref: '#/definitions/Endpoint'
//  default:
//    $ref: '#/responses/error'
func GetEndpoint(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	ep := ctx.Param(endpointParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Endpoint":  ep,
	}).Debug("Get endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResource()), ctx)
		return
	}

	endpoint, err := kube.GetEndpoint(namespace, ep)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResource()), ctx)
		return
	}

	ret, err := model.ParseKubeEndpoint(endpoint)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/endpoints Endpoint CreateEndpoint
// Create endpoint.
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
//      $ref: '#/definitions/Endpoint'
// responses:
//  '201':
//    description: endpoint created
//    schema:
//      $ref: '#/definitions/Endpoint'
//  default:
//    $ref: '#/responses/error'
func CreateEndpoint(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var endpointReq model.Endpoint
	if err := ctx.ShouldBindWith(&endpointReq, binding.JSON); err != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableCreateResource()), ctx)
		return
	}

	newEndpoint, errs := endpointReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	endpointAfter, err := kube.CreateEndpoint(newEndpoint)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableCreateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/endpoints/{endpoint} Endpoint UpdateEndpoint
// Update endpoint.
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
//  - name: endpoint
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Endpoint'
// responses:
//  '201':
//    description: endpoint updated
//    schema:
//      $ref: '#/definitions/Endpoint'
//  default:
//    $ref: '#/responses/error'
func UpdateEndpoint(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	ep := ctx.Param(endpointParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Endpoint":  ep,
	}).Debug("Create endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var endpointReq model.Endpoint
	if err := ctx.ShouldBindWith(&endpointReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	_, err = kube.GetEndpoint(namespace, ep)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	endpointReq.Name = ep

	newEndpoint, errs := endpointReq.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	endpointAfter, err := kube.UpdateEndpoint(newEndpoint)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	ret, err := model.ParseKubeEndpoint(endpointAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/endpoints/{endpoint} Endpoint DeleteEndpoint
// Delete endpoint.
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
//  - name: endpoint
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: endpoint deleted
//  default:
//    $ref: '#/responses/error'
func DeleteEndpoint(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	ep := ctx.Param(endpointParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Endpoint":  ep,
	}).Debug("Delete endpoint Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteEndpoint(namespace, ep)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}
