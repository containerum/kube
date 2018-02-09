package middleware

import (
	"errors"
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/model"
	kube_types "git.containerum.net/ch/kube-client/pkg/model"

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

func RequiredHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		/* Check User-Role and User-Namespace, X-User-Volume */
		if ok, err := checkUserRole(c.GetHeader(userRoleXHeader)); err != nil {
			log.WithField("Value", c.GetHeader(userRoleXHeader)).WithError(err).Warn("Check User-Role Error")
			c.AbortWithStatus(http.StatusForbidden)
		} else {
			//User-Role: user, check User-Namespace, X-User-Volume
			if !ok {
				userNs, errNs := checkUserNamespace(c.GetHeader(userNamespaceXHeader))
				userVol, errVol := checkUserVolume(c.GetHeader(userVolumeXHeader))
				if errNs != nil {
					log.WithField("Value", c.GetHeader(userNamespaceXHeader)).WithError(errNs).Warn("Check User-Namespace header Error")
					c.AbortWithStatus(http.StatusForbidden)
				}
				if errVol != nil {
					log.WithField("Value", c.GetHeader(userVolumeXHeader)).WithError(errVol).Warn("Check User-Volume header Error")
					c.AbortWithStatus(http.StatusForbidden)
				}
				c.Set(UserNamespaces, userNs)
				c.Set(UserVolumes, userVol)
			} else {
				c.Set(UserRole, c.GetHeader(userRoleXHeader))
			}
		}
	}
}

func checkUserRole(userRole string) (bool, error) {
	switch userRole {
	case "admin":
		return true, nil
	case "user":
		return false, nil
	}
	return false, ErrInvalidUserRole
}

func checkUserNamespace(userNamespace string) (*kube_types.UserHeaderData, error) {
	return model.ParseUserHeaderData(userNamespace)
}

func checkUserVolume(userVolume string) (*kube_types.UserHeaderData, error) {
	return model.ParseUserHeaderData(userVolume)
}
