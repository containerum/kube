package model

import (
	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ParseIngressList(ingressi interface{}) []kube_types.Ingress {
	ingresses := ingressi.(*api_extensions.IngressList)

	var newIngresses []kube_types.Ingress
	for _, ingress := range ingresses.Items {
		newIngress := ParseIngress(&ingress)
		newIngresses = append(newIngresses, *newIngress)
	}
	return newIngresses
}

func ParseIngress(ingressi interface{}) *kube_types.Ingress {
	ingress := ingressi.(*api_extensions.Ingress)

	createdAt := ingress.CreationTimestamp.Unix()

	newIngress := kube_types.Ingress{}
	newIngress.Name = ingress.GetName()
	newIngress.CreatedAt = &createdAt

	if len(ingress.Spec.Rules) != 0 {
		newIngress.Rule.Host = ingress.Spec.Rules[0].Host
		if len(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths) != 0 {
			newIngress.Rule.Path.Path = ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
			newIngress.Rule.Path.ServiceName = ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName
			newIngress.Rule.Path.ServicePort = int(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal)
		}
	}
	if len(ingress.Spec.TLS) != 0 {
		newIngress.TLSSecret = ingress.Spec.TLS[0].SecretName
	}

	return &newIngress
}

func MakeIngress(ingress kube_types.Ingress) *api_extensions.Ingress {
	newIngress := api_extensions.Ingress{}
	newIngress.Kind = "Ingress"
	newIngress.APIVersion = "extensions/v1beta1"
	newIngress.SetName(ingress.Name)
	newIngress.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "nginx"}

	var path api_extensions.HTTPIngressPath
	path.Path = ingress.Rule.Path.Path
	path.Backend.ServicePort = intstr.FromInt(ingress.Rule.Path.ServicePort)
	path.Backend.ServiceName = ingress.Rule.Path.ServiceName

	var rulevalue = api_extensions.HTTPIngressRuleValue{}
	rulevalue.Paths = []api_extensions.HTTPIngressPath{path}

	var rule api_extensions.IngressRule
	rule.Host = ingress.Rule.Host
	rule.HTTP = &rulevalue

	newIngress.Spec.Rules = []api_extensions.IngressRule{rule}

	if ingress.TLSSecret != "" {
		newIngress.ObjectMeta.Annotations["kubernetes.io/tls-acme"] = "true"

		var tls api_extensions.IngressTLS
		tls.Hosts = []string{ingress.Rule.Host}
		tls.SecretName = ingress.TLSSecret

		newIngress.Spec.TLS = []api_extensions.IngressTLS{tls}
	}

	return &newIngress
}
