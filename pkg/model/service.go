package model

import (
	"errors"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"

	"strconv"

	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	minport = 11000
	maxport = 65535
)

type ServiceWithOwner struct {
	kube_types.Service
	Owner      string `json:"owner,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
	NoSelector bool   `json:"no_selector,omitempty"`
}

const (
	hiddenLabel = "hidden"
)

// ParseServiceList parses kubernetes v1.ServiceList to more convenient Service struct.
func ParseServiceList(ns interface{}) ([]ServiceWithOwner, error) {
	nativeServices := ns.(*api_core.ServiceList)
	if nativeServices == nil {
		return nil, ErrUnableConvertServiceList
	}

	serviceList := make([]ServiceWithOwner, 0, nativeServices.Size())
	for _, nativeService := range nativeServices.Items {
		service, err := ParseService(&nativeService)
		if err != nil {
			return nil, err
		}

		if !service.Hidden {
			serviceList = append(serviceList, *service)
		}
	}
	return serviceList, nil
}

// ParseService parses kubernetes v1.Service to more convenient Service struct.
func ParseService(srv interface{}) (*ServiceWithOwner, error) {
	native := srv.(*api_core.Service)
	if native == nil {
		return nil, ErrUnableConvertService
	}

	ports := make([]kube_types.ServicePort, 0, 1)

	createdAt := native.GetCreationTimestamp().Unix()
	owner := native.GetObjectMeta().GetLabels()[ownerLabel]

	service := ServiceWithOwner{
		Service: kube_types.Service{
			Name:      native.Name,
			CreatedAt: &createdAt,
			Deploy:    native.GetObjectMeta().GetLabels()[appLabel], // TODO: check if app key doesn't exists!
			Ports:     ports,
		},
		Owner: owner,
	}
	if len(native.Spec.ExternalIPs) > 0 {
		service.IPs = native.Spec.ExternalIPs
	} else {
		service.IPs = []string{}
	}
	for _, nativePort := range native.Spec.Ports {
		service.Ports = append(service.Ports,
			parseServicePort(nativePort))
	}

	if s, err := strconv.ParseBool(native.GetObjectMeta().GetLabels()[hiddenLabel]); err == nil {
		service.Hidden = s
	}

	return &service, nil
}

func parseServicePort(np interface{}) kube_types.ServicePort {
	nativePort := np.(api_core.ServicePort)
	targetPort := int(nativePort.TargetPort.IntVal)
	return kube_types.ServicePort{
		Name:       nativePort.Name,
		Port:       int(nativePort.Port),
		TargetPort: &targetPort,
		Protocol:   kube_types.Protocol(nativePort.Protocol),
	}
}

// MakeService creates kubernetes v1.Service from Service struct and namespace labels
func MakeService(nsName string, service ServiceWithOwner, labels map[string]string) (*api_core.Service, []error) {
	err := validateService(service)
	if err != nil {
		return nil, err
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = service.Name
	labels[ownerLabel] = service.Owner
	labels[nameLabel] = service.Name
	labels[hiddenLabel] = strconv.FormatBool(service.Hidden)

	newService := api_core.Service{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:          labels,
			Name:            service.Name,
			Namespace:       nsName,
			ResourceVersion: "TEST",
		},
		Spec: api_core.ServiceSpec{
			Type:  "ClusterIP",
			Ports: makeServicePorts(service.Ports),
		},
	}

	if !service.NoSelector {
		newService.Spec.Selector = labels
	}

	if service.IPs != nil {
		newService.Spec.ExternalIPs = service.IPs
	}

	return &newService, nil
}

func makeServicePorts(ports []kube_types.ServicePort) []api_core.ServicePort {
	var serviceports []api_core.ServicePort
	if ports != nil {
		for _, v := range ports {
			var targetport intstr.IntOrString
			if v.TargetPort != nil {
				targetport = intstr.FromInt(*v.TargetPort)
			}
			serviceports = append(serviceports, api_core.ServicePort{Name: v.Name, Protocol: api_core.Protocol(v.Protocol), Port: int32(v.Port), TargetPort: targetport})
		}
	}
	return serviceports
}

func validateService(service ServiceWithOwner) []error {
	errs := []error{}
	if service.Owner == "" {
		errs = append(errs, errors.New(noOwner))
	} else {
		if !IsValidUUID(service.Owner) {
			errs = append(errs, errors.New(invalidOwner))
		}
	}
	if len(api_validation.IsDNS1123Subdomain(service.Name)) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, service.Name))
	}
	if service.Ports == nil || len(service.Ports) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Ports"))
	}
	for _, v := range service.Ports {
		if len(api_validation.IsValidPortName(v.Name)) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name))
		}
		if v.Protocol != kube_types.UDP && v.Protocol != kube_types.TCP {
			errs = append(errs, fmt.Errorf(invalidProtocol, v.Protocol))
		}
		if len(api_validation.IsInRange(v.Port, minport, maxport)) > 0 {
			errs = append(errs, fmt.Errorf(invalidPort, v.Port, minport, maxport))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
