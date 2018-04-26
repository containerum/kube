package middleware

import (
	"git.containerum.net/ch/api-gateway/pkg/utils/headers"
	cherry "git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/model"
	kubeModel "git.containerum.net/ch/kube-client/pkg/model"

	"git.containerum.net/ch/cherry/adaptors/gonic"
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

var (
	writeLevels = []AccessLevel{
		levelOwner,
		levelWrite,
	}
)

const (
	namespaceParam = "namespace"
	RoleUser       = "user"
	RoleAdmin      = "admin"
)

func IsAdmin(ctx *gin.Context) {
	if role := GetHeader(ctx, headers.UserRoleXHeader); role != RoleAdmin {
		gonic.Gonic(cherry.ErrAdminRequired(), ctx)
		return
	}
	return
}

func ReadAccess(c *gin.Context) {
	ns := c.MustGet(NamespaceKey).(string)
	if c.MustGet(UserRole).(string) == RoleUser {
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
			gonic.Gonic(cherry.ErrAccessError(), c)
			return
		}
		gonic.Gonic(cherry.ErrResourceNotExist(), c)
		return
	}
	return
}

func WriteAccess(c *gin.Context) {
	ns := c.MustGet(NamespaceKey).(string)
	if c.MustGet(UserRole).(string) == RoleUser {
		var userNsData *kubeModel.UserHeaderData
		nsList := c.MustGet(UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {
			if ns == n.Label {
				userNsData = &n
				break
			}
		}
		if userNsData != nil {
			if ok := containsAccess(userNsData.Access, writeLevels...); ok {
				c.Set(NamespaceKey, userNsData.ID)
				c.Set(NamespaceLabelKey, userNsData.Label)
				return
			}
			gonic.Gonic(cherry.ErrAccessError(), c)
			return
		}
		gonic.Gonic(cherry.ErrResourceNotExist(), c)
		return
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
