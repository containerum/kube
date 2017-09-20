package server

import "github.com/gin-gonic/gin"

func CreateNamespace(c *gin.Context) {
	c.AbortWithStatus(201)
}
