package router

const (
	fieldShouldBeEmail  = "%v should be email address. Please, enter your valid email"
	fieldShouldExist    = "Field %v should be provided"
	fieldDefaultProblem = "%v should be %v"
)

const (
	invalidCPUFormat            = "Invalid cpu quota format: %s"
	invalidMemoryFormat         = "Invalid memory quota format: %s"
	namespaceCreationError      = "Namespace %s creation error: %s"
	namespaceQuotaCreationError = "Namespace %s quota creation error: %s"
	namespaceNotMatchError      = "Namespace %s does not match namespace %s in deployment"
	serviceCreationError        = "Service %s creation error: %s"
	deploymentCreationError     = "Deployment %s creation error: %s"
	deploymentUpdateError       = "Deployment %s update error: %s"
	invalidUpdateDeploymentName = "Deployment name in URI (%s) does not match deployment name in deployment (%s)"
	containerNotFoundError      = "Container %s is not found in deployment %s"
)
