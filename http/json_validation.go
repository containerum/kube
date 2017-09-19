package http

import (
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
)

// parseJSON parses a JSON payload into a kubernetes struct.
func parseJSON(jsn []byte) (interface{}, string, error) {
	var kind struct {
		Kind string `json:"kind"`
	}

	err := json.Unmarshal(jsn, &kind)
	if err != nil {
		return nil, "", err
	}
	switch kind.Kind {
	case "Namespace":
		var obj *corev1.Namespace
		err = json.Unmarshal(jsn, &obj)
		return obj, kind.Kind, err
	case "Deployment":
		var obj *appsv1beta2.Deployment
		err = json.Unmarshal(jsn, &obj)
		return obj, kind.Kind, err
	case "Service":
		var obj *corev1.Service
		err = json.Unmarshal(jsn, &obj)
		return obj, kind.Kind, err
	case "Endpoints":
		var obj *corev1.Endpoints
		err = json.Unmarshal(jsn, &obj)
		return obj, kind.Kind, err
	}
}
