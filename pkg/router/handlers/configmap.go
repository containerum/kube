package handlers

import (
	"net/http"

	"git.containerum.net/ch/cherry/adaptors/gonic"
	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	configMapParam = "configmap"
)

// swagger:operation GET /namespaces/{namespace}/configmaps ConfigMap GetConfigMapList
// Get config maps list.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#configmaps-list
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
//    description: config maps list
//    schema:
//      $ref: '#/definitions/ConfigMapsList'
//  default:
//    $ref: '#/responses/error'
func GetConfigMapList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Get config maps list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	cmList, err := kube.GetConfigMapList(namespace)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMapList(cmList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/configmaps/{configmap} ConfigMap GetConfigMap
// Get config map.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-configmap
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
//  - name: configmap
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: config map
//    schema:
//      $ref: '#/definitions/ConfigMapWithOwner'
//  default:
//    $ref: '#/responses/error'
func GetConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"ConfigMap":       configMap,
	}).Debug("Get config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	cm, err := kube.GetConfigMap(namespace, configMap)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMap(cm, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/configmaps ConfigMap CreateConfigMap
// Create config map.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#create-configmap
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
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/ConfigMapWithOwner'
// responses:
//  '201':
//    description: config map created
//    schema:
//      $ref: '#/definitions/ConfigMapWithOwner'
//  default:
//    $ref: '#/responses/error'
func CreateConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
	}).Debug("Create config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cmReq model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cmReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	cm, errs := cmReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	cmAfter, err := kube.CreateConfigMap(cm)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMap(cmAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/configmaps/{configmap} ConfigMap UpdateConfigMap
// Update config map.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#update-configmap
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
//  - name: configmap
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/ConfigMapWithOwner'
// responses:
//  '202':
//    description: config map updated
//    schema:
//      $ref: '#/definitions/ConfigMapWithOwner'
//  default:
//    $ref: '#/responses/error'
func UpdateConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"ConfigMap":       configMap,
	}).Debug("Create config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cmReq model.ConfigMapWithOwner
	if err := ctx.ShouldBindWith(&cmReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	quota, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldCm, err := kube.GetConfigMap(namespace, ctx.Param(deploymentParam))
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	cmReq.Name = configMap
	cmReq.Owner = oldCm.GetObjectMeta().GetLabels()[ownerQuery]

	newCm, errs := cmReq.ToKube(namespace, quota.Labels)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	cmAfter, err := kube.UpdateConfigMap(newCm)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMap(cmAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/configmaps/{configmap} ConfigMap DeleteConfigMap
// Delete config map.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#delete-configmap
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
//  - name: configmap
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: config map deleted
//  default:
//    $ref: '#/responses/error'
func DeleteConfigMap(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"ConfigMap":       configMap,
	}).Debug("Delete config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	err := kube.DeleteConfigMap(namespace, configMap)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	ctx.Status(http.StatusAccepted)
}

// swagger:operation GET /configmaps ConfigMap GetSelectedConfigMaps
// Get config maps from all user namespaces.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
// responses:
//  '200':
//    description: config maps list from all users namespaces
//    schema:
//      $ref: '#/definitions/SelectedConfigMapsList'
//  default:
//    $ref: '#/responses/error'
func GetSelectedConfigMaps(ctx *gin.Context) {
	log.Debug("Get selected config maps Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	ret := make(model.SelectedConfigMapsList, 0)

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {
			cmList, err := kube.GetConfigMapList(n.ID)
			if err != nil {
				gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
				return
			}

			cm, err := model.ParseKubeConfigMapList(cmList, role == m.RoleUser)
			if err != nil {
				gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
				return
			}

			ret[n.Label] = *cm
		}
	}

	ctx.JSON(http.StatusOK, ret)
}
