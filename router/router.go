package router

import (
	"net/http"

	"bitbucket.org/exonch/kube-api/router/middleware"
	"bitbucket.org/exonch/kube-api/server"
	"bitbucket.org/exonch/kube-api/utils"

	"github.com/gin-gonic/gin"
)

func Load(debug bool, middlewares ...gin.HandlerFunc) http.Handler {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(middleware.SetRequestID)
	e.Use(utils.AddLogger)
	e.Use(middlewares...)

	e.Use(func(c *gin.Context) {
		c.Set("debug", debug)
	})

	e.NoRoute(func(c *gin.Context) {
		c.Status(404)
	})

	namespace := e.Group("/api/namespace")
	{
		namespace.Use(middleware.SetRandomKubeClient)

		namespace.GET("",
			server.ListNamespaces)
		namespace.POST("",
			middleware.ParseJSON,
			server.CreateNamespace)
		namespace.GET("/:namespace",
			middleware.SetNamespace,
			server.GetNamespace)
		namespace.DELETE("/:namespace",
			middleware.SetNamespace,
			server.DeleteNamespace)
	}

	deployment := e.Group("/api/namespace/:namespace/deployment")
	{
		deployment.Use(middleware.SetNamespace)
		deployment.Use(middleware.SetRandomKubeClient)

		deployment.GET("",
			server.ListDeployments)
		deployment.POST("",
			middleware.ParseJSON,
			server.CreateDeployment)
		deployment.GET("/:objname",
			middleware.SetObjectName,
			server.GetDeployment)
		deployment.DELETE("/:objname",
			middleware.SetObjectName,
			server.DeleteDeployment)
	}

	return e
}
