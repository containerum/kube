package model

import (
	"errors"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"
	api_extensions "k8s.io/api/extensions/v1beta1"

	"fmt"
	"strings"

	"time"

	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	api_validation "k8s.io/apimachinery/pkg/util/validation"
)

type IngressesList struct {
	Ingress []IngressWithOwner `json:"ingresses"`
}

type IngressWithOwner struct {
	kube_types.Ingress
	Owner string `json:"owner,omitempty"`
}

const (
	ingressKind       = "Ingress"
	ingressApiVersion = "extensions/v1beta1"
)

// ParseIngressList parses kubernetes v1beta1.IngressList to more convenient []Ingress struct
func ParseIngressList(ingressi interface{}, parseforuser bool) (*IngressesList, error) {
	ingresses := ingressi.(*api_extensions.IngressList)
	if ingresses == nil {
		return nil, ErrUnableConvertIngressList
	}
	newIngresses := make([]IngressWithOwner, 0)
	for _, ingress := range ingresses.Items {
		newIngress, err := ParseIngress(&ingress, parseforuser)
		if err != nil {
			return nil, err
		}
		newIngresses = append(newIngresses, *newIngress)
	}
	return &IngressesList{newIngresses}, nil
}

// ParseIngress parses kubernetes v1beta1.Ingress to more convenient Ingress struct
func ParseIngress(ingressi interface{}, parseforuser bool) (*IngressWithOwner, error) {
	ingress := ingressi.(*api_extensions.Ingress)
	if ingress == nil {
		return nil, ErrUnableConvertIngress
	}
	createdAt := ingress.CreationTimestamp.Format(time.RFC3339)
	owner := ingress.GetObjectMeta().GetLabels()[ownerLabel]

	secrets := parseTLS(ingress.Spec.TLS)

	rules := parseRules(ingress.Spec.Rules, secrets)

	newIngress := IngressWithOwner{
		Ingress: kube_types.Ingress{
			Name:      ingress.GetName(),
			CreatedAt: &createdAt,
			Rules:     rules,
		},
		Owner: owner,
	}

	if parseforuser {
		newIngress.Owner = ""
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

// MakeIngress creates kubernetes v1beta1.Ingress from Ingress struct and namespace labels
func MakeIngress(nsName string, ingress IngressWithOwner, labels map[string]string) (*api_extensions.Ingress, []error) {
	err := ValidateIngress(ingress)
	if err != nil {
		return nil, err
	}
	if labels == nil {
		labels = make(map[string]string, 0)
	}
	labels[ownerLabel] = ingress.Owner

	rules, secrets, tls := makeIngressRules(ingress.Rules)

	newIngress := api_extensions.Ingress{
		TypeMeta: api_meta.TypeMeta{
			Kind:       ingressKind,
			APIVersion: ingressApiVersion,
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

	if tls == true {
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

func ValidateIngress(ingress IngressWithOwner) []error {
	errs := []error{}
	if ingress.Owner == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Owner"))
	} else if !IsValidUUID(ingress.Owner) {
		errs = append(errs, errors.New(invalidOwner))
	}
	if ingress.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Subdomain(ingress.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, ingress.Name, strings.Join(err, ","))))
	}
	if ingress.Rules == nil || len(ingress.Rules) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Rules"))
	}
	for _, v := range ingress.Rules {
		if v.Path == nil || len(v.Path) == 0 {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Path"))
		}
		for _, p := range v.Path {
			if p.ServiceName == "" {
				errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
			} else if err := api_validation.IsDNS1035Label(p.ServiceName); len(err) > 0 {
				errs = append(errs, errors.New(fmt.Sprintf(invalidName, p.ServiceName, strings.Join(err, ","))))
			}
			if len(api_validation.IsValidPortNum(p.ServicePort)) > 0 {
				errs = append(errs, fmt.Errorf(invalidPort, p.ServicePort, 1, maxport))
			}
			if p.Path == "" {
				errs = append(errs, fmt.Errorf(fieldShouldExist, "Path"))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateIngressFromFile(ingress *api_extensions.Ingress) []error {
	errs := []error{}

	if ingress.Kind != ingressKind {
		errs = append(errs, fmt.Errorf(invalidResourceKind, ingress.Kind, ingressKind))
	}

	if ingress.APIVersion != "" && ingress.APIVersion != ingressApiVersion {
		errs = append(errs, fmt.Errorf(invalidApiVersion, ingress.APIVersion, ingressApiVersion))
	}

	if ingress.GetLabels()[ownerLabel] == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Label: Owner"))
	} else if !IsValidUUID(ingress.GetLabels()[ownerLabel]) {
		errs = append(errs, errors.New(invalidOwner))
	}

	if ingress.Name == "" {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
	} else if err := api_validation.IsDNS1123Subdomain(ingress.Name); len(err) > 0 {
		errs = append(errs, errors.New(fmt.Sprintf(invalidName, ingress.Name, strings.Join(err, ","))))
	}

	if ingress.Spec.Rules == nil || len(ingress.Spec.Rules) == 0 {
		errs = append(errs, fmt.Errorf(fieldShouldExist, "Rules"))
	}
	for _, r := range ingress.Spec.Rules {
		if r.HTTP.Paths == nil || len(r.HTTP.Paths) == 0 {
			errs = append(errs, fmt.Errorf(fieldShouldExist, "Path"))
		}
		for _, p := range r.HTTP.Paths {
			if p.Backend.ServiceName == "" {
				errs = append(errs, fmt.Errorf(fieldShouldExist, "Name"))
			} else if err := api_validation.IsDNS1035Label(p.Backend.ServiceName); len(err) > 0 {
				errs = append(errs, errors.New(fmt.Sprintf(invalidName, p.Backend.ServiceName, strings.Join(err, ","))))
			}
			if len(api_validation.IsValidPortNum(int(p.Backend.ServicePort.IntVal))) > 0 {
				errs = append(errs, fmt.Errorf(invalidPort, p.Backend.ServicePort.IntVal, 1, maxport))
			}
			if p.Path == "" {
				errs = append(errs, fmt.Errorf(fieldShouldExist, "Path"))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
