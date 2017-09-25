package main

import (
	"net/http"
	"time"

	"bitbucket.org/exonch/kube-api/router"
	m_server "bitbucket.org/exonch/kube-api/server"
	"bitbucket.org/exonch/kube-api/utils"

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

	//create logger
	log := utils.Logger(c.Bool("debug"))

	//create handler
	handler := router.Load(
		c.Bool("debug"),
		ginrus.Ginrus(log, time.RFC3339, true),
	)

	if c.Bool("debug") {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	}

	//run http server
	return http.ListenAndServe(":1212", handler)
}
