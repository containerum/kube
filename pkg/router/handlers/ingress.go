package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/containerum/cherry/adaptors/gonic"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	ingressParam = "ingress"
)

// swagger:operation GET /projects/{project}/namespaces/{namespace}/ingresses Ingress GetIngressList
// Get ingresses list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: project
//    in: path
//    type: string
//    required: true
//  - name: namespace
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: ingresses list
//    schema:
//      $ref: '#/definitions/IngressesList'
//  default:
//    $ref: '#/responses/error'
func GetIngressList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get ingress list")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	ingressList, err := kube.GetIngressList(namespace)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := httputil.MustGetUserID(ctx.Request.Context())
	ret, err := model.ParseKubeIngressList(ingressList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /projects/{project}/namespaces/{namespace}/ingresses/{ingress} Ingress GetIngress
// Get ingresses list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: project
//    in: path
//    type: string
//    required: true
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
//      $ref: '#/definitions/Ingress'
//  default:
//    $ref: '#/responses/error'
func GetIngress(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Ingress":   ingr,
	}).Debug("Get ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	ingress, err := kube.GetIngress(namespace, ingr)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := httputil.MustGetUserID(ctx.Request.Context())
	ret, err := model.ParseKubeIngress(ingress, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /projects/{project}/namespaces/{namespace}/ingresses Ingress CreateIngress
// Create ingress.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: project
//    in: path
//    type: string
//    required: true
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Ingress'
// responses:
//  '201':
//    description: ingress created
//    schema:
//      $ref: '#/definitions/Ingress'
//  default:
//    $ref: '#/responses/error'
func CreateIngress(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingressReq model.IngressKubeAPI
	if err := ctx.ShouldBindWith(&ingressReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	newIngress, errs := ingressReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kube.CreateIngress(newIngress)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := httputil.MustGetUserID(ctx.Request.Context())
	ret, err := model.ParseKubeIngress(ingressAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /projects/{project}/namespaces/{namespace}/ingresses/{ingress} Ingress UpdateIngress
// Update ingress.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: project
//    in: path
//    type: string
//    required: true
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
//      $ref: '#/definitions/Ingress'
// responses:
//  '201':
//    description: ingress updated
//    schema:
//      $ref: '#/definitions/Ingress'
//  default:
//    $ref: '#/responses/error'
func UpdateIngress(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Ingress":   ingr,
	}).Debug("Update ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var ingressReq model.IngressKubeAPI
	if err := ctx.ShouldBindWith(&ingressReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldIngress, err := kube.GetIngress(namespace, ingr)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	ingressReq.Name = ingr
	ingressReq.Owner = oldIngress.GetObjectMeta().GetLabels()[ownerQuery]

	newIngress, errs := ingressReq.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	ingressAfter, err := kube.UpdateIngress(newIngress)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := httputil.MustGetUserID(ctx.Request.Context())
	ret, err := model.ParseKubeIngress(ingressAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /projects/{project}/namespaces/{namespace}/ingresses/{ingress} Ingress DeleteIngress
// Delete ingress.
//
// ---
// x-method-visibility: private
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - name: project
//    in: path
//    type: string
//    required: true
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
//    $ref: '#/responses/error'
func DeleteIngress(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	ingr := ctx.Param(ingressParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Ingress":   ingr,
	}).Debug("Delete ingress Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteIngress(namespace, ingr)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
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
// responses:
//  '200':
//    description: ingresses list from all users namespaces
//    schema:
//      $ref: '#/definitions/SelectedIngressesList'
//  default:
//    $ref: '#/responses/error'
func GetSelectedIngresses(ctx *gin.Context) {
	log.Debug("Get selected ingresses Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ingresses := make(kube_types.SelectedIngressesList, 0)

	role := httputil.MustGetUserID(ctx.Request.Context())
	if role == m.RoleUser {
		accesses := ctx.MustGet(httputil.AllAccessContext).([]httputil.ProjectAccess)
		for _, p := range accesses {
			for _, n := range p.NamespacesAccesses {
				ingressList, err := kube.GetIngressList(n.NamespaceID)
				if err != nil {
					gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
					return
				}
				ingress, err := model.ParseKubeIngressList(ingressList, role == m.RoleUser)
				if err != nil {
					gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
					return
				}
				ingresses[n.NamespaceID] = *ingress
			}
		}
	}

	ctx.JSON(http.StatusOK, ingresses)
}
