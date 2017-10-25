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
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Deployment", access.Create),
					server.CreateDeployment,
				)

				deployment.GET("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Deployment", access.Read),
					server.GetDeployment,
				)

				deployment.DELETE("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Deployment", access.Delete),
					server.DeleteDeployment,
				)

				deployment.PUT("/:objname",
					middleware.SetObjectName,
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Deployment", access.Edit),
					server.ReplaceDeployment,
				)

				deployment.PATCH("/:objname",
					middleware.SetObjectName,
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Deployment", access.Edit),
					server.PatchDeployment,
				)

				deployment.PATCH("/:objname/image",
					middleware.SetObjectName,
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Deployment", access.SetImage),
					server.GetDeployment,
					server.ChangeDeploymentImage,
				)

				deployment.PUT("/:objname/replicas",
					middleware.SetObjectName,
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Deployment", access.SetReplicas),
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
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Service", access.Create),
					server.CreateService,
				)

				service.GET("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Service", access.Read),
					server.GetService,
				)

				service.DELETE("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Service", access.Delete),
					server.DeleteService,
				)

				service.PUT("/:objname",
					middleware.SetObjectName,
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Service", access.Edit),
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
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Endpoints", access.Create),
					server.CreateEndpoints,
				)
				endpoints.GET("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Endpoints", access.Read),
					server.GetEndpoints,
				)
				endpoints.DELETE("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Endpoints", access.Delete),
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
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("ConfigMap", access.Create),
					server.CreateConfigMap,
				)
				configmaps.GET("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("ConfigMap", access.Read),
					server.GetConfigMap,
				)
				configmaps.DELETE("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("ConfigMap", access.Delete),
					server.DeleteConfigMap,
				)
			}

			secrets := subns.Group("/secrets")
			{
				secrets.GET("",
					access.CheckAccess("Secret", access.List),
					server.ListSecrets,
				)
				secrets.POST("",
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Secret", access.Create),
					server.CreateSecret,
				)
				secrets.GET("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Secret", access.Read),
					server.GetSecret,
				)
				secrets.DELETE("/:objname",
					middleware.SetObjectName,
					access.CheckAccess("Secret", access.Delete),
					server.DeleteSecret,
				)
			}

			ingress := subns.Group("/ingress")
			{
				ingress.GET("",
					access.CheckAccess("Ingress", access.List),
					server.ListIngresses,
				)
				ingress.POST("",
					middleware.ParseJSON,
					middleware.SubstitutionsFromHeadersFor("requestObject", false),
					access.CheckAccess("Ingress", access.Create),
					server.CreateIngress,
				)
			}
		}
	}

	return e
}
