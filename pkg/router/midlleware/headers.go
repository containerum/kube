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
			gonic.Gonic(kubeerrors.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
			return
		}
		// Check User-Role and User-Namespace
		if isUser, err := checkIsUserRole(GetHeader(ctx, headers.UserRoleXHeader)); err != nil {
			log.WithField("Value", GetHeader(ctx, headers.UserRoleXHeader)).WithError(err).Warn("Check User-Role Error")
			gonic.Gonic(kubeerrors.ErrInvalidRole(), ctx)
		} else {
			// User-Role: user, check User-Namespace
			if isUser {
				notFoundHeaders := requireHeaders(ctx, headers.UserRoleXHeader, headers.UserNamespacesXHeader, headers.UserIDXHeader)
				if len(notFoundHeaders) > 0 {
					gonic.Gonic(kubeerrors.ErrRequiredHeadersNotProvided().AddDetails(notFoundHeaders...), ctx)
					return
				}
				userNs, errNs := checkUserNamespace(GetHeader(ctx, headers.UserNamespacesXHeader))
				if errNs != nil {
					log.WithField("Value", GetHeader(ctx, headers.UserNamespacesXHeader)).WithError(errNs).Warn("Check User-Namespace header Error")
					gonic.Gonic(kubeerrors.ErrRequestValidationFailed().AddDetails(fmt.Sprintf("%v: %v", headers.UserNamespacesXHeader, errNs)), ctx)
					return
				}
				ctx.Set(UserNamespaces, userNs)
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
