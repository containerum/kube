package kubernetes

import (
	"errors"
)

var (
	ErrUnableGetNamespaceList     = errors.New("Unable to get namespaces list")
	ErrUnableGetNamespace         = errors.New("Unable to get namespace")
	ErrUnableCreateNamespace      = errors.New("Unable to create namespace")
	ErrUnableCreateNamespaceQuota = errors.New("Unable to create namespace quota")
	ErrUnableUpdateNamespaceQuota = errors.New("Unable to update namespace quota")
	ErrUnableDeleteNamespace      = errors.New("Unable to delete namespace")

	ErrUnableGetServiceList = errors.New("Unable to get service list")
	ErrUnableGetService     = errors.New("Unable to get service")
	ErrUnableCreateService  = errors.New("Unable to create service")
	ErrUnableUpdateService  = errors.New("Unable update service")
	ErrUnableDeleteService  = errors.New("Unable to delete service")

	ErrUnableGetDeploymentList = errors.New("Unable to get deployments list")
	ErrUnableGetDeployment     = errors.New("Unable to get deployment")
	ErrUnableCreateDeployment  = errors.New("Unable to create deployment")
	ErrUnableUpdateDeployment  = errors.New("Unable to update deployment")
	ErrUnableDeleteDeployment  = errors.New("Unable to delete deployment")

	ErrUnableGetIngressList = errors.New("Unable to get deployments list")
	ErrUnableGetIngress     = errors.New("Unable to get deployment")
	ErrUnableCreateIngress  = errors.New("Unable to create deployment")
	ErrUnableUpdateIngress  = errors.New("Unable to update deployment")
	ErrUnableDeleteIngress  = errors.New("Unable to delete deployment")

	ErrUnableGetSecretList = errors.New("Unable to get secrets list")
	ErrUnableGetSecret     = errors.New("Unable to get secret")
	ErrUnableCreateSecret  = errors.New("Unable to create secret")
	ErrUnableUpdateSecret  = errors.New("Unable to update secret")
	ErrUnableDeleteSecret  = errors.New("Unable to delete secret")

	ErrUnableGetEndpointList = errors.New("Unable to get endpoints list")
	ErrUnableGetEndpoint     = errors.New("Unable to get endpoint")
	ErrUnableCreateEndpoint  = errors.New("Unable to create endpoint")
	ErrUnableDeleteEndpoint  = errors.New("Unable to delete endpoint")

	ErrUnableGetPodList = errors.New("Unable to get pod list")
	ErrUnableGetPod     = errors.New("Unable to get pod")
	ErrUnableDeletePod  = errors.New("Unable to delete pod")
)
