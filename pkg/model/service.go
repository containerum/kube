package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin/binding"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	serviceTypeExternal = "external"
	serviceTypeInternal = "internal"
)

// ParseServiceList parses kubernetes v1.ServiceList to more convenient Service struct.
func ParseServiceList(ns interface{}) ([]kube_types.Service, error) {
	nativeServices := ns.(*api_core.ServiceList)
	if nativeServices == nil {
		return nil, ErrUnableConvertServiceList
	}

	serviceList := make([]kube_types.Service, 0, nativeServices.Size())
	for _, nativeService := range nativeServices.Items {
		service, err := ParseService(&nativeService)
		if err != nil {
			return nil, err
		}
		serviceList = append(serviceList, *service)
	}
	return serviceList, nil
}

// ParseService parses kubernetes v1.Service to more convenient Service struct.
func ParseService(srv interface{}) (*kube_types.Service, error) {
	native := srv.(*api_core.Service)
	if native == nil {
		return nil, ErrUnableConvertServiceList
	}

	ports := make([]kube_types.Port, 0, 1)

	createdAt := native.GetCreationTimestamp().Unix()
	owner := native.GetLabels()[ownerLabel]

	service := &kube_types.Service{
		Name:      native.Name,
		CreatedAt: &createdAt,
		Deploy:    native.GetObjectMeta().GetLabels()[appLabel], // TODO: check if app key doesn't exists!
		Ports:     ports,
		Owner:     &owner,
	}
	if len(native.Spec.ExternalIPs) > 0 {
		service.Type = serviceTypeExternal
		service.IP = &native.Spec.ExternalIPs
	} else {
		service.Type = serviceTypeInternal
		service.IP = &[]string{}
	}
	for _, nativePort := range native.Spec.Ports {
		service.Ports = append(service.Ports,
			parseServicePort(nativePort))
	}
	return service, nil
}

func parseServicePort(np interface{}) kube_types.Port {
	nativePort := np.(api_core.ServicePort)
	targetPort := int(nativePort.TargetPort.IntVal)
	return kube_types.Port{
		Name:       nativePort.Name,
		Port:       int(nativePort.Port),
		TargetPort: &targetPort,
		Protocol:   kube_types.Protocol(nativePort.Protocol),
	}
}

// MakeService creates kubernetes v1.Service from Service struct and namespace labels
func MakeService(nsName string, service *kube_types.Service, labels map[string]string) (*api_core.Service, error) {
	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = service.Name
	labels[ownerLabel] = *service.Owner

	newService := api_core.Service{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      service.Name,
			Namespace: nsName,
		},
		Spec: api_core.ServiceSpec{
			Selector:    labels,
			ExternalIPs: *service.IP,
		},
	}

	if sp, err := makeServicePorts(service.Ports); err != nil {
		return nil, err
	} else {
		newService.Spec.Ports = sp
	}

	return &newService, nil
}

func makeServicePorts(ports []kube_types.Port) ([]api_core.ServicePort, error) {
	var serviceports []api_core.ServicePort
	if ports != nil {
		for _, v := range ports {
			err := binding.Validator.ValidateStruct(v)
			if err != nil {
				return nil, err
			}
			serviceports = append(serviceports, api_core.ServicePort{Name: v.Name, Protocol: api_core.Protocol(v.Protocol), Port: int32(v.Port), TargetPort: intstr.FromInt(*v.TargetPort)})
		}
	}
	return serviceports, nil
}
