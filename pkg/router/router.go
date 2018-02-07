package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
		namespace.Use(m.IsAdmin()).GET("", getNamespaceList)
		namespace.Use(m.IsAdmin()).POST("", сreateNamespace)
		namespace.Use(m.IsAdmin()).GET("/:namespace", getNamespace)
		namespace.Use(m.IsAdmin()).DELETE("/:namespace", deleteNamespace)
		namespace.Use(m.IsAdmin()).PUT("/:namespace", updateNamespace)

		service := namespace.Group("/:namespace/services")
		{
			service.GET("", getServiceList)
			service.POST("/", createService)
			service.GET("/:service", getService)
			service.Use(m.IsAdmin()).PUT("/:service", updateService)
		}

		deployment := namespace.Group("/:namespace/deployments")
		{
			deployment.GET("", getDeploymentList)
			deployment.GET("/:deployment", getDeployment)
			deployment.POST("", createDeployment)
			deployment.PUT("/:deployment", updateDeployment)
			deployment.PUT("/:deployment/replicas", updateDeploymentReplicas)
			deployment.PUT("/:deployment/image", updateDeploymentImage)
			deployment.DELETE("/:deployment", deleteDeployment)
		}
		pod := namespace.Group("/:namespace/pods")
		{
			pod.GET("", getPodList)
			pod.GET("/:pod", getPod)
			pod.GET("/:pod/log", getPodLogs)
		}
	}
}
