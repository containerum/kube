package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccessLevel string

const (
	levelOwner      AccessLevel = "owner"
	levelRead       AccessLevel = "read"
	levelWrite      AccessLevel = "write"
	levelReadDelete AccessLevel = "read-delete"
	levelNone       AccessLevel = ""
)

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role := c.GetHeader(userRoleXHeader); role != "admin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}
