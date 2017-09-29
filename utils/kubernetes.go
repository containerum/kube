package utils

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

func KubeErrorHTTPStatus(e error) int {
	if e == nil {
		return 200
	}
	statusError, ok := e.(*errors.StatusError)
	if !ok {
		return 502
	}
	code := int(statusError.Status().Code)
	// logrus.StandardLogger().Debugf("KubeErrorHTTPStatus: return (%d)", code)
	// logrus.StandardLogger().Debugf("struct: %#v", statusError)
	// logrus.StandardLogger().Debugf("details: %#v", *statusError.Status().Details)
	return code
}
