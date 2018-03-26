package model

import (
	"errors"

	ch "git.containerum.net/ch/kube-client/pkg/cherry"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrInvalidCPUFormat    = errors.New("Invalid cpu quota format")
	ErrInvalidMemoryFormat = errors.New("Invalid memory quota format")

	ErrUnableDecodeUserHeaderData    = errors.New("Unbale to decode user header data")
	ErrUnableUnmarshalUserHeaderData = errors.New("Unable to unmarshal user header data")

	ErrUnableConvertServiceList = errors.New("Unable to decode services list")
	ErrUnableConvertService     = errors.New("Unable to decode service")

	ErrUnableConvertNamespaceList = errors.New("Unable to decode namespaces list")
	ErrUnableConvertNamespace     = errors.New("Unable to decode namespace")

	ErrUnableConvertSecretList = errors.New("Unable to decode secrets list")
	ErrUnableConvertSecret     = errors.New("Unable to decode secret")

	ErrUnableConvertIngressList = errors.New("Unable to decode ingresses list")
	ErrUnableConvertIngress     = errors.New("Unable to decode ingress")

	ErrUnableConvertDeploymentList = errors.New("Unable to decode deployment list")
	ErrUnableConvertDeployment     = errors.New("Unable to decode deployment")

	ErrUnableConvertEndpointList = errors.New("Unable to decode services list")
	ErrUnableConvertEndpoint     = errors.New("Unable to decode service")

	ErrUnableConvertConfigMapList = errors.New("Unable to decode config maps list")
	ErrUnableConvertConfigMap     = errors.New("Unable to decode config map")
)

const (
	noContainer        = "Container %v is not found in deployment"
	fieldShouldExist   = "Field %v should be provided"
	invalidReplicas    = "Invalid replicas number: %v. It must be between 1 and %v"
	invalidPort        = "Invalid port: %v. It must be between %v and %v"
	invalidProtocol    = "Invalid protocol: %v. It must be TCP or UDP"
	invalidOwner       = "Owner should be UUID"
	invalidName        = "Invalid name: %v. %v"
	invalidIP          = "Invalid IP: %v. It must be a valid IP address, (e.g. 10.9.8.7)"
	invalidCPUQuota    = "Invalid CPU quota: %v. It must be between %v and %v"
	invalidMemoryQuota = "Invalid memory quota: %v. It must be between %v and %v"
	subPathRelative    = "Invalid Sub Path: %v. It must be relative path"
)

//ParseResourceError checks error status
func ParseResourceError(in interface{}, defaulterr *ch.Err) *ch.Err {
	sE, isStatusErrorCode := in.(*api_errors.StatusError)
	if isStatusErrorCode {
		switch sE.ErrStatus.Reason {
		case api_meta.StatusReasonNotFound:
			return cherry.ErrResourceNotExist()
		case api_meta.StatusReasonAlreadyExists:
			return cherry.ErrResourceAlreadyExists()
		default:
			return defaulterr
		}
	}
	return defaulterr
}
