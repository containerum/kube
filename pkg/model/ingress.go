package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_extensions "k8s.io/api/extensions/v1beta1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type IngressWithOwner struct {
	kube_types.Ingress
	Owner string `json:"owner,omitempty" binding:"required,uuid"`
}

// ParseIngressList parses kubernetes v1beta1.IngressList to more convenient []Ingress struct
func ParseIngressList(ingressi interface{}) ([]IngressWithOwner, error) {
	ingresses := ingressi.(*api_extensions.IngressList)
	if ingresses == nil {
		return nil, ErrUnableConvertIngressList
	}
	newIngresses := make([]IngressWithOwner, 0)
	for _, ingress := range ingresses.Items {
		newIngress, err := ParseIngress(&ingress)
		if err != nil {
			return nil, err
		}
		newIngresses = append(newIngresses, *newIngress)
	}
	return newIngresses, nil
}

// ParseIngress parses kubernetes v1beta1.Ingress to more convenient Ingress struct
func ParseIngress(ingressi interface{}) (*IngressWithOwner, error) {
	ingress := ingressi.(*api_extensions.Ingress)
	if ingress == nil {
		return nil, ErrUnableConvertIngress
	}
	createdAt := ingress.CreationTimestamp.Unix()
	owner := ingress.GetObjectMeta().GetLabels()[ownerLabel]

	newIngress := IngressWithOwner{
		Ingress: kube_types.Ingress{
			Name:      ingress.GetName(),
			CreatedAt: &createdAt,
			Rule:      *parseRule(ingress.Spec.Rules),
		},
		Owner: owner,
	}
	if len(ingress.Spec.TLS) != 0 {
		newIngress.TLSSecret = &ingress.Spec.TLS[0].SecretName
	}

	return &newIngress, nil
}

func parseRule(rules []api_extensions.IngressRule) *kube_types.Rule {
	rule := kube_types.Rule{}
	for _, v := range rules {
		rule.Host = v.Host
		for _, p := range v.HTTP.Paths {
			rule.Path.Path = p.Path
			rule.Path.ServiceName = p.Backend.ServiceName
			rule.Path.ServicePort = int(p.Backend.ServicePort.IntVal)
			break
		}
		break
	}
	return &rule
}

// MakeIngress creates kubernetes v1beta1.Ingress from Ingress struct and namespace labels
func MakeIngress(nsName string, ingress IngressWithOwner, labels map[string]string) *api_extensions.Ingress {
	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[appLabel] = ingress.Name
	labels[ownerLabel] = ingress.Owner
	labels[nameLabel] = ingress.Name

	newIngress := api_extensions.Ingress{
		TypeMeta: api_meta.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:      labels,
			Name:        ingress.Name,
			Namespace:   nsName,
			Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"},
		},
		Spec: api_extensions.IngressSpec{
			Rules: []api_extensions.IngressRule{
				{
					Host: ingress.Rule.Host,
					IngressRuleValue: api_extensions.IngressRuleValue{
						HTTP: &api_extensions.HTTPIngressRuleValue{
							Paths: []api_extensions.HTTPIngressPath{
								{
									Path: ingress.Rule.Path.Path,
									Backend: api_extensions.IngressBackend{
										ServiceName: ingress.Rule.Path.ServiceName,
										ServicePort: intstr.FromInt(ingress.Rule.Path.ServicePort),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if ingress.TLSSecret != nil {
		newIngress.ObjectMeta.Annotations["kubernetes.io/tls-acme"] = "true"
		newIngress.Spec.TLS = []api_extensions.IngressTLS{
			{
				Hosts:      []string{ingress.Rule.Host},
				SecretName: *ingress.TLSSecret,
			},
		}
	}

	return &newIngress
}
