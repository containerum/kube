package model

import (
	"fmt"
	"strings"

	"time"

	"git.containerum.net/ch/kube-api/pkg/kubeerrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

// EndpointsList -- model for endpoints list
//
// swagger:model
type EndpointsList struct {
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint -- model for endpoint
//
// swagger:model
type Endpoint struct {
	// required: true
	Name  string `json:"name"`
	Owner string `json:"owner,omitempty"`
	//creation date in RFC3339 format
	CreatedAt *string `json:"created_at,omitempty"`
	// required: true
	Addresses []string `json:"addresses"`
	// required: true
	Ports []Port `json:"ports"`
}

// Port -- model for endpoint port
//
// swagger:model
type Port struct {
	// required: true
	Name string `json:"name"`
	// required: true
	Port int `json:"port"`
	// required: true
	Protocol kube_types.Protocol `json:"protocol"`
}

// ParseKubeEndpointList parses kubernetes v1.EndpointsList to more convenient []Endpoint struct
func ParseKubeEndpointList(endpointi interface{}) (*EndpointsList, error) {
	endpoints := endpointi.(*api_core.EndpointsList)
	if endpoints == nil {
		return nil, ErrUnableConvertEndpointList
	}
	newEndpoints := make([]Endpoint, 0)
	for _, ingress := range endpoints.Items {
		newEndpoint, err := ParseKubeEndpoint(&ingress)
		if err != nil {
			return nil, err
		}
		newEndpoints = append(newEndpoints, *newEndpoint)
	}
	return &EndpointsList{Endpoints: newEndpoints}, nil
}

// ParseKubeEndpoint parses kubernetes v1.Endpoint to more convenient Endpoint struct
func ParseKubeEndpoint(endpointi interface{}) (*Endpoint, error) {
	endpoint := endpointi.(*api_core.Endpoints)
	if endpoint == nil {
		return nil, ErrUnableConvertEndpoint
	}

	ports := make([]Port, 0)
	addresses := make([]string, 0)

	createdAt := endpoint.GetCreationTimestamp().UTC().Format(time.RFC3339)

	newEndpoint := Endpoint{
		Name:      endpoint.Name,
		Owner:     endpoint.GetObjectMeta().GetLabels()[ownerLabel],
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

func parseEndpointPort(np interface{}) Port {
	nativePort := np.(api_core.EndpointPort)
	return Port{
		Name:     nativePort.Name,
		Port:     int(nativePort.Port),
		Protocol: kube_types.Protocol(nativePort.Protocol),
	}
}

// ToKube creates kubernetes v1.Endpoint from Endpoint struct and namespace labels
func (endpoint *Endpoint) ToKube(nsName string, labels map[string]string) (*api_core.Endpoints, []error) {
	if err := endpoint.Validate(); err != nil {
		return nil, err
	}

	ipaddrs := make([]api_core.EndpointAddress, 0)
	for _, v := range endpoint.Addresses {
		ipaddrs = append(ipaddrs, api_core.EndpointAddress{
			IP: v,
		})
	}

	if labels == nil {
		return nil, []error{kubeerrors.ErrInternalError().AddDetails("invalid project labels")}
	}

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

func makeEndpointPorts(ports []Port) []api_core.EndpointPort {
	endpointports := make([]api_core.EndpointPort, 0)
	for _, v := range ports {
		endpointports = append(endpointports, api_core.EndpointPort{Name: v.Name, Protocol: api_core.Protocol(v.Protocol), Port: int32(v.Port)})
	}
	return endpointports
}

func (endpoint *Endpoint) Validate() []error {
	var errs []error
	if endpoint.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1123Label(endpoint.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, endpoint.Name, strings.Join(err, ",")))
	}
	if endpoint.Addresses == nil || len(endpoint.Addresses) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "addresses"))
	}
	for _, v := range endpoint.Addresses {
		if len(api_validation.IsValidIP(v)) > 0 {
			errs = append(errs, fmt.Errorf(invalidIP, v))
		}
	}
	if endpoint.Ports == nil || len(endpoint.Ports) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "ports"))
	}
	for _, v := range endpoint.Ports {
		if v.Name == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "ports.name"))
		} else if err := api_validation.IsDNS1123Label(v.Name); len(err) > 0 {
			errs = append(errs, fmt.Errorf(invalidName, v.Name, strings.Join(err, ",")))
		}
		if v.Protocol == "" {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "ports.protocol"))
		} else if v.Protocol != "TCP" && v.Protocol != "UDP" {
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
