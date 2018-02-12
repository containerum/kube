package middleware

import (
	"errors"
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/model"

	"github.com/gin-gonic/gin"

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
	return func(c *gin.Context) {
		/* Check User-Role and User-Namespace, X-User-Volume */
		if isUser, err := checkIsUserRole(c.GetHeader(userRoleXHeader)); err != nil {
			log.WithField("Value", c.GetHeader(userRoleXHeader)).WithError(err).Warn("Check User-Role Error")
			c.AbortWithStatus(http.StatusForbidden)
		} else {
			//User-Role: user, check User-Namespace, X-User-Volume
			if isUser {
				userNs, errNs := checkUserNamespace(c.GetHeader(userNamespaceXHeader))
				userVol, errVol := checkUserVolume(c.GetHeader(userVolumeXHeader))
				if errNs != nil {
					log.WithField("Value", c.GetHeader(userNamespaceXHeader)).WithError(errNs).Warn("Check User-Namespace header Error")
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
				if errVol != nil {
					log.WithField("Value", c.GetHeader(userVolumeXHeader)).WithError(errVol).Warn("Check User-Volume header Error")
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
				c.Set(UserNamespaces, userNs)
				c.Set(UserVolumes, userVol)
			}
		}
		c.Set(UserRole, c.GetHeader(userRoleXHeader))
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
