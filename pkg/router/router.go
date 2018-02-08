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
	e.Use(m.RequiredHeaders())
	e.Use(m.RegisterKubeClient(kube))
}

func initRoutes(e *gin.Engine) {
	e.NoRoute(func(c *gin.Context) {
		c.Status(404)
	})
	namespace := e.Group("/namespaces")
	{
		namespace.GET("", getNamespaceList)
		namespace.GET("/:namespace", getNamespace)
		namespace.Use(m.IsAdmin()).POST("", сreateNamespace)
		namespace.Use(m.IsAdmin()).PUT("/:namespace", updateNamespace)
		namespace.Use(m.IsAdmin()).DELETE("/:namespace", deleteNamespace)

		service := namespace.Group("/:namespace/services")
		{
			service.GET("", getServiceList)
			service.GET("/:service", getService)
			service.Use(m.IsAdmin()).POST("", createService)
			service.Use(m.IsAdmin()).PUT("/:service", updateService)
			service.Use(m.IsAdmin()).DELETE("/:service", deleteService)
		}

		deployment := namespace.Group("/:namespace/deployments")
		{
			deployment.GET("", getDeploymentList)
			deployment.GET("/:deployment", getDeployment)
			deployment.Use(m.IsAdmin()).POST("", createDeployment)
			deployment.Use(m.IsAdmin()).PUT("/:deployment", updateDeployment)
			deployment.Use(m.IsAdmin()).PUT("/:deployment/replicas", updateDeploymentReplicas)
			deployment.Use(m.IsAdmin()).PUT("/:deployment/image", updateDeploymentImage)
			deployment.Use(m.IsAdmin()).DELETE("/:deployment", deleteDeployment)
		}

		secret := namespace.Group("/:namespace/secrets")
		{
			secret.GET("", getSecretList)
			secret.GET("/:secret", getSecret)
			secret.Use(m.IsAdmin()).POST("", createSecret)
			secret.Use(m.IsAdmin()).DELETE("/:secret", deleteSecret)
		}

		pod := namespace.Group("/:namespace/pods")
		{
			pod.GET("", getPodList)
			pod.GET("/:pod", getPod)
			pod.GET("/:pod/log", getPodLogs)
		}
	}
}
