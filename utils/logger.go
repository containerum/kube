package utils

import "github.com/sirupsen/logrus"

var log *logrus.Logger

func init() {
	log = logrus.New()
}

func Logger(debug bool) *logrus.Logger {
	if debug {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
		log.Formatter = &logrus.JSONFormatter{}
	}
	return log
}
