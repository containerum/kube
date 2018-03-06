package middleware

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/model"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	kubeModel "git.containerum.net/ch/kube-client/pkg/model"

	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
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
	readLevels = []AccessLevel{
		levelOwner,
		levelWrite,
		levelReadDelete,
		levelRead,
	}
)

const (
	namespaceParam = "namespace"
)

func IsAdmin(ctx *gin.Context) {
	if role := ctx.GetHeader(userRoleXHeader); role != "admin" {
		gonic.Gonic(cherry.ErrAdminRequired(), ctx)
		return
	}
	return
}

func ReadAccess(c *gin.Context) {
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
				c.Set(NamespaceLabelKey, userNsData.Label)
				return
			}
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
	return
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
