package model

import (
	"errors"

	kubeCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	ErrUnableConvertServiceList = errors.New("unable decode service list")
	ErrUnableConvertService     = errors.New("unable convert cubernetes service to user representation")
)

// ServicePort is an user friendly service port representation
// Name is DNS_LABEL
// TargetPort is an int32 or IANA_SVC_NAME
// Protocol is TCP or UDP
type ServicePort struct {
	Name       string              `json:"name"`
	Port       uint32              `json:"port"`
	TargetPort intstr.IntOrString  `json:"target_port"`
	Protocol   kubeCoreV1.Protocol `json:"protocol"`
}

// ServicePortFromNativeKubePort converts native
// cubernetes service port representation to user friendly ServicePort struct
func ServicePortFromNativeKubePort(nativePort kubeCoreV1.ServicePort) ServicePort {
	return ServicePort{
		Name:       nativePort.Name,
		Port:       uint32(nativePort.Port),
		TargetPort: nativePort.TargetPort,
		Protocol:   nativePort.Protocol,
	}
}

// Service is an user friendly kebernetes service representation
// CreatedAt is an unix timestamp
type Service struct {
	CreatedAt int64                  `json:"created_at"`
	Deploy    string                 `json:"deploy"`
	IP        []string               `json:"ip"`
	Domain    string                 `json:"domain, omitempty"`
	Type      kubeCoreV1.ServiceType `json:"type"`
	Ports     []ServicePort          `json:"ports"`
}

// ServiceFromNativeKubeService creates
// user friendly service representation
func ServiceFromNativeKubeService(native *kubeCoreV1.Service) (*Service, error) {
	if native == nil {
		return nil, ErrUnableConvertService
	}
	service := &Service{
		CreatedAt: native.GetCreationTimestamp().Unix(),
		Deploy:    native.GetObjectMeta().GetLabels()["app"], // TODO: check if app key doesn't exists!
		IP:        native.Spec.ExternalIPs,
		Domain:    "", // TODO : add domain info!
		Type:      native.Spec.Type,
		Ports:     make([]ServicePort, 0, 1),
	}
	for _, nativePort := range native.Spec.Ports {
		service.Ports = append(service.Ports,
			ServicePortFromNativeKubePort(nativePort))
	}
	return service, nil
}

func ParseServiceList(nativeServices *kubeCoreV1.ServiceList) ([]Service, error) {
	if nativeServices == nil {
		return nil, ErrUnableConvertServiceList
	}
	serviceList := make([]Service, 0, nativeServices.Size())
	for _, nativeService := range nativeServices.Items {
		// error can be ignored because ServiceList provides
		// Service stucts by values
		service, _ := ServiceFromNativeKubeService(&nativeService)
		serviceList = append(serviceList, *service)
	}
	return serviceList, nil
}
