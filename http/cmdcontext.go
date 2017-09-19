package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"bitbucket.org/exonch/ch-kube-api/server"
)

type cmdContext struct {
	*gin.Context
	cmdID string
	log   logrus.FieldLogger
	body  []byte

	// Server instance, used for pulling out various initialized & configured
	// 3rd party API clients.
	server server.Server
}
