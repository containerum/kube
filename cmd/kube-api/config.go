package main

import (
	"github.com/gin-gonic/gin"
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
		EnvVar: "CH_KUBE_API_PORT",
		Name:   "port",
		Value:  "1212",
		Usage:  "port for kube-api server",
	},
	cli.StringFlag{
		EnvVar: "CH_KUBE_API_KUBE_CONF",
		Name:   "kubeconf",
		Usage:  "config file for kubernetes apiserver client",
	},
	cli.BoolFlag{
		EnvVar: "CH_KUBE_API_TEXTLOG",
		Name:   "textlog",
		Usage:  "output log in text format",
	},
	cli.BoolFlag{
		EnvVar: "CH_KUBE_API_CORS",
		Name:   "cors",
		Usage:  "enable CORS",
	},
	cli.StringFlag{
		EnvVar: "CH_KUBE_PERMISSIONS_ADDR",
		Name:   "permissions_addr",
		Value:  "http://permissions:4242",
		Usage:  "permissions service URL",
	},
}

func setupLogs(c *cli.Context) {
	if c.Bool("debug") {
		gin.SetMode(gin.DebugMode)
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		logrus.SetLevel(logrus.InfoLevel)
	}

	if c.Bool("textlog") {
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}
