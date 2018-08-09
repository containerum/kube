package handlers

import (
	"net/http"

	"sync"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/containerum/cherry/adaptors/gonic"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	configMapParam = "configmap"
)

// swagger:operation GET /namespaces/{namespace}/configmaps ConfigMap GetConfigMapList
// Get config maps list.
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
//    description: config maps list
//    schema:
//      $ref: '#/definitions/ConfigMapsList'
//  default:
//    $ref: '#/responses/error'
func GetConfigMapList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get config maps list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	cmList, err := kube.GetConfigMapList(namespace)
	if err != nil {
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMapList(cmList, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/configmaps/{configmap} ConfigMap GetConfigMap
// Get config map.
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
//  - name: configmap
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: config map
//    schema:
//      $ref: '#/definitions/ConfigMap'
//  default:
//    $ref: '#/responses/error'
func GetConfigMap(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"ConfigMap": configMap,
	}).Debug("Get config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResource()), ctx)
		return
	}

	cm, err := kube.GetConfigMap(namespace, configMap)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeConfigMap(cm, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/configmaps ConfigMap CreateConfigMap
// Create config map.
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
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/ConfigMap'
// responses:
//  '201':
//    description: config map created
//    schema:
//      $ref: '#/definitions/ConfigMap'
//  default:
//    $ref: '#/responses/error'
func CreateConfigMap(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cmReq model.ConfigMapKubeAPI
	if err := ctx.ShouldBindWith(&cmReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableCreateResource()), ctx)
		return
	}

	cm, errs := cmReq.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	cmAfter, err := kube.CreateConfigMap(cm)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableCreateResource()), ctx)
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
//  - name: configmap
//    in: path
//    type: string
//    required: true
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/ConfigMap'
// responses:
//  '202':
//    description: config map updated
//    schema:
//      $ref: '#/definitions/ConfigMap'
//  default:
//    $ref: '#/responses/error'
func UpdateConfigMap(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"ConfigMap": configMap,
	}).Debug("Update config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var cmReq model.ConfigMapKubeAPI
	if err := ctx.ShouldBindWith(&cmReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldCm, err := kube.GetConfigMap(namespace, ctx.Param(configMapParam))
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
		return
	}

	cmReq.Name = configMap
	cmReq.Owner = oldCm.GetObjectMeta().GetLabels()[ownerQuery]

	newCm, errs := cmReq.ToKube(namespace, ns.Labels)
	if errs != nil {
		gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	cmAfter, err := kube.UpdateConfigMap(newCm)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableUpdateResource()), ctx)
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
	namespace := ctx.Param(namespaceParam)
	configMap := ctx.Param(configMapParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"ConfigMap": configMap,
	}).Debug("Delete config map Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteConfigMap(namespace, configMap)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeerrors.ErrUnableDeleteResource()), ctx)
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

	ret := make(kube_types.SelectedConfigMapsList)

	role := ctx.MustGet(m.UserRole).(string)
	if role == m.RoleUser {
		nsList := ctx.MustGet(m.UserNamespaces).(*model.UserHeaderDataMap)
		var g errgroup.Group
		var mutex = &sync.Mutex{}
		for _, n := range *nsList {
			currentNs := n
			g.Go(func() error {
				cmList, err := kube.GetConfigMapList(currentNs.ID)
				if err != nil {
					return err
				}
				cm, err := model.ParseKubeConfigMapList(cmList, role == m.RoleUser)
				if err != nil {
					return err
				}
				mutex.Lock()
				ret[currentNs.ID] = *cm
				mutex.Unlock()
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			ctx.Error(err)
			gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
			return
		}
	}

	ctx.JSON(http.StatusOK, ret)
}
