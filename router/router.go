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
	e.Use(middleware.WriteResponseObject)    //not reversed
	e.Use(middleware.RedactResponseMetadata) //not reversed
	e.Use(middleware.SetRequestID)
	e.Use(utils.AddLogger)
	e.Use(middleware.CheckHTTP411)
	e.Use(middlewares...)

	e.Use(func(c *gin.Context) {
		c.Set("debug", debug)
	})

	e.NoRoute(func(c *gin.Context) {
		c.Status(404)
	})

	namespace := e.Group("/api/v1/namespaces")
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

		subns := namespace.Group("/:namespace")
		{
			subns.Use(middleware.SetNamespace)

			deployment := subns.Group("/deployments")
			{
				deployment.GET("", server.ListDeployments)
				deployment.POST("", middleware.ParseJSON, server.CreateDeployment)
				deployment.GET("/:objname", middleware.SetObjectName, server.GetDeployment)
				deployment.DELETE("/:objname", middleware.SetObjectName, server.DeleteDeployment)
			}

			service := subns.Group("/services")
			{
				service.GET("", server.ListServices)
				service.POST("", middleware.ParseJSON, server.CreateService)
				service.GET("/:objname", middleware.SetObjectName, server.GetService)
				service.DELETE("/:objname", middleware.SetObjectName, server.DeleteService)
			}

			endpoints := subns.Group("/endpoints")
			{
				endpoints.GET("", server.ListEndpoints)
				endpoints.POST("", middleware.ParseJSON, server.CreateEndpoints)
				endpoints.GET("/:objname", middleware.SetObjectName, server.GetEndpoints)
				endpoints.DELETE("/:objname", middleware.SetObjectName, server.DeleteEndpoints)
			}

			configmaps := subns.Group("/configmaps")
			{
				configmaps.GET("", server.ListConfigMaps)
				configmaps.POST("", middleware.ParseJSON, server.CreateConfigMap)
				configmaps.GET("/:objname", middleware.SetObjectName, server.GetConfigMap)
				configmaps.DELETE("/:objname", middleware.SetObjectName, server.DeleteConfigMap)
			}

			secrets := subns.Group("/secrets")
			{
				secrets.GET("", server.ListSecrets)
				secrets.POST("", middleware.ParseJSON, server.CreateSecret)
				secrets.GET("/:objname", middleware.SetObjectName, server.GetSecret)
				secrets.DELETE("/:objname", middleware.SetObjectName, server.DeleteSecret)
			}
		}
	}

	return e
}
