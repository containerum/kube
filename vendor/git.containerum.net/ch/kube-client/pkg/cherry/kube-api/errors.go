package kubeerr

import (
	"net/http"

	"git.containerum.net/ch/kube-client/pkg/cherry"
)

var kubeApiErr = cherry.BuildErr(2)

//Errors
func ErrAdminRequired() *cherry.Err {
	return kubeApiErr("Admin access required", http.StatusForbidden, 1)
}
func ErrRequiredHeadersNotProvided() *cherry.Err {
	return kubeApiErr("Required headers not provided", http.StatusForbidden, 2)
}
func ErrRequestValidationFailed() *cherry.Err {
	return kubeApiErr("Request validation failed", http.StatusBadRequest, 3)
}

func ErrUnableGetResourcesList() *cherry.Err {
	return kubeApiErr("Unable to get resources list", http.StatusInternalServerError, 4)
}
func ErrUnableGetResource() *cherry.Err {
	return kubeApiErr("Unable to get resource", http.StatusInternalServerError, 5)
}
func ErrUnableCreateResource() *cherry.Err {
	return kubeApiErr("Unable to create resource", http.StatusInternalServerError, 6)
}
func ErrUnableUpdateResource() *cherry.Err {
	return kubeApiErr("Unable to update resource", http.StatusInternalServerError, 7)
}
func ErrUnableDeleteResource() *cherry.Err {
	return kubeApiErr("Unable to delete resource", http.StatusInternalServerError, 8)
}
func ErrResourceAlreadyExists() *cherry.Err {
	return kubeApiErr("Resource with this name already exists", http.StatusConflict, 9)
}
func ErrResourceNotExist() *cherry.Err {
	return kubeApiErr("Resource with this name doesn't exist", http.StatusNotFound, 10)
}
