package middleware

import (
	"errors"

	"git.containerum.net/ch/kube-api/pkg/model"

	"github.com/gin-gonic/gin"

	"fmt"

	"net/textproto"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"github.com/containerum/cherry/adaptors/gonic"
	headers "github.com/containerum/utils/httputil"
	log "github.com/sirupsen/logrus"
)

var (
	ErrInvalidUserRole = errors.New("Invalid user role")
)

func RequiredUserHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.WithField("Headers", ctx.Request.Header).Debug("Header list")
		notFoundHeaders := requireHeaders(ctx, headers.UserRoleXHeader)
		if len(notFoundHeaders) > 0 {
			gonic.Gonic(kubeErrors.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
			return
		}
		/* Check User-Role and User-Namespace, X-User-Volume */
		if isUser, err := checkIsUserRole(GetHeader(ctx, headers.UserRoleXHeader)); err != nil {
			log.WithField("Value", GetHeader(ctx, headers.UserRoleXHeader)).WithError(err).Warn("Check User-Role Error")
			gonic.Gonic(kubeErrors.ErrInvalidRole(), ctx)
		} else {
			//User-Role: user, check User-Namespace, X-User-Volume
			if isUser {
				notFoundHeaders := requireHeaders(ctx, headers.UserRoleXHeader, headers.UserNamespacesXHeader, headers.UserVolumesXHeader, headers.UserIDXHeader)
				if len(notFoundHeaders) > 0 {
					gonic.Gonic(kubeErrors.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
					return
				}
				userNs, errNs := checkUserNamespace(GetHeader(ctx, headers.UserNamespacesXHeader))
				userVol, errVol := checkUserVolume(GetHeader(ctx, headers.UserVolumesXHeader))
				if errNs != nil {
					log.WithField("Value", GetHeader(ctx, headers.UserNamespacesXHeader)).WithError(errNs).Warn("Check User-Namespace header Error")
					gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetails(fmt.Sprintf("%v: %v", headers.UserNamespacesXHeader, errNs)), ctx)
					return
				}
				if errVol != nil {
					log.WithField("Value", GetHeader(ctx, headers.UserVolumesXHeader)).WithError(errVol).Warn("Check User-Volume header Error")
					gonic.Gonic(kubeErrors.ErrRequestValidationFailed().AddDetails(fmt.Sprintf("%v: %v", headers.UserVolumesXHeader, errVol)), ctx)
					return
				}
				ctx.Set(UserNamespaces, userNs)
				ctx.Set(UserVolumes, userVol)
				ctx.Set(UserID, GetHeader(ctx, headers.UserIDXHeader))
			}
		}
		ctx.Set(UserRole, GetHeader(ctx, headers.UserRoleXHeader))
	}
}

func checkIsUserRole(userRole string) (bool, error) {
	switch userRole {
	case "", RoleAdmin:
		return false, nil
	case RoleUser:
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
		if GetHeader(ctx, v) == "" {
			notFoundHeaders = append(notFoundHeaders, v)
		}
	}
	return
}

func GetHeader(ctx *gin.Context, header string) string {
	return ctx.GetHeader(textproto.CanonicalMIMEHeaderKey(header))
}
