package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeerrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	h "git.containerum.net/ch/kube-api/pkg/router/handlers"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-api/static"
	"github.com/containerum/cherry/adaptors/cherrylog"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-contrib/cors"
	"github.com/sirupsen/logrus"

	"time"

	headers "github.com/containerum/utils/httputil"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
)

func CreateRouter(kube *kubernetes.Kube, enableCORS bool) http.Handler {
	e := gin.New()
	initMiddlewares(e, kube, enableCORS)
	initRoutes(e)
	return e
}

func initMiddlewares(e gin.IRouter, kube *kubernetes.Kube, enableCORS bool) {
	/* CORS */
	if enableCORS {
		cfg := cors.DefaultConfig()
		cfg.AllowAllOrigins = true
		cfg.AddAllowMethods(http.MethodDelete)
		cfg.AddAllowHeaders(headers.UserRoleXHeader, headers.UserIDXHeader, headers.UserNamespacesXHeader)
		e.Use(cors.New(cfg))
	}
	e.Group("/static").
		StaticFS("/", static.HTTP)
	/* System */
	e.Use(ginrus.Ginrus(logrus.WithField("component", "gin"), time.RFC3339, true))
	e.Use(gonic.Recovery(kubeerrors.ErrInternalError, cherrylog.NewLogrusAdapter(logrus.WithField("component", "gin"))))
	/* Custom */
	e.Use(headers.SaveHeaders)
	e.Use(headers.PrepareContext)
	e.Use(m.RequiredUserHeaders())
	e.Use(m.RegisterKubeClient(kube))
}

func initRoutes(e gin.IRouter) {
	e.GET("/ingresses", h.GetSelectedIngresses)
	e.GET("/configmaps", h.GetSelectedConfigMaps)
	e.GET("/storage", h.GetStorageList)

	namespace := e.Group("/namespaces")
	{
		namespace.GET("", h.GetNamespaceList)
		namespace.GET("/:namespace", m.ReadAccess, h.GetNamespace)
		namespace.POST("", h.CreateNamespace)
		namespace.PUT("/:namespace", h.UpdateNamespace)
		namespace.DELETE("/:namespace", h.DeleteNamespace)
		namespace.DELETE("", h.DeleteUserNamespaces)

		solutions := namespace.Group("/:namespace/solutions")
		{
			solutions.GET("/:solution/deployments", m.ReadAccess, h.GetDeploymentSolutionList)
			solutions.GET("/:solution/services", m.ReadAccess, h.GetServiceSolutionList)

			solutions.DELETE("/:solution/deployments", m.WriteAccess, h.DeleteDeploymentsSolution)
			solutions.DELETE("/:solution/services", m.WriteAccess, h.DeleteServicesSolution)
		}

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
			deployment.GET("/:deployment/pods", m.ReadAccess, h.GetDeploymentPodList)
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
			secret.POST("/tls", m.WriteAccess, h.CreateTLSSecret)
			secret.POST("/docker", m.WriteAccess, h.CreateDockerSecret)
			secret.PUT("/:secret", m.WriteAccess, h.UpdateSecret)
			secret.DELETE("/:secret", m.DeleteAccess, h.DeleteSecret)
		}

		ingress := namespace.Group("/:namespace/ingresses")
		{
			ingress.GET("", m.ReadAccess, h.GetIngressList)
			ingress.GET("/:ingress", m.ReadAccess, h.GetIngress)
			ingress.POST("", h.CreateIngress)
			ingress.PUT("/:ingress", h.UpdateIngress)
			ingress.DELETE("/:ingress", h.DeleteIngress)
		}

		endpoint := namespace.Group("/:namespace/endpoints", m.IsAdmin)
		{
			endpoint.GET("", h.GetEndpointList)
			endpoint.GET("/:endpoint", h.GetEndpoint)
			endpoint.POST("", h.CreateEndpoint)
			endpoint.PUT("/:endpoint", h.UpdateEndpoint)
			endpoint.DELETE("/:endpoint", h.DeleteEndpoint)
		}

		configmap := namespace.Group("/:namespace/configmaps")
		{
			configmap.GET("", m.ReadAccess, h.GetConfigMapList)
			configmap.GET("/:configmap", m.ReadAccess, h.GetConfigMap)
			configmap.POST("", m.WriteAccess, h.CreateConfigMap)
			configmap.PUT("/:configmap", m.WriteAccess, h.UpdateConfigMap)
			configmap.DELETE("/:configmap", m.DeleteAccess, h.DeleteConfigMap)
		}

		volume := namespace.Group("/:namespace/volumes")
		{
			volume.GET("", m.ReadAccess, h.GetVolumeList)
			volume.GET("/:volume", m.ReadAccess, h.GetVolume)
			volume.POST("", m.WriteAccess, h.CreateVolume)
			volume.PUT("/:volume", m.WriteAccess, h.UpdateVolume)
			volume.DELETE("/:volume", m.DeleteAccess, h.DeleteVolume)
		}

		pod := namespace.Group("/:namespace/pods")
		{
			pod.GET("", m.ReadAccess, h.GetPodList)
			pod.GET("/:pod", m.ReadAccess, h.GetPod)
			pod.GET("/:pod/log", m.ReadAccess, h.GetPodLogs)
			pod.DELETE("/:pod", m.DeleteAccess, h.DeletePod)
		}
	}
}
