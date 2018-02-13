package router

const (
	userIDNotProvided    = "UserID not provided"
	userIDHeaderRequired = "X-User-ID header required"

	containerNotFoundError      = "Container %s is not found in deployment %s"
	invalidUpdateDeploymentName = "Deployment name in URI %s does not match deployment name in request %s"
	invalidUpdateIngressName    = "Ingress name in URI %s does not match ingress name in request %s"
	invalidUpdateSecretName     = "Secret name in URI %s does not match secret name in request %s"
	invalidUpdateServiceName    = "Service name in URI %s does not match service name in request %s"
)
