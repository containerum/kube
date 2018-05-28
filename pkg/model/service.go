package model

import (
	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"

	"strconv"

	"fmt"
	"strings"

	"time"

	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	domainLabel = "domain"
	hiddenLabel = "hidden"

	minport = 11000
	maxport = 65535
)

const (
	serviceKind       = "Service"
	serviceAPIVersion = "v1"
)

// ServicesList -- model for services list
//
// swagger:model
type ServicesList struct {
	Services []ServiceWithOwner `json:"services"`
}

// ServiceWithOwner -- model for service with owner
//
// swagger:model
type ServiceWithOwner struct {
	// swagger: allOf
	kube_types.Service
	// required: true
	Owner string `json:"owner,omitempty"`
	//hide service from users
	Hidden bool `json:"hidden,omitempty"`
	//don't add selectors to service (so don't create endpoint)
	NoSelector bool `json:"no_selector,omitempty"`
}

// ParseKubeServiceList parses kubernetes v1.ServiceList to more convenient Service struct.
func ParseKubeServiceList(ns interface{}, parseforuser bool) (*ServicesList, error) {
	nativeServices := ns.(*api_core.ServiceList)
	if nativeServices == nil {
		return nil, ErrUnableConvertServiceList
	}

	serviceList := make([]ServiceWithOwner, 0)
	for _, nativeService := range nativeServices.Items {
		service, err := ParseKubeService(&nativeService, parseforuser)
		if err != nil {
			return nil, err
		}

		if !service.Hidden || !parseforuser {
			serviceList = append(serviceList, *service)
		}
	}
	return &ServicesList{serviceList}, nil
}

// ParseKubeService parses kubernetes v1.Service to more convenient Service struct.
func ParseKubeService(srv interface{}, parseforuser bool) (*ServiceWithOwner, error) {
	native := srv.(*api_core.Service)
	if native == nil {
		return nil, ErrUnableConvertService
	}

	ports := make([]kube_types.ServicePort, 0)

	createdAt := native.GetCreationTimestamp().UTC().UTC().Format(time.RFC3339)
	owner := native.GetObjectMeta().GetLabels()[ownerLabel]
	deploy := native.GetObjectMeta().GetLabels()[appLabel]
	domain := native.GetObjectMeta().GetLabels()[domainLabel]

	service := ServiceWithOwner{
		Service: kube_types.Service{
			Name:      native.Name,
			CreatedAt: &createdAt,
			Ports:     ports,
			Deploy:    deploy,
			Domain:    domain,
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

	if parseforuser {
		service.ParseForUser()
	}

	return &service, nil
}

func parseServicePort(np interface{}) kube_types.ServicePort {
	nativePort := np.(api_core.ServicePort)
	port := int(nativePort.Port)
	targetPort := int(nativePort.TargetPort.IntVal)
	return kube_types.ServicePort{
		Name:       nativePort.Name,
		Port:       &port,
		TargetPort: targetPort,
		Protocol:   kube_types.Protocol(nativePort.Protocol),
	}
}

// ToKube creates kubernetes v1.Service from Service struct and namespace labels
func (service *ServiceWithOwner) ToKube(nsName string, labels map[string]string) (*api_core.Service, []error) {
	err := service.Validate()
	if err != nil {
		return nil, err
	}

	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError().AddDetails("invalid namespace labels")}
	}
	labels[appLabel] = service.Deploy
	labels[hiddenLabel] = strconv.FormatBool(service.Hidden)

	newService := api_core.Service{
		TypeMeta: api_meta.TypeMeta{
			Kind:       serviceKind,
			APIVersion: serviceAPIVersion,
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      service.Name,
			Namespace: nsName,
		},
		Spec: api_core.ServiceSpec{
			Type:  "ClusterIP",
			Ports: makeServicePorts(service.Ports),
		},
	}

	if !service.NoSelector {
		selector := make(map[string]string, 0)
		selector[appLabel] = service.Deploy
		selector[ownerLabel] = labels[ownerLabel]
		newService.Spec.Selector = selector
	}

	if service.IPs != nil {
		newService.Spec.ExternalIPs = service.IPs

		if service.Domain != "" {
			newService.ObjectMeta.Labels[domainLabel] = service.Domain
		}
	}

	return &newService, nil
}

func makeServicePorts(ports []kube_types.ServicePort) []api_core.ServicePort {
	var serviceports []api_core.ServicePort
	if ports != nil {
		for _, port := range ports {
			if port.Port == nil {
				port.Port = &port.TargetPort
			}
			serviceports = append(serviceports, api_core.ServicePort{
				Name:       port.Name,
				Protocol:   api_core.Protocol(port.Protocol),
				Port:       int32(*port.Port),
				TargetPort: intstr.FromInt(port.TargetPort),
			})
		}
	}
	return serviceports
}

func (service *ServiceWithOwner) Validate() []error {
	errs := []error{}
	if service.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1035Label(service.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, service.Name, strings.Join(err, ",")))
	}
	if service.Deploy == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "deploy"))
	}
	if service.Ports == nil || len(service.Ports) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "ports"))
	}
	for _, v := range service.Ports {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "ports.name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.Protocol == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "ports.protocol"))
		} else if v.Protocol != kube_types.UDP && v.Protocol != kube_types.TCP {
			errs = append(errs, fmt.Errorf(invalidProtocol, v.Protocol))
		}
		if len(service.IPs) > 0 {
			if len(api_validation.IsInRange(*v.Port, minport, maxport)) > 0 {
				errs = append(errs, fmt.Errorf(invalidPort, *v.Port, minport, maxport))
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ParseForUser removes information not interesting for users
func (service *ServiceWithOwner) ParseForUser() {
	if service.Owner == "" {
		service.Hidden = true
		return
	}
	service.Owner = ""
}
