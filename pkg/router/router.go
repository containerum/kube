package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

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
	e.Use(gin.Logger())
	e.Use(gin.Recovery())
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
		namespace.Use(m.IsAdmin()).GET("/:namespace", getNamespace)
	}
}
