package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeerrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// swagger:operation GET /storages Service GetStoragesList
// Get storages list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
// responses:
//  '200':
//    description: storages list
//    schema:
//      $ref: '#/definitions/StorageList'
//  default:
//    $ref: '#/responses/error'
func GetStoragesList(ctx *gin.Context) {
	log.Debug("Get storages list call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	storageList, err := kube.GetStorageClassesList()
	if err != nil {
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}
	ret, err := model.ParseStoragesList(storageList)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeerrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}
