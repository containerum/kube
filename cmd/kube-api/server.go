package main

import (
	"net/http"
	"time"

	"git.containerum.net/ch/kube-api/router"
	m_server "git.containerum.net/ch/kube-api/server"
	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "CH_KUBE_API_DEBUG",
		Name:   "debug",
		Usage:  "start the server in debug mode",
	},
	cli.StringFlag{
		EnvVar: "CH_KUBE_API_KUBE_CONF",
		Name:   "kubeconf",
		Usage:  "config file for kubernetes apiserver client",
	},
}

func server(c *cli.Context) error {
	m_server.LoadKubeClients(c.String("kubeconf"))

	//setup logger
	if c.Bool("debug") {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	}

	//create handler
	handler := router.Load(
		c.Bool("debug"),
		gin.RecoveryWithWriter(logrus.WithField("component", "gin_recovery").WriterLevel(logrus.ErrorLevel)),
		ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true),
	)

	//run http server
	return http.ListenAndServe(":1212", handler)
}
