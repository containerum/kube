package model

import (
	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	api_extensions "k8s.io/api/extensions/v1beta1"

	"fmt"
	"strings"

	"time"

	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

const (
	ingressKind       = "Ingress"
	ingressAPIVersion = "extensions/v1beta1"
)

type IngressKubeAPI kube_types.Ingress

// ParseKubeIngressList parses kubernetes v1beta1.IngressList to more convenient []Ingress struct
func ParseKubeIngressList(ingressi interface{}, parseforuser bool) (*kube_types.IngressesList, error) {
	ingresses := ingressi.(*api_extensions.IngressList)
	if ingresses == nil {
		return nil, ErrUnableConvertIngressList
	}
	newIngresses := make([]kube_types.Ingress, 0)
	for _, ingress := range ingresses.Items {
		newIngress, err := ParseKubeIngress(&ingress, parseforuser)
		if err != nil {
			return nil, err
		}
		newIngresses = append(newIngresses, *newIngress)
	}
	return &kube_types.IngressesList{newIngresses}, nil
}

// ParseKubeIngress parses kubernetes v1beta1.Ingress to more convenient Ingress struct
func ParseKubeIngress(ingressi interface{}, parseforuser bool) (*kube_types.Ingress, error) {
	ingress := ingressi.(*api_extensions.Ingress)
	if ingress == nil {
		return nil, ErrUnableConvertIngress
	}

	newIngress := kube_types.Ingress{
		Name:      ingress.GetName(),
		CreatedAt: ingress.CreationTimestamp.UTC().Format(time.RFC3339),
		Rules:     parseRules(ingress.Spec.Rules, parseTLS(ingress.Spec.TLS)),
		Owner:     ingress.GetObjectMeta().GetLabels()[ownerLabel],
	}

	if parseforuser {
		newIngress.Mask()
	}

	return &newIngress, nil
}

func parseRules(rules []api_extensions.IngressRule, secrets map[string]string) []kube_types.Rule {
	newRules := make([]kube_types.Rule, 0)

	for _, v := range rules {
		newRule := kube_types.Rule{}
		newRule.Host = v.Host

		secret, ok := secrets[newRule.Host]
		if ok {
			newRule.TLSSecret = &secret
		}

		for _, p := range v.HTTP.Paths {
			newPath := kube_types.Path{}
			newPath.Path = p.Path
			newPath.ServiceName = p.Backend.ServiceName
			newPath.ServicePort = int(p.Backend.ServicePort.IntVal)
			newRule.Path = append(newRule.Path, newPath)
		}
		newRules = append(newRules, newRule)
	}
	return newRules
}

func parseTLS(tlss []api_extensions.IngressTLS) map[string]string {
	secrets := make(map[string]string, 0)

	for _, v := range tlss {
		for _, h := range v.Hosts {
			secrets[h] = v.SecretName
		}
	}
	return secrets
}

// ToKube creates kubernetes v1beta1.Ingress from Ingress struct and namespace labels
func (ingress *IngressKubeAPI) ToKube(nsName string, labels map[string]string) (*api_extensions.Ingress, []error) {
	err := ingress.Validate()
	if err != nil {
		return nil, err
	}
	if labels == nil {
		return nil, []error{kubeErrors.ErrInternalError().AddDetails("invalid namespace labels")}
	}

	rules, secrets, tls := makeIngressRules(ingress.Rules)

	newIngress := api_extensions.Ingress{
		TypeMeta: api_meta.TypeMeta{
			Kind:       ingressKind,
			APIVersion: ingressAPIVersion,
		},
		ObjectMeta: api_meta.ObjectMeta{
			Labels:      labels,
			Name:        ingress.Name,
			Namespace:   nsName,
			Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"},
		},
		Spec: api_extensions.IngressSpec{
			Rules: rules,
			TLS:   secrets,
		},
	}

	if tls {
		newIngress.ObjectMeta.Annotations["kubernetes.io/tls-acme"] = "true"
	}
	return &newIngress, nil
}

func makeIngressRules(rules []kube_types.Rule) ([]api_extensions.IngressRule, []api_extensions.IngressTLS, bool) {
	newRules := make([]api_extensions.IngressRule, 0)
	secrets := make([]api_extensions.IngressTLS, 0)
	tls := false

	for _, v := range rules {
		paths := []api_extensions.HTTPIngressPath{}
		for _, p := range v.Path {
			paths = append(paths, api_extensions.HTTPIngressPath{
				Path: p.Path,
				Backend: api_extensions.IngressBackend{
					ServiceName: p.ServiceName,
					ServicePort: intstr.FromInt(p.ServicePort),
				},
			})
		}
		newRules = append(newRules, api_extensions.IngressRule{
			Host: v.Host,
			IngressRuleValue: api_extensions.IngressRuleValue{
				HTTP: &api_extensions.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})

		if v.TLSSecret != nil {
			tls = true
			secrets = append(secrets, api_extensions.IngressTLS{
				Hosts:      []string{v.Host},
				SecretName: *v.TLSSecret,
			})
		}
	}
	return newRules, secrets, tls
}

func (ingress *IngressKubeAPI) Validate() []error {
	var errs []error
	if ingress.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "name"))
	} else if err := api_validation.IsDNS1123Subdomain(ingress.Name); len(err) > 0 {
		errs = append(errs, fmt.Errorf(invalidName, ingress.Name, strings.Join(err, ",")))
	}
	if ingress.Rules == nil || len(ingress.Rules) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "rules"))
	}
	for _, v := range ingress.Rules {
		if v.Path == nil || len(v.Path) == 0 {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "rules.path"))
		}
		for _, p := range v.Path {
			if p.ServiceName == "" {
				errs = append(errs, fmt.Errorf(fieldShouldExist, "rules.path.name"))
			} else if err := api_validation.IsDNS1035Label(p.ServiceName); len(err) > 0 {
				errs = append(errs, fmt.Errorf(invalidName, p.ServiceName, strings.Join(err, ",")))
			}
			if len(api_validation.IsValidPortNum(p.ServicePort)) > 0 {
				errs = append(errs, fmt.Errorf(invalidPort, p.ServicePort, 1, maxport))
			}
			if p.Path == "" {
				errs = append(errs, fmt.Errorf(fieldShouldExist, "rules.path.name.path"))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
