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
	api_core "k8s.io/api/core/v1"
)

const (
	secretParam = "secret"
)

// swagger:operation GET /namespaces/{namespace}/secrets Secret GetSecretList
// Get TLS secrets list.
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
//  - name: docker
//    in: query
//    type: string
//    required: false
// responses:
//  '200':
//    description: secrets list
//    schema:
//      $ref: '#/definitions/SecretsList'
//  default:
//    $ref: '#/responses/error'
func GetSecretList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Get secret list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

	_, isDocker := ctx.GetQuery("docker")

	var secrets *api_core.SecretList
	if isDocker {
		secrets, err = kube.GetDockerSecretList(namespace)
	} else {
		secrets, err = kube.GetTLSSecretList(namespace)
	}
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecretList(secrets, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation GET /namespaces/{namespace}/secrets/{secret} Secret GetSecret
// Get secret.
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
//  - name: secret
//    in: path
//    type: string
//    required: true
// responses:
//  '200':
//    description: secret
//    schema:
//      $ref: '#/definitions/Secret'
//  default:
//    $ref: '#/responses/error'
func GetSecret(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	sct := ctx.Param(secretParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Secret":    sct,
	}).Debug("Get secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	secret, err := kube.GetSecret(namespace, sct)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecret(secret, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetResource(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

// swagger:operation POST /namespaces/{namespace}/secrets/tls Secret CreateTLSSecret
// Create secret.
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
//      $ref: '#/definitions/Secret'
// responses:
//  '201':
//    description: secret created
//    schema:
//      $ref: '#/definitions/Secret'
//  default:
//    $ref: '#/responses/error'
func CreateTLSSecret(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secretReq model.SecretKubeAPI
	if err := ctx.ShouldBindWith(&secretReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	newSecret, errs := secretReq.ToKube(namespace, ns.Labels, api_core.SecretTypeOpaque)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	secretAfter, err := kube.CreateSecret(newSecret)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecret(secretAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation POST /namespaces/{namespace}/secrets/docker Secret CreateDockerSecret
// Create secret.
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
//      $ref: '#/definitions/Secret'
// responses:
//  '201':
//    description: secret created
//    schema:
//      $ref: '#/definitions/Secret'
//  default:
//    $ref: '#/responses/error'
func CreateDockerSecret(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
	}).Debug("Create secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secretReq model.SecretKubeAPI
	if err := ctx.ShouldBindWith(&secretReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	newSecret, errs := secretReq.ToKube(namespace, ns.Labels, api_core.SecretTypeDockerConfigJson)
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	secretAfter, err := kube.CreateSecret(newSecret)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableCreateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecret(secretAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

// swagger:operation PUT /namespaces/{namespace}/secrets/{secret} Secret UpdateSecret
// Update secret.
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
//  - name: secret
//    in: path
//    type: string
//    required: true
//  - name: docker
//    in: query
//    type: string
//    required: false
//  - name: body
//    in: body
//    schema:
//      $ref: '#/definitions/Secret'
// responses:
//  '202':
//    description: secret updated
//    schema:
//      $ref: '#/definitions/Secret'
//  default:
//    $ref: '#/responses/error'
func UpdateSecret(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	sct := ctx.Param(secretParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Secret":    sct,
	}).Debug("Update secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secretReq model.SecretKubeAPI
	if err := ctx.ShouldBindWith(&secretReq, binding.JSON); err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed(), ctx)
		return
	}

	ns, err := kube.GetNamespaceQuota(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	oldSecret, err := kube.GetSecret(namespace, sct)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	secretReq.Name = sct
	secretReq.Owner = oldSecret.GetObjectMeta().GetLabels()[ownerQuery]

	_, isDocker := ctx.GetQuery("docker")

	var newSecret *api_core.Secret
	var errs []error
	if isDocker {
		newSecret, errs = secretReq.ToKube(namespace, ns.Labels, api_core.SecretTypeDockerConfigJson)
	} else {
		newSecret, errs = secretReq.ToKube(namespace, ns.Labels, api_core.SecretTypeOpaque)
	}
	if errs != nil {
		gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	secretAfter, err := kube.UpdateSecret(newSecret)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableUpdateResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	ret, err := model.ParseKubeSecret(secretAfter, role == m.RoleUser)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

// swagger:operation DELETE /namespaces/{namespace}/secrets/{secret} Secret DeleteSecret
// Delete secret.
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
//  - name: secret
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: secret deleted
//  default:
//    $ref: '#/responses/error'
func DeleteSecret(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	sct := ctx.Param(secretParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Secret":    sct,
	}).Debug("Delete secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeleteSecret(namespace, sct)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
