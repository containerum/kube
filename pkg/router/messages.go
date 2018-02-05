package router

const (
	fieldShouldBeEmail  = "%v should be email address. Please, enter your valid email"
	fieldShouldExist    = "Field %v should be provided"
	fieldDefaultProblem = "%v should be %v"
)

const (
	requestNotService           = "Request is not a Service request"
	requestNotNamespace         = "Request is not a Namespace request"
	invalidCPUFormat            = "Invalid cpu quota format: %v"
	invalidMemoryFormat         = "Invalid memory quota format: %v"
	namespaceCreationError      = "Namespace %v creation error: %v"
	namespaceQuotaCreationError = "Namespace %v quota creation error: %v"
	namespaceNotMatchError      = "Namespace %v does not match namespace %v in deployment"
	serviceCreationError        = "Service %v creation error: %v"
)
