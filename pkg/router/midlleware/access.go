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
	deleteLevels = []AccessLevel{
		levelOwner,
		levelWrite,
		levelReadDelete,
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

func ReadAccess(ctx *gin.Context) {
	ns := ctx.Param("namespace")
	if GetHeader(ctx, headers.UserRoleXHeader) == RoleUser {
		var userNsData *kubeModel.UserHeaderData
		nsList := ctx.MustGet(UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {
			if ns == n.ID {
				userNsData = &n
				break
			}
		}
		if userNsData != nil {
			if ok := containsAccess(userNsData.Access, readLevels...); ok {
				return
			}
			gonic.Gonic(kubeErrors.ErrAccessError(), ctx)
			return
		}
		gonic.Gonic(kubeErrors.ErrResourceNotExist(), ctx)
		return
	}
	return
}

func DeleteAccess(ctx *gin.Context) {
	ns := ctx.Param("namespace")
	if GetHeader(ctx, headers.UserRoleXHeader) == RoleUser {
		var userNsData *kubeModel.UserHeaderData
		nsList := ctx.MustGet(UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {
			if ns == n.ID {
				userNsData = &n
				break
			}
		}
		if userNsData != nil {
			if ok := containsAccess(userNsData.Access, readLevels...); ok {
				return
			}
			gonic.Gonic(kubeErrors.ErrAccessError(), ctx)
			return
		}
		gonic.Gonic(kubeErrors.ErrResourceNotExist(), ctx)
		return
	}
	return
}

func WriteAccess(ctx *gin.Context) {
	ns := ctx.Param("namespace")
	if GetHeader(ctx, headers.UserRoleXHeader) == RoleUser {
		var userNsData *kubeModel.UserHeaderData
		nsList := ctx.MustGet(UserNamespaces).(*model.UserHeaderDataMap)
		for _, n := range *nsList {
			if ns == n.ID {
				userNsData = &n
				break
			}
		}
		if userNsData != nil {
			if ok := containsAccess(userNsData.Access, writeLevels...); ok {
				return
			}
			gonic.Gonic(kubeErrors.ErrAccessError(), ctx)
			return
		}
		gonic.Gonic(kubeErrors.ErrResourceNotExist(), ctx)
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
