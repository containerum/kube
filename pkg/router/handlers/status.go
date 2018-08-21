package handlers

import (
	"net/http"

	"github.com/containerum/kube-client/pkg/model"
	"github.com/gin-gonic/gin"
)

// swagger:operation GET /status Status ServiceStatus
// Get service status list.
//
// ---
// x-method-visibility: public
// responses:
//  '200':
//    description: service status
//    schema:
//      $ref: '#/definitions/ServiceStatus'
//  default:
//    $ref: '#/responses/error'
func ServiceStatus(status *model.ServiceStatus) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var httpStatus int
		if status.StatusOK {
			httpStatus = http.StatusOK
		} else {
			httpStatus = http.StatusInternalServerError
		}
		ctx.JSON(httpStatus, status)
	}
}
