package router

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	h "git.containerum.net/ch/kube-api/pkg/router/handlers"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-api/static"
	"github.com/containerum/cherry/adaptors/cherrylog"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/containerum/utils/httputil"
	"github.com/gin-contrib/cors"
	"github.com/sirupsen/logrus"

	"time"

	"github.com/containerum/kube-client/pkg/model"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
)

func CreateRouter(kube *kubernetes.Kube, access httputil.AccessChecker, enableCORS bool) http.Handler {
	e := gin.New()
	initMiddlewares(e, kube, enableCORS)
	initRoutes(e, access)
	return e
}

func initMiddlewares(e gin.IRouter, kube *kubernetes.Kube, enableCORS bool) {
	/* CORS */
	if enableCORS {
		cfg := cors.DefaultConfig()
		cfg.AllowAllOrigins = true
		cfg.AddAllowMethods(http.MethodDelete)
		cfg.AddAllowHeaders(httputil.UserRoleXHeader, httputil.UserIDXHeader, httputil.UserNamespacesXHeader)
		e.Use(cors.New(cfg))
	}
	e.Group("/static").
		StaticFS("/", static.HTTP)
	/* System */
	e.Use(ginrus.Ginrus(logrus.WithField("component", "gin"), time.RFC3339, true))
	e.Use(gonic.Recovery(kubeErrors.ErrInternalError, cherrylog.NewLogrusAdapter(logrus.WithField("component", "gin"))))
	/* Custom */
	e.Use(httputil.SaveHeaders)
	e.Use(httputil.PrepareContext)
	e.Use(m.RegisterKubeClient(kube))
}

func initRoutes(e gin.IRouter, access httputil.AccessChecker) {
	e.GET("/ingresses", h.GetSelectedIngresses)
	e.GET("/configmaps", h.GetSelectedConfigMaps)

	namespace := e.Group("/projects/:project/namespaces")
	{
		namespace.GET("", h.GetNamespaceList)
		namespace.GET("/:namespace", access.CheckAccess(model.AccessGuest), h.GetNamespace)
		namespace.POST("", h.CreateNamespace)
		namespace.PUT("/:namespace", h.UpdateNamespace)
		namespace.DELETE("/:namespace", h.DeleteNamespace)

		solutions := namespace.Group("/:namespace/solutions")
		{
			solutions.GET("/:solution/deployments", access.CheckAccess(model.AccessGuest), h.GetDeploymentSolutionList)
			solutions.GET("/:solution/services", access.CheckAccess(model.AccessGuest), h.GetServiceSolutionList)

			solutions.DELETE("/:solution/deployments", access.CheckAccess(model.AccessMember), h.DeleteDeploymentsSolution)
			solutions.DELETE("/:solution/services", access.CheckAccess(model.AccessMember), h.DeleteServicesSolution)
		}

		service := namespace.Group("/:namespace/services")
		{
			service.GET("", access.CheckAccess(model.AccessGuest), h.GetServiceList)
			service.GET("/:service", access.CheckAccess(model.AccessGuest), h.GetService)
			service.POST("", h.CreateService)
			service.PUT("/:service", h.UpdateService)
			service.DELETE("/:service", h.DeleteService)
		}

		deployment := namespace.Group("/:namespace/deployments")
		{
			deployment.GET("", access.CheckAccess(model.AccessGuest), h.GetDeploymentList)
			deployment.GET("/:deployment", access.CheckAccess(model.AccessGuest), h.GetDeployment)
			deployment.POST("", h.CreateDeployment)
			deployment.PUT("/:deployment", h.UpdateDeployment)
			deployment.PUT("/:deployment/replicas", h.UpdateDeploymentReplicas)
			deployment.PUT("/:deployment/image", h.UpdateDeploymentImage)
			deployment.DELETE("/:deployment", h.DeleteDeployment)
		}

		secret := namespace.Group("/:namespace/secrets")
		{
			secret.GET("", access.CheckAccess(model.AccessAdmin), h.GetSecretList)
			secret.GET("/:secret", access.CheckAccess(model.AccessAdmin), h.GetSecret)
			secret.POST("", access.CheckAccess(model.AccessAdmin), h.CreateSecret)
			secret.PUT("/:secret", access.CheckAccess(model.AccessAdmin), h.UpdateSecret)
			secret.DELETE("/:secret", access.CheckAccess(model.AccessAdmin), h.DeleteSecret)
		}

		ingress := namespace.Group("/:namespace/ingresses")
		{
			ingress.GET("", access.CheckAccess(model.AccessGuest), h.GetIngressList)
			ingress.GET("/:ingress", access.CheckAccess(model.AccessGuest), h.GetIngress)
			ingress.POST("", h.CreateIngress)
			ingress.PUT("/:ingress", h.UpdateIngress)
			ingress.DELETE("/:ingress", h.DeleteIngress)
		}

		configmap := namespace.Group("/:namespace/configmaps")
		{
			configmap.GET("", access.CheckAccess(model.AccessGuest), h.GetConfigMapList)
			configmap.GET("/:configmap", access.CheckAccess(model.AccessGuest), h.GetConfigMap)
			configmap.POST("", access.CheckAccess(model.AccessMaster), h.CreateConfigMap)
			configmap.PUT("/:configmap", access.CheckAccess(model.AccessMaster), h.UpdateConfigMap)
			configmap.DELETE("/:configmap", access.CheckAccess(model.AccessMaster), h.DeleteConfigMap)
		}

		volume := namespace.Group("/:namespace/volumes")
		{
			volume.GET("", h.GetVolumeList)
			volume.GET("/:volume", h.GetVolume)
			volume.POST("", h.CreateVolume)
			volume.PUT("/:volume", h.UpdateVolume)
			volume.DELETE("/:volume", h.DeleteVolume)
		}

		pod := namespace.Group("/:namespace/pods")
		{
			pod.GET("", access.CheckAccess(model.AccessGuest), h.GetPodList)
			pod.GET("/:pod", access.CheckAccess(model.AccessGuest), h.GetPod)
			pod.GET("/:pod/log", access.CheckAccess(model.AccessMember), h.GetPodLogs)
			pod.DELETE("/:pod", access.CheckAccess(model.AccessMaster), h.DeletePod)
		}

		//Unused at this moment
		endpoint := namespace.Group("/:namespace/endpoints")
		{
			endpoint.GET("", h.GetEndpointList)
			endpoint.GET("/:endpoint", h.GetEndpoint)
			endpoint.POST("", h.CreateEndpoint)
			endpoint.PUT("/:endpoint", h.UpdateEndpoint)
			endpoint.DELETE("/:endpoint", h.DeleteEndpoint)
		}
	}
}
