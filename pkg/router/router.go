package router

import (
	"net/http"

	"git.containerum.net/ch/cherry/adaptors/cherrylog"
	"git.containerum.net/ch/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	h "git.containerum.net/ch/kube-api/pkg/router/handlers"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-api/static"
	"github.com/gin-contrib/cors"
	"github.com/sirupsen/logrus"

	"time"

	"git.containerum.net/ch/api-gateway/pkg/utils/headers"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
)

func CreateRouter(kube *kubernetes.Kube) http.Handler {
	e := gin.New()
	initMiddlewares(e, kube)
	initRoutes(e)
	return e
}

func initMiddlewares(e gin.IRouter, kube *kubernetes.Kube) {
	/* CORS */
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AddAllowMethods("DELETE")
	cfg.AddAllowHeaders(headers.UserRoleXHeader, headers.UserIDXHeader, headers.UserNamespacesXHeader, headers.UserVolumesXHeader)
	e.Use(cors.New(cfg))
	e.Group("/static").
		StaticFS("/", static.HTTP)
	/* System */
	e.Use(ginrus.Ginrus(logrus.WithField("component", "gin"), time.RFC3339, true))
	e.Use(gonic.Recovery(cherry.ErrInternalError, cherrylog.NewLogrusAdapter(logrus.WithField("component", "gin"))))
	/* Custom */
	e.Use(m.RequiredUserHeaders())
	e.Use(m.SetNamespace())
	e.Use(m.RegisterKubeClient(kube))
}

func initRoutes(e gin.IRouter) {
	e.GET("/ingresses", h.GetSelectedIngresses)
	e.GET("/configmaps", h.GetSelectedConfigMaps)

	namespace := e.Group("/namespaces")
	{
		namespace.GET("", h.GetNamespaceList)
		namespace.GET("/:namespace", m.ReadAccess, h.GetNamespace)
		namespace.POST("", h.CreateNamespace)
		namespace.PUT("/:namespace", h.UpdateNamespace)
		namespace.DELETE("/:namespace", h.DeleteNamespace)

		service := namespace.Group("/:namespace/services")
		{
			service.GET("", m.ReadAccess, h.GetServiceList)
			service.GET("/:service", m.ReadAccess, h.GetService)
			service.POST("", h.CreateService)
			service.PUT("/:service", h.UpdateService)
			service.DELETE("/:service", h.DeleteService)
		}

		deployment := namespace.Group("/:namespace/deployments")
		{
			deployment.GET("", m.ReadAccess, h.GetDeploymentList)
			deployment.GET("/:deployment", m.ReadAccess, h.GetDeployment)
			deployment.POST("", h.CreateDeployment)
			deployment.PUT("/:deployment", h.UpdateDeployment)
			deployment.PUT("/:deployment/replicas", h.UpdateDeploymentReplicas)
			deployment.PUT("/:deployment/image", h.UpdateDeploymentImage)
			deployment.DELETE("/:deployment", h.DeleteDeployment)
		}

		secret := namespace.Group("/:namespace/secrets")
		{
			secret.GET("", m.ReadAccess, h.GetSecretList)
			secret.GET("/:secret", m.ReadAccess, h.GetSecret)
			secret.POST("", m.ReadAccess, h.CreateSecret)
			secret.PUT("/:secret", m.ReadAccess, h.UpdateSecret)
			secret.DELETE("/:secret", m.ReadAccess, h.DeleteSecret)
		}

		ingress := namespace.Group("/:namespace/ingresses")
		{
			ingress.GET("", m.ReadAccess, h.GetIngressList)
			ingress.GET("/:ingress", m.ReadAccess, h.GetIngress)
			ingress.POST("", h.CreateIngress)
			ingress.PUT("/:ingress", h.UpdateIngress)
			ingress.DELETE("/:ingress", h.DeleteIngress)
		}

		endpoint := namespace.Group("/:namespace/endpoints")
		{
			endpoint.GET("", m.IsAdmin, h.GetEndpointList)
			endpoint.GET("/:endpoint", m.IsAdmin, h.GetEndpoint)
			endpoint.POST("", h.CreateEndpoint)
			endpoint.PUT("/:endpoint", h.UpdateEndpoint)
			endpoint.DELETE("/:endpoint", h.DeleteEndpoint)
		}

		configmap := namespace.Group("/:namespace/configmaps")
		{
			configmap.GET("", m.ReadAccess, h.GetConfigMapList)
			configmap.GET("/:configmap", m.ReadAccess, h.GetConfigMap)
			configmap.POST("", m.ReadAccess, h.CreateConfigMap)
			configmap.PUT("/:configmap", m.ReadAccess, h.UpdateConfigMap)
			configmap.DELETE("/:configmap", m.ReadAccess, h.DeleteConfigMap)
		}

		pod := namespace.Group("/:namespace/pods")
		{
			pod.GET("", m.ReadAccess, h.GetPodList)
			pod.GET("/:pod", m.ReadAccess, h.GetPod)
			pod.GET("/:pod/log", m.ReadAccess, h.GetPodLogs)
			pod.DELETE("/:pod", m.ReadAccess, h.DeletePod)
		}
	}
}
