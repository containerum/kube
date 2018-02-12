package router

const (
	userIDNotProvided    = "UserID not provided"
	userIDHeaderRequired = "X-User-ID header required"

	containerNotFoundError      = "Container %s is not found in deployment %s"
	invalidUpdateDeploymentName = "Deployment name in URI %s does not match deployment name in deployment %s"
)
