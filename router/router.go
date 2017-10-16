package router

import (
	"net/http"

	"bitbucket.org/exonch/kube-api/access"
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
	e.Use(middleware.WriteResponseObject)    //order is alright
	e.Use(middleware.RedactResponseMetadata) //order is alright
	e.Use(middleware.SubstitutionsFromHeadersFor("responseObject", true))
	e.Use(middleware.SetRequestID)
	e.Use(utils.AddLogger)
	e.Use(middleware.CheckHTTP411)
	e.Use(middleware.ParseUserData)
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
			access.CheckAccess("Namespace", access.List),
			server.ListNamespaces,
		)
		namespace.POST("",
			middleware.ParseJSON,
			middleware.SubstitutionsFromHeadersFor("requestObject", false),
			access.CheckAccess("Namespace", access.Create),
			server.CreateNamespace,
		)
		namespace.GET("/:namespace",
			middleware.SetNamespace,
			access.CheckAccess("Namespace", access.Read),
			server.GetNamespace,
		)
		namespace.DELETE("/:namespace",
			middleware.SetNamespace,
			access.CheckAccess("Namespace", access.Delete),
			server.DeleteNamespace,
		)

		subns := namespace.Group("/:namespace")
		{
			subns.Use(middleware.SetNamespace)

			subns.GET("/resourcequotas/quota", server.GetNamespaceQuota)

			deployment := subns.Group("/deployments")
			{
				deployment.GET("",
					access.CheckAccess("Deployment", access.List),
					server.ListDeployments,
				)

				deployment.POST("",
					access.CheckAccess("Deployment", access.Create),
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					server.CreateDeployment,
				)

				deployment.GET("/:objname",
					access.CheckAccess("Deployment", access.Read),
					middleware.SetObjectName,
					server.GetDeployment,
				)

				deployment.DELETE("/:objname",
					access.CheckAccess("Deployment", access.Delete),
					middleware.SetObjectName,
					server.DeleteDeployment,
				)

				deployment.PUT("/:objname",
					access.CheckAccess("Deployment", access.Edit),
					middleware.ParseJSON,
					middleware.SetObjectName,
					server.ReplaceDeployment,
				)

				deployment.PATCH("/:objname",
					access.CheckAccess("Deployment", access.Edit),
					middleware.ParseJSON,
					middleware.SetObjectName,
					server.PatchDeployment,
				)

				deployment.PATCH("/:objname/image",
					access.CheckAccess("Deployment", access.Edit),
					middleware.ParseJSON,
					middleware.SetObjectName,
					server.GetDeployment,
					server.ChangeDeploymentImage,
				)

				deployment.PUT("/:objname/replicas",
					access.CheckAccess("Deployment", access.Edit),
					middleware.ParseJSON,
					middleware.SetObjectName,
					server.GetDeployment,
					server.ChangeDeploymentReplicas,
				)
			}

			service := subns.Group("/services")
			{
				service.GET("",
					access.CheckAccess("Service", access.List),
					server.ListServices,
				)

				service.POST("",
					access.CheckAccess("Service", access.Create),
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					server.CreateService,
				)

				service.GET("/:objname",
					access.CheckAccess("Service", access.Read),
					middleware.SetObjectName,
					server.GetService,
				)

				service.DELETE("/:objname",
					access.CheckAccess("Service", access.Delete),
					middleware.SetObjectName,
					server.DeleteService,
				)

				service.PUT("/:objname",
					access.CheckAccess("Service", access.Edit),
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					server.ReplaceService,
				)
			}

			endpoints := subns.Group("/endpoints")
			{
				endpoints.GET("",
					access.CheckAccess("Endpoints", access.List),
					server.ListEndpoints,
				)
				endpoints.POST("",
					access.CheckAccess("Endpoints", access.Create),
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					server.CreateEndpoints,
				)
				endpoints.GET("/:objname",
					access.CheckAccess("Endpoints", access.Read),
					middleware.SetObjectName,
					server.GetEndpoints,
				)
				endpoints.DELETE("/:objname",
					access.CheckAccess("Endpoints", access.Delete),
					middleware.SetObjectName,
					server.DeleteEndpoints,
				)
			}

			configmaps := subns.Group("/configmaps")
			{
				configmaps.GET("",
					access.CheckAccess("ConfigMap", access.List),
					server.ListConfigMaps,
				)
				configmaps.POST("",
					access.CheckAccess("ConfigMap", access.Create),
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					server.CreateConfigMap,
				)
				configmaps.GET("/:objname",
					access.CheckAccess("ConfigMap", access.Read),
					middleware.SetObjectName,
					server.GetConfigMap,
				)
				configmaps.DELETE("/:objname",
					access.CheckAccess("ConfigMap", access.Delete),
					middleware.SetObjectName,
					server.DeleteConfigMap,
				)
			}

			secrets := subns.Group("/secrets")
			{
				secrets.GET("",
					access.CheckAccess("Secret", access.Read),
					server.ListSecrets,
				)
				secrets.POST("",
					access.CheckAccess("Secret", access.Read),
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					server.CreateSecret,
				)
				secrets.GET("/:objname",
					access.CheckAccess("Secret", access.Read),
					middleware.SetObjectName,
					server.GetSecret,
				)
				secrets.DELETE("/:objname",
					access.CheckAccess("Secret", access.Read),
					middleware.SetObjectName,
					server.DeleteSecret,
				)
			}
		}
	}

	return e
}
