package router

import (
	"net/http"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	h "git.containerum.net/ch/kube-api/pkg/router/handlers"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func CreateRouter(kube *kubernetes.Kube, debug bool) http.Handler {
	e := gin.New()
	initMiddlewares(e, kube)
	initRoutes(e)
	return e
}

func initMiddlewares(e *gin.Engine, kube *kubernetes.Kube) {
	/* System */
	e.Use(ginrus.Ginrus(logrus.WithField("component", "gin"), time.RFC3339, true))
	e.Use(gin.RecoveryWithWriter(logrus.WithField("component", "gin_recovery").WriterLevel(logrus.ErrorLevel)))
	/* Custom */
	e.Use(m.RequiredUserHeaders())
	e.Use(m.SetNamespace())
	e.Use(m.RegisterKubeClient(kube))
}

func initRoutes(e *gin.Engine) {
	e.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	namespace := e.Group("/namespaces")
	{
		namespace.Use(m.IsAdmin()).GET("", h.GetNamespaceList)
		namespace.GET("/:namespace", h.GetNamespace)
		namespace.POST("", h.CreateNamespace)
		namespace.PUT("/:namespace", h.UpdateNamespace)
		namespace.DELETE("/:namespace", h.DeleteNamespace)

		service := namespace.Group("/:namespace/services")
		{
			service.Use(m.ReadAccess()).GET("", h.GetServiceList)
			service.Use(m.ReadAccess()).GET("/:service", h.GetService)
			service.POST("", h.CreateService)
			service.PUT("/:service", h.UpdateService)
			service.DELETE("/:service", h.DeleteService)
		}

		deployment := namespace.Group("/:namespace/deployments")
		{
			deployment.Use(m.ReadAccess()).GET("", h.GetDeploymentList)
			deployment.Use(m.ReadAccess()).GET("/:deployment", h.GetDeployment)
			deployment.POST("", h.CreateDeployment)
			deployment.PUT("/:deployment", h.UpdateDeployment)
			deployment.PUT("/:deployment/replicas", h.UpdateDeploymentReplicas)
			deployment.PUT("/:deployment/image", h.UpdateDeploymentImage)
			deployment.DELETE("/:deployment", h.DeleteDeployment)
		}

		secret := namespace.Group("/:namespace/secrets")
		{
			secret.Use(m.ReadAccess()).GET("", h.GetSecretList)
			secret.GET("/:secret", h.GetSecret)
			secret.POST("", h.CreateSecret)
			secret.PUT("/:secret", h.UpdateSecret)
			secret.DELETE("/:secret", h.DeleteSecret)
		}

		ingress := namespace.Group("/:namespace/ingresses")
		{
			ingress.Use(m.IsAdmin()).GET("", h.GetIngressList)
			ingress.GET("/:ingress", h.GetIngress)
			ingress.POST("", h.CreateIngress)
			ingress.PUT("/:ingress", h.UpdateIngress)
			ingress.DELETE("/:ingress", h.DeleteIngress)
		}

		endpoint := namespace.Group("/:namespace/endpoints")
		{
			endpoint.Use(m.IsAdmin()).GET("", h.GetEndpointList)
			endpoint.GET("/:endpoint", h.GetEndpoint)
			endpoint.POST("", h.CreateEndpoint)
			endpoint.PUT("/:endpoint", h.UpdateEndpoint)
			endpoint.DELETE("/:endpoint", h.DeleteEndpoint)
		}

		configmap := namespace.Group("/:namespace/configmaps")
		{
			configmap.Use(m.IsAdmin()).GET("", h.GetConfigMapList)
			configmap.GET("/:configmap", h.GetConfigMap)
			configmap.POST("", h.CreateConfigMap)
			configmap.PUT("/:configmap", h.UpdateConfigMap)
			configmap.DELETE("/:configmap", h.DeleteConfigMap)
		}

		pod := namespace.Group("/:namespace/pods")
		{
			pod.Use(m.ReadAccess()).GET("", h.GetPodList)
			pod.Use(m.ReadAccess()).GET("/:pod", h.GetPod)
			pod.GET("/:pod/log", h.GetPodLogs)
			pod.DELETE("/:pod", h.DeletePod)
		}
	}
}
