package router

import (
	"fmt"

	"net/http"

	"strings"

	"git.containerum.net/ch/kube-api/pkg/model"
	"gopkg.in/go-playground/validator.v8"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	fieldShouldBeEmail  = "%v should be email address. Please, enter your valid email"
	fieldShouldExist    = "Field %v should be provided"
	fieldDefaultProblem = "%v should be %v"
)

const (
	userIDNotProvided    = "UserID not provided"
	userIDHeaderRequired = "X-User-ID header required"
	invalidCPUFormat     = "Invalid cpu quota format"
	invalidMemoryFormat  = "Invalid memory quota format"
	alreadyExists        = "%s already exists in %s"
	fieldError           = "Validation failed for fields: %v"

	containerNotFoundError      = "Container %s is not found in deployment %s"
	invalidUpdateDeploymentName = "Deployment name in URI %s does not match deployment name in deployment %s"
)

//KubeError is a type for bind errors
type KubeError struct {
	Error string `json:"error"`
}

//ParseBindErorrs parses different types of errors
func ParseErorrs(in interface{}) (code int, out []KubeError) {

	//Error from kubernetes
	sE, isStatusErrorCode := in.(*errors.StatusError)
	if isStatusErrorCode {
		switch sE.Status().Code {
		case 409:
			return http.StatusBadRequest, []KubeError{{Error: fmt.Sprintf(alreadyExists, sE.Status().Details.Name, sE.Status().Details.Kind)}}
		case 422:
			var causes []string
			for _, c := range sE.Status().Details.Causes {
				causes = append(causes, c.Field)
			}
			return http.StatusBadRequest, []KubeError{{Error: fmt.Sprintf(fieldError, strings.Join(causes, ", "))}}
			//TODO Parse more errors
		case 0:
			return http.StatusInternalServerError, []KubeError{{Error: sE.Status().Message}}
		default:
			return int(sE.Status().Code), []KubeError{{Error: sE.Status().Message}}
		}
	}

	//Simple error with code
	mE, isErrorWithCode := in.(*model.Error)
	if isErrorWithCode {
		if mE.Code != 0 {
			return mE.Code, []KubeError{{Error: mE.Text}}
		}
		return http.StatusInternalServerError, []KubeError{{Error: mE.Text}}
	}

	//Validation error
	vE, isValidationError := in.(validator.ValidationErrors)
	if isValidationError {
		for _, v := range vE {
			switch v.Tag {
			case "required":
				out = append(out, KubeError{fmt.Sprintf(fieldShouldExist, v.Name)})
			case "email":
				out = append(out, KubeError{fmt.Sprintf(fieldShouldBeEmail, v.Name)})
			default:
				out = append(out, KubeError{fmt.Sprintf(fieldDefaultProblem, v.Name, v.Tag)})
			}
		}
		return http.StatusBadRequest, out
	}
	return http.StatusInternalServerError, []KubeError{{Error: in.(error).Error()}}
}
