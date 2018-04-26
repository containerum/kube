package model

import (
	"errors"

	"fmt"

	ch "git.containerum.net/ch/cherry"
	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrUnableDecodeUserHeaderData    = errors.New("unable to decode user header data")
	ErrUnableUnmarshalUserHeaderData = errors.New("unable to unmarshal user header data")

	ErrUnableConvertServiceList = errors.New("unable to decode services list")
	ErrUnableConvertService     = errors.New("unable to decode service")

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
	noContainer           = "container %v is not found in deployment"
	fieldShouldExist      = "field %v should be provided"
	invalidReplicas       = "invalid replicas number: %v. It must be between 1 and %v"
	invalidPort           = "invalid port: %v. It must be between %v and %v"
	invalidProtocol       = "invalid protocol: %v. It must be TCP or UDP"
	invalidOwner          = "owner should be UUID"
	invalidName           = "invalid name: %v. %v"
	invalidIP             = "invalid IP: %v. It must be a valid IP address, (e.g. 10.9.8.7)"
	invalidCPUQuota       = "invalid CPU quota: %v. It must be between %vm and %vm"
	invalidMemoryQuota    = "invalid memory quota: %v. It must be between %vMi and %vMi"
	subPathRelative       = "invalid Sub Path: %v. It must be relative path"
	invalidResourceKind   = "invalid resource kind: %v. Shoud be %v"
	invalidAPIVersion     = "invalid API Version: %v. Shoud be %v"
	noResource            = "unable to find %v in %v"
	noNamespace           = "unable to find namesapce"
	resourceAlreadyExists = "%v already exists in %v"
)

//ParseKubernetesResourceError checks error status
func ParseKubernetesResourceError(in interface{}, defaultErr *ch.Err) *ch.Err {
	sE, isStatusErrorCode := in.(*api_errors.StatusError)
	if isStatusErrorCode {
		switch sE.ErrStatus.Reason {
		case api_meta.StatusReasonNotFound:
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
