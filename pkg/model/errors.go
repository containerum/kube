package model

import (
	"fmt"
	"net/http"

	"gopkg.in/go-playground/validator.v8"

	"k8s.io/apimachinery/pkg/api/errors"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	alreadyExists       = "%s already exists in %s"
	fieldError          = "Validation failed for field: %s"
	fildNotFound        = "%s is not found"
	fieldShouldBeEmail  = "%v should be email address. Please, enter your valid email"
	fieldShouldExist    = "Field %v should be provided"
	fieldDefaultProblem = "%v should be %v"
)

var (
	ErrInvalidCPUFormat              = NewErrorWithCode("Invalid cpu quota format", http.StatusBadRequest)
	ErrInvalidMemoryFormat           = NewErrorWithCode("Invalid memory quota format", http.StatusBadRequest)
	ErrNoContainerInRequest          = NewErrorWithCode("No container in request", http.StatusNotFound)
	ErrUnableEncodeUserHeaderData    = NewErrorWithCode("Unbale to encode user header data", http.StatusInternalServerError)
	ErrUnableUnmarshalUserHeaderData = NewErrorWithCode("Unable unmarshal user header data", http.StatusInternalServerError)
	ErrUnableConvertServiceList      = NewErrorWithCode("unable decode service list", http.StatusInternalServerError)
	ErrUnableConvertService          = NewErrorWithCode("unable convert cubernetes service to user representation", http.StatusInternalServerError)
)

type Error struct {
	Text string `json:"error"`
	Code int    `json:"code,omitempty"`
}

func (e *Error) Error() string {
	if e.Code == 0 {
		return e.Text
	}
	return fmt.Sprintf("description: %s, code: %d", e.Text, e.Code)
}

func NewError(text string) *Error {
	return &Error{
		Text: text,
	}
}

func NewErrorWithCode(text string, code int) *Error {
	return &Error{
		Text: text,
		Code: code,
	}
}

//ParseBindErorrs parses different types of errors
func ParseErorrs(in interface{}) (code int, out []Error) {

	//Error from kubernetes
	sE, isStatusErrorCode := in.(*errors.StatusError)
	if isStatusErrorCode {
		switch sE.Status().Code {
		case 409:
			return http.StatusBadRequest, []Error{{Text: fmt.Sprintf(alreadyExists, sE.Status().Details.Name, sE.Status().Details.Kind)}}
		case 422:
			for _, c := range sE.Status().Details.Causes {
				switch c.Type {
				case api_meta.CauseTypeFieldValueNotFound:
					out = append(out, Error{Text: fmt.Sprintf(fildNotFound, c.Field)})
				case api_meta.CauseTypeFieldValueDuplicate:
					out = append(out, Error{Text: fmt.Sprintf(alreadyExists, sE.Status().Details.Name, sE.Status().Details.Kind)})
				default:
					out = append(out, Error{Text: fmt.Sprintf(fieldError, c.Field)})
				}
			}
			return http.StatusBadRequest, out
			//TODO Parse more errors
		case 0:
			return http.StatusInternalServerError, []Error{{Text: sE.Status().Message}}
		default:
			return int(sE.Status().Code), []Error{{Text: sE.Status().Message}}
		}
	}

	//Validation error
	vE, isValidationError := in.(validator.ValidationErrors)
	if isValidationError {
		for _, v := range vE {
			switch v.Tag {
			case "required":
				out = append(out, Error{Text: fmt.Sprintf(fieldShouldExist, v.Name)})
			case "email":
				out = append(out, Error{Text: fmt.Sprintf(fieldShouldBeEmail, v.Name)})
			default:
				out = append(out, Error{Text: fmt.Sprintf(fieldDefaultProblem, v.Name, v.Tag)})
			}
		}
		return http.StatusBadRequest, out
	}

	//Simple error with code
	mE, isErrorWithCode := in.(Error)
	if isErrorWithCode {
		return http.StatusInternalServerError, []Error{mE}
	}

	return http.StatusInternalServerError, []Error{{Text: in.(error).Error()}}
}
