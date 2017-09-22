package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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

func AddLogger(c *gin.Context) {
	logentry := logrus.NewEntry(log).
		WithField("request-id", c.MustGet("request-id").(string)).
		WithField("client-ip", c.ClientIP())
	c.Set("logger", logentry)
}

func Log(c *gin.Context) *logrus.Entry {
	return c.MustGet("logger").(*logrus.Entry)
}

func AddLogField(c *gin.Context, key string, value interface{}) {
	logentry := Log(c).WithField(key, value)
	c.Set("logger", logentry)
}
