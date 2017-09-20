package router

import (
	"net/http"

	"bitbucket.org/exonch/kube-api/server"

	"github.com/gin-gonic/gin"
)

func Load(debug bool, middleware ...gin.HandlerFunc) http.Handler {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(middleware...)

	e.Use(func(c *gin.Context) {
		c.Set("debug", debug)
	})

	e.NoRoute(func(c *gin.Context) {
		c.Status(404)
	})

	namespace := e.Group("/api/namespace")
	{
		namespace.POST("", server.CreateNamespace)
	}

	return e
}
