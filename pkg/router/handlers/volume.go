package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeerrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	volumeParam = "volume"
)

// swagger:operation GET /namespaces/{namespace}/volumes Volume GetVolumeList
// Get volumes list.
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
//    description: volumes list
//    schema:
//      $ref: '#/definitions/VolumesList'
//  default:
//    $ref: '#/responses/error'
func GetVolumeList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get volumes list call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	svcList, err := kube.GetPersistentVolumeClaimsList(namespace)
	if err != nil {
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubePersistentVolumeClaimList(svcList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}
	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/volumes/{volume} Volume GetVolume
// Get volumes list.
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
//  - name: volume
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: volume
//    schema:
//      $ref: '#/definitions/Volume'
//  default:
//    $ref: '#/responses/error'
func GetVolume(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	volume := ctx.Param(volumeParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Volume":    volume,
	}).Debug("Get volume call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResource()), ctx)
		return
	}

	svc, err := kube.GetPersistentVolumeClaim(namespace, volume)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubePersistentVolumeClaim(svc, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/volume Volume CreateVolume
// Create volume.
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
//      $ref: '#/definitions/Volume'
// responses:
//  '201':
//    description: volume created
//    schema:
//      $ref: '#/definitions/Volume'
//  default:
//    $ref: '#/responses/error'
func CreateVolume(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create volume Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var pvc model.VolumeKubeAPI
	if err := ctx.ShouldBindWith(&pvc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableCreateResource(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableCreateResource()), ctx)
		return
	}

	newPvc, errs := pvc.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	pvcAfter, err := kube.CreatePersistentVolumeClaim(newPvc)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubePersistentVolumeClaim(pvcAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/volumes/{volume} Volume UpdateVolume
// Update volume.
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
//  - name: volume
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Volume'
// responses:
//  '202':
//    description: volume updated
//    schema:
//      $ref: '#/definitions/Volume'
//  default:
//    $ref: '#/responses/error'
func UpdateVolume(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	vol := ctx.Param(volumeParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Volume":    vol,
	}).Debug("Update volume Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var pvc model.VolumeKubeAPI
	if err := ctx.ShouldBindWith(&pvc, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableUpdateResource(), ctx)
		return
	}

	_, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldPvc, err := kube.GetPersistentVolumeClaim(namespace, vol)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	newPvc, err := pvc.Resize(oldPvc)
	if err != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	updatedPvc, err := kube.UpdatePersistentVolumeClaim(newPvc)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubePersistentVolumeClaim(updatedPvc, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/volume/{volume} Volume DeleteVolume
// Delete volume.
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
//  - name: volume
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: volume deleted
//  default:
//    $ref: '#/responses/error'
func DeleteVolume(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	volume := ctx.Param(volumeParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Volume":    volume,
	}).Debug("Delete volume call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeletePersistentVolumeClaim(namespace, volume)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
