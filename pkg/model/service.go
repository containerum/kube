package model

import (
	json_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"

	"github.com/gin-gonic/gin/binding"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	serviceTypeExternal = "external"
	serviceTypeInternal = "internal"
)

// ParseServicePort converts native
// cubernetes service port representation to user friendly ServicePort struct
func ParseServicePort(nativePort api_core.ServicePort) json_types.ServicePort {
	targetPort := int(nativePort.TargetPort.IntVal)
	return json_types.ServicePort{
		Name:       nativePort.Name,
		Port:       int(nativePort.Port),
		TargetPort: targetPort,
		Protocol:   json_types.Protocol(nativePort.Protocol),
	}
}

// ParseService creates
// user friendly service representation
func ParseService(native *api_core.Service) *json_types.Service {
	ports := make([]json_types.ServicePort, 0, 1)

	service := &json_types.Service{
		Name:      native.Name,
		CreatedAt: native.GetCreationTimestamp().Unix(),
		Deploy:    native.GetObjectMeta().GetLabels()["app"], // TODO: check if app key doesn't exists!
		Domain:    "",                                        // TODO : add domain info!
		Ports:     ports,
		Owner:     native.GetObjectMeta().GetLabels()["owner"],
	}
	if len(native.Spec.ExternalIPs) > 0 {
		service.Type = serviceTypeExternal
		service.IP = native.Spec.ExternalIPs
	} else {
		service.Type = serviceTypeInternal
		service.IP = []string{}
	}
	for _, nativePort := range native.Spec.Ports {
		service.Ports = append(service.Ports,
			ParseServicePort(nativePort))
	}
	return service
}

func ParseServiceList(nativeServices *api_core.ServiceList) ([]json_types.Service, error) {
	if nativeServices == nil {
		return nil, ErrUnableConvertServiceList
	}
	serviceList := make([]json_types.Service, 0, nativeServices.Size())
	for _, nativeService := range nativeServices.Items {
		service := ParseService(&nativeService)
		serviceList = append(serviceList, *service)
	}
	return serviceList, nil
}

func MakeService(nsName string, service *json_types.Service) (*api_core.Service, error) {
	var ports []api_core.ServicePort
	if service.Ports != nil {
		for _, v := range service.Ports {
			err := binding.Validator.ValidateStruct(v)
			if err != nil {
				return nil, err
			}
			ports = append(ports, api_core.ServicePort{Name: v.Name, Protocol: api_core.Protocol(v.Protocol), Port: int32(v.Port), TargetPort: intstr.FromInt(v.TargetPort)})
		}
	}

	var newService api_core.Service
	newService.Spec.Selector = map[string]string{"app": service.Deploy, "owner": service.Owner}
	newService.SetLabels(map[string]string{"app": service.Deploy, "owner": service.Owner})
	newService.Spec.Ports = ports
	newService.Spec.ExternalIPs = service.IP
	newService.SetName(service.Name)
	newService.SetNamespace(nsName)
	return &newService, nil
}
