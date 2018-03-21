package middleware

import (
	"errors"

	"git.containerum.net/ch/kube-api/pkg/model"

	"github.com/gin-gonic/gin"

	"fmt"

	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	log "github.com/sirupsen/logrus"
)

const (
	userRoleXHeader      = "X-User-Role"
	userNamespaceXHeader = "X-User-Namespace"
	userVolumeXHeader    = "X-User-Volume"
)

var (
	ErrInvalidUserRole = errors.New("Invalid user role")
)

func RequiredUserHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.WithField("Headers", ctx.Request.Header).Debug("Header list")
		notFoundHeaders := requireHeaders(ctx, userRoleXHeader)
		if len(notFoundHeaders) > 0 {
			gonic.Gonic(cherry.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
			return
		}
		/* Check User-Role and User-Namespace, X-User-Volume */
		if isUser, err := checkIsUserRole(ctx.GetHeader(userRoleXHeader)); err != nil {
			log.WithField("Value", ctx.GetHeader(userRoleXHeader)).WithError(err).Warn("Check User-Role Error")
			gonic.Gonic(cherry.ErrInvalidRole(), ctx)
		} else {
			//User-Role: user, check User-Namespace, X-User-Volume
			if isUser {
				notFoundHeaders := requireHeaders(ctx, userNamespaceXHeader, userVolumeXHeader)
				if len(notFoundHeaders) > 0 {
					gonic.Gonic(cherry.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
					return
				}
				userNs, errNs := checkUserNamespace(ctx.GetHeader(userNamespaceXHeader))
				userVol, errVol := checkUserVolume(ctx.GetHeader(userVolumeXHeader))
				if errNs != nil {
					log.WithField("Value", ctx.GetHeader(userNamespaceXHeader)).WithError(errNs).Warn("Check User-Namespace header Error")
					gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetails(fmt.Sprintf("%v: %v", userNamespaceXHeader, errNs)), ctx)
					return
				}
				if errVol != nil {
					log.WithField("Value", ctx.GetHeader(userVolumeXHeader)).WithError(errVol).Warn("Check User-Volume header Error")
					gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetails(fmt.Sprintf("%v: %v", userVolumeXHeader, errVol)), ctx)
					return
				}
				ctx.Set(UserNamespaces, userNs)
				ctx.Set(UserVolumes, userVol)
			}
		}
		ctx.Set(UserRole, ctx.GetHeader(userRoleXHeader))
	}
}

func checkIsUserRole(userRole string) (bool, error) {
	switch userRole {
	case "", "admin":
		return false, nil
	case "user":
		return true, nil
	}
	return false, ErrInvalidUserRole
}

func checkUserNamespace(userNamespace string) (*model.UserHeaderDataMap, error) {
	return model.ParseUserHeaderData(userNamespace)
}

func checkUserVolume(userVolume string) (*model.UserHeaderDataMap, error) {
	return model.ParseUserHeaderData(userVolume)
}

func requireHeaders(ctx *gin.Context, headers ...string) (notFoundHeaders []string) {
	for _, v := range headers {
		if ctx.GetHeader(v) == "" {
			notFoundHeaders = append(notFoundHeaders, v)
		}
	}
	return
}
