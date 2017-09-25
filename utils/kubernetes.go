package utils

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

func KubeErrorHTTPStatus(e error) int {
	if e == nil {
		logrus.StandardLogger().Debugf("KubeErrorHTTPStatus: no error (200)")
		return 200
	}
	statusError, ok := e.(*errors.StatusError)
	if !ok {
		logrus.StandardLogger().Debugf("KubeErrorHTTPStatus: return default (503)")
		return 502
	}
	code := int(statusError.Status().Code)
	logrus.StandardLogger().Debugf("KubeErrorHTTPStatus: return (%d)", code)
	logrus.StandardLogger().Debugf("struct: %#v", statusError)
	logrus.StandardLogger().Debugf("details: %#v", *statusError.Status().Details)
	return code
}
