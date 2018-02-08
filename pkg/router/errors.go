package router

import (
	"fmt"

	"gopkg.in/go-playground/validator.v8"
)

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

//BindError is a type for bind errors
type BindError struct {
	Error string `json:"error"`
}

//ParseBindErorrs parses errors from message content binding
func ParseErorrs(in error) []BindError {
	var out []BindError

	t, isValidationError := in.(validator.ValidationErrors)

	if isValidationError {
		for _, v := range t {
			switch v.Tag {
			case "required":
				out = append(out, BindError{fmt.Sprintf(fieldShouldExist, v.Name)})
			case "email":
				out = append(out, BindError{fmt.Sprintf(fieldShouldBeEmail, v.Name)})
			default:
				out = append(out, BindError{fmt.Sprintf(fieldDefaultProblem, v.Name, v.Tag)})
			}
		}
		return out
	}
	return []BindError{{Error: in.Error()}}
}
