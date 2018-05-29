package model

import (
	"errors"

	"fmt"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"github.com/containerum/cherry"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrUnableDecodeUserHeaderData    = errors.New("unable to decode user header data")
	ErrUnableUnmarshalUserHeaderData = errors.New("unable to unmarshal user header data")

	ErrUnableConvertServiceList = errors.New("unable to decode services list")
	ErrUnableConvertService     = errors.New("unable to decode service")

	ErrUnableConvertVolumeList = errors.New("unable to decode volumes list")
	ErrUnableConvertVolume     = errors.New("unable to decode volume")

	ErrUnableConvertNamespaceList = errors.New("unable to decode namespaces list")
	ErrUnableConvertNamespace     = errors.New("unable to decode namespace")

	ErrUnableConvertSecretList = errors.New("unable to decode secrets list")
	ErrUnableConvertSecret     = errors.New("unable to decode secret")

	ErrUnableConvertIngressList = errors.New("unable to decode ingresses list")
	ErrUnableConvertIngress     = errors.New("unable to decode ingress")

	ErrUnableConvertDeploymentList = errors.New("unable to decode deployment list")
	ErrUnableConvertDeployment     = errors.New("unable to decode deployment")

	ErrUnableConvertEndpointList = errors.New("unable to decode services list")
	ErrUnableConvertEndpoint     = errors.New("unable to decode service")

	ErrUnableConvertConfigMapList = errors.New("unable to decode config maps list")
	ErrUnableConvertConfigMap     = errors.New("unable to decode config map")
)

const (
	noContainer           = "container '%v' is not found in deployment"
	fieldShouldExist      = "field '%v' should be provided"
	invalidReplicas       = "invalid replicas number: %v. It must be between 1 and %v"
	invalidPort           = "invalid port: %v. It must be between %v and %v"
	invalidProtocol       = "invalid protocol: %v. It must be TCP or UDP"
	invalidOwner          = "invalid owner ID. It must be UUID"
	invalidName           = "invalid name: %v. %v"
	invalidIP             = "invalid IP: %v. It must be a valid IP address, (e.g. 10.9.8.7)"
	invalidCPUQuota       = "invalid CPU quota: %v. It must be between %v(m) and %v(m)"
	invalidMemoryQuota    = "invalid memory quota: %v. It must be between %v(Mi) and %v(Mi)"
	subPathRelative       = "invalid Sub Path: %v. It must be relative path"
	noResource            = "resource '%v' is not found in %v"
	noNamespace           = "namespace is not found"
	resourceAlreadyExists = "resource '%v' already exists in %v"
	duplicateVolume       = "duplicate volume name '%v'"
	duplicateConfigMap    = "duplicate config map name '%v'"
	duplicateMountPath    = "duplicate mount path '%v'"
)

//ParseKubernetesResourceError checks error status
func ParseKubernetesResourceError(in interface{}, defaultErr *cherry.Err) *cherry.Err {
	sE, isStatusErrorCode := in.(*api_errors.StatusError)
	if isStatusErrorCode {
		switch sE.ErrStatus.Reason {
		case api_meta.StatusReasonNotFound:
			fmt.Println("TEST1", sE.Status().Details.Kind)
			if sE.Status().Details.Kind == "resourcequotas" {
				return kubeErrors.ErrResourceNotExist().AddDetails(noNamespace)
			}
			return kubeErrors.ErrResourceNotExist().AddDetailsErr(fmt.Errorf(noResource, sE.Status().Details.Name, sE.Status().Details.Kind))
		case api_meta.StatusReasonAlreadyExists:
			return kubeErrors.ErrResourceAlreadyExists().AddDetailsErr(fmt.Errorf(resourceAlreadyExists, sE.Status().Details.Name, sE.Status().Details.Kind))
		default:
			return defaultErr
		}
	}
	return defaultErr
}
