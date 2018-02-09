package model

import (
	"fmt"
	"net/http"
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
	Text string
	Code int
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
