package model

import (
	"errors"
	"fmt"

	json_types "git.containerum.net/ch/json-types/kube-api"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

// ParseEndpointList parses kubernetes v1.EndpointsList to more convenient []Endpoint struct
func ParseEndpointList(endpointi interface{}) ([]json_types.Endpoint, error) {
	endpoints := endpointi.(*api_core.EndpointsList)
	if endpoints == nil {
		return nil, ErrUnableConvertEndpointList
	}
	newEndpoints := make([]json_types.Endpoint, 0)
	for _, ingress := range endpoints.Items {
		newEndpoint, err := ParseEndpoint(&ingress)
		if err != nil {
			return nil, err
		}
		newEndpoints = append(newEndpoints, *newEndpoint)
	}
	return newEndpoints, nil
}

// ParseEndpoint parses kubernetes v1.Endpoint to more convenient Endpoint struct
func ParseEndpoint(endpointi interface{}) (*json_types.Endpoint, error) {
	endpoint := endpointi.(*api_core.Endpoints)
	if endpoint == nil {
		return nil, ErrUnableConvertEndpoint
	}

	ports := make([]json_types.Port, 0)
	addresses := make([]string, 0)

	createdAt := endpoint.GetCreationTimestamp().Unix()
	owner := endpoint.GetObjectMeta().GetLabels()[ownerLabel]

	newEndpoint := json_types.Endpoint{
		Name:      endpoint.Name,
		Owner:     &owner,
		CreatedAt: &createdAt,
		Ports:     ports,
		Addresses: addresses,
	}

	if len(endpoint.Subsets) != 0 {
		for _, nativePort := range endpoint.Subsets[0].Ports {
			newEndpoint.Ports = append(newEndpoint.Ports,
				parseEndpointPort(nativePort))
		}

		for _, v := range endpoint.Subsets[0].Addresses {
			newEndpoint.Addresses = append(newEndpoint.Addresses, v.IP)
		}
	}

	return &newEndpoint, nil
}

func parseEndpointPort(np interface{}) json_types.Port {
	nativePort := np.(api_core.EndpointPort)
	return json_types.Port{
		Name:     nativePort.Name,
		Port:     int(nativePort.Port),
		Protocol: json_types.Protocol(nativePort.Protocol),
	}
}

// MakeEndpoint creates kubernetes v1.Endpoint from Endpoint struct and namespace labels
func MakeEndpoint(nsName string, endpoint json_types.Endpoint, labels map[string]string) (*api_core.Endpoints, []error) {
	err := validateEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	ipaddrs := []api_core.EndpointAddress{}
	for _, v := range endpoint.Addresses {
		ipaddrs = append(ipaddrs, api_core.EndpointAddress{
			IP: v,
		})
	}

	if labels == nil {
		labels = make(map[string]string, 0)
	}

	labels[appLabel] = endpoint.Name
	labels[ownerLabel] = *endpoint.Owner
	labels[nameLabel] = endpoint.Name

	newEndpoint := api_core.Endpoints{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:    labels,
			Name:      endpoint.Name,
			Namespace: nsName,
		},
		Subsets: []api_core.EndpointSubset{
			{
				Addresses: ipaddrs,
				Ports:     makeEndpointPorts(endpoint.Ports),
			},
		},
	}

	return &newEndpoint, nil
}

func makeEndpointPorts(ports []json_types.Port) []api_core.EndpointPort {
	endpointports := make([]api_core.EndpointPort, 0)
	if ports != nil {
		for _, v := range ports {
			endpointports = append(endpointports, api_core.EndpointPort{Name: v.Name, Protocol: api_core.Protocol(v.Protocol), Port: int32(v.Port)})
		}
	}
	return endpointports
}

func validateEndpoint(endpoint json_types.Endpoint) []error {
	errs := []error{}
	if endpoint.Owner == nil {
		errs = append(errs, errors.New(noOwner))
	} else {
		if !IsValidUUID(*endpoint.Owner) {
			errs = append(errs, errors.New(invalidOwner))
		}
	}
	if len(api_validation.IsDNS1123Subdomain(endpoint.Name)) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, endpoint.Name))
	}
	if endpoint.Addresses == nil || len(endpoint.Addresses) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Addresses"))
	}
	for _, v := range endpoint.Addresses {
		if len(api_validation.IsValidIP(v)) > 0 {
			errs = append(errs, fmt.Errorf(invalidIP, v))
		}
	}
	if endpoint.Ports == nil || len(endpoint.Ports) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Ports"))
	}
	for _, v := range endpoint.Ports {
		if len(api_validation.IsValidPortName(v.Name)) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name))
		}
		if v.Protocol != "TCP" && v.Protocol != "UDP" {
			errs = append(errs, fmt.Errorf(invalidProtocol, v.Protocol))
		}
		if len(api_validation.IsValidPortNum(v.Port)) > 0 {
			errs = append(errs, fmt.Errorf(invalidPort, v.Port, 1, maxport))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
