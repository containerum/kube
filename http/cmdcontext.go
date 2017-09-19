package http

import (
	"bitbucket.org/exonch/ch-kube-api/server"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type cmdContext struct {
	*gin.Context
	cmdID   string
	log     logrus.FieldLogger
	rawbody []byte
	body    interface{}

	// Server instance, used for pulling out various initialized & configured
	// 3rd party API clients.
	server server.Server
}

func (c *cmdContext) ErrorJSON(code int, errstr string) {
	c.Context.JSON(code, map[string]string{"error": errstr})
}
