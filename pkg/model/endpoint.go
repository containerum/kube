package model

import (
	api_core "k8s.io/api/core/v1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeEndpoint(nsName string, ips []string) *api_core.Endpoints {
	ipaddrs := []api_core.EndpointAddress{}

	for _, v := range ips {
		ipaddrs = append(ipaddrs, api_core.EndpointAddress{
			IP: v,
		})
	}

	return &api_core.Endpoints{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: "v1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Name:      glusterServiceName,
			Namespace: nsName,
		},
		Subsets: []api_core.EndpointSubset{
			{
				Addresses: ipaddrs,
				Ports: []api_core.EndpointPort{
					{
						Port:     1,
						Protocol: "TCP",
					},
				},
			},
		},
	}
}
