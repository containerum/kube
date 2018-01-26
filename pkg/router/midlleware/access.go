package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role := c.MustGet(UserRole).(string); role != "admin" {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}
