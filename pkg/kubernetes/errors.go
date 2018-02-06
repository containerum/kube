package kubernetes

import (
	"errors"
)

var (
	ErrUnableGetNamespaceQuotaList = errors.New("Unable to get namespace quota list")
	ErrUnableGetNamespaceQuota     = errors.New("Unable to get namespace quota")
	ErrUnableGetNamespaceList      = errors.New("Unable to get namespace list")
	ErrUnableGetNamespace          = errors.New("Unable to get namespace")
	ErrUnableCreateNamespace       = errors.New("Unable to create namespace")
	ErrUnableCreateNamespaceQuota  = errors.New("Unable to create namespace quota")
	ErrUnableUpdateNamespaceQuota  = errors.New("Unable to update namespace quota")
	ErrUnableDeleteNamespace       = errors.New("Unable to delete namespace")

	ErrUnableGetServiceList = errors.New("Unable to get service list")
	ErrUnableCreateService  = errors.New("Unable to create service")

	ErrUnableGetDeploymentList = errors.New("Unable to get deployment list")
	ErrUnableGetDeployment     = errors.New("Unable to get deployment")

	ErrUnableGetPodList = errors.New("Unable to get pod list")
	ErrUnableGetPod     = errors.New("Unable to get pod")

	ErrUnableGetService = errors.New("Unable to get service")
)
