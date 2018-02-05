package model

import (
	kubeCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

// ServicePortFromNativePort converts native
// cubernetes service port representation to user friendly ServicePort struct
func ServicePortFromNativePort(nativePort kubeCoreV1.ServicePort) ServicePort {
	return ServicePort{
		Name:       nativePort.Name,
		Port:       uint32(nativePort.Port),
		TargetPort: nativePort.TargetPort,
		Protocol:   nativePort.Protocol,
	}
}
