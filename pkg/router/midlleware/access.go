package middleware

import (
	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/model"
	headers "github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"

	"github.com/containerum/cherry/adaptors/gonic"
	kubeModel "github.com/containerum/kube-client/pkg/model"
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
		gonic.Gonic(kubeErrors.ErrAdminRequired(), ctx)
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
			gonic.Gonic(kubeErrors.ErrAccessError(), c)
			return
		}
		gonic.Gonic(kubeErrors.ErrResourceNotExist(), c)
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
			gonic.Gonic(kubeErrors.ErrAccessError(), c)
			return
		}
		gonic.Gonic(kubeErrors.ErrResourceNotExist(), c)
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
