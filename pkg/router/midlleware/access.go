package middleware

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/model"
	kubeModel "git.containerum.net/ch/kube-client/pkg/model"

	"github.com/gin-gonic/gin"
)

type AccessLevel string

const (
	levelOwner      AccessLevel = "owner"
	levelWrite      AccessLevel = "write"
	levelReadDelete AccessLevel = "read-delete"
	levelRead       AccessLevel = "read"
	levelNone       AccessLevel = ""
)

var (
	readLevels []AccessLevel = []AccessLevel{
		levelOwner,
		levelWrite,
		levelReadDelete,
		levelRead,
	}
)

const (
	namespaceParam = "namespace"
)

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role := c.GetHeader(userRoleXHeader); role != "admin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}

func ReadAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.MustGet(NamespaceKey).(string)
		if c.MustGet(UserRole).(string) == "user" {
			var userNsData *kubeModel.UserHeaderData
			nsList := c.MustGet(UserNamespaces).(*model.UserHeaderDataMap)
			for _, n := range *nsList {
				if ns == n.Label {
					userNsData = &n
					break
				}
			}
			if userNsData != nil {
				if ok := containsAccess(userNsData.Access, readLevels...); ok {
					c.Set(NamespaceKey, userNsData.ID)
					return
				}
			} else {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
	}
}

func containsAccess(access string, in ...AccessLevel) bool {
	contains := false
	userAccess := AccessLevel(access)
	for _, acc := range in {
		if acc == userAccess {
			return true
		}
	}
	return contains
}
