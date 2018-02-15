package router

import (
	"net/http"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
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
		c.Status(404)
	})
	namespace := e.Group("/namespaces")
	{
		namespace.Use(m.IsAdmin()).GET("", getNamespaceList)
		namespace.GET("/:namespace", getNamespace)
		namespace.POST("", сreateNamespace)
		namespace.PUT("/:namespace", updateNamespace)
		namespace.DELETE("/:namespace", deleteNamespace)

		service := namespace.Group("/:namespace/services")
		{
			service.Use(m.ReadAccess()).GET("", getServiceList)
			service.Use(m.ReadAccess()).GET("/:service", getService)
			service.POST("", createService)
			service.PUT("/:service", updateService)
			service.DELETE("/:service", deleteService)
		}

		deployment := namespace.Group("/:namespace/deployments")
		{
			deployment.Use(m.ReadAccess()).GET("", getDeploymentList)
			deployment.Use(m.ReadAccess()).GET("/:deployment", getDeployment)
			deployment.POST("", createDeployment)
			deployment.PUT("/:deployment", updateDeployment)
			deployment.PUT("/:deployment/replicas", updateDeploymentReplicas)
			deployment.PUT("/:deployment/image", updateDeploymentImage)
			deployment.DELETE("/:deployment", deleteDeployment)
		}

		secret := namespace.Group("/:namespace/secrets")
		{
			secret.Use(m.ReadAccess()).GET("", getSecretList)
			secret.Use(m.ReadAccess()).GET("/:secret", getSecret)
			secret.POST("", createSecret)
			secret.PUT("/:secret", updateSecret)
			secret.DELETE("/:secret", deleteSecret)
		}

		ingress := namespace.Group("/:namespace/ingresses")
		{
			ingress.Use(m.IsAdmin()).GET("", getIngressList)
			ingress.GET("/:ingress", getIngress)
			ingress.POST("", createIngress)
			ingress.PUT("/:ingress", updateIngress)
			ingress.DELETE("/:ingress", deleteIngress)
		}

		endpoint := namespace.Group("/:namespace/endpoints")
		{
			endpoint.Use(m.IsAdmin()).GET("", getEndpointList)
			endpoint.GET("/:endpoint", getEndpoint)
			endpoint.POST("", createEndpoint)
			endpoint.PUT("/:endpoint", updateEndpoint)
			endpoint.DELETE("/:endpoint", deleteEndpoint)
		}

		pod := namespace.Group("/:namespace/pods")
		{
			pod.Use(m.ReadAccess()).GET("", getPodList)
			pod.Use(m.ReadAccess()).GET("/:pod", getPod)
			pod.GET("/:pod/log", getPodLogs)
			pod.DELETE("/:pod", deletePod)
		}
	}
}
