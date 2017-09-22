package main

import (
	"net/http"
	"time"

	"bitbucket.org/exonch/kube-api/router"
	"bitbucket.org/exonch/kube-api/utils"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "CH_KUBE_API_DEBUG",
		Name:   "debug",
		Usage:  "start the server in debug mode",
	},
}

func server(c *cli.Context) error {
	//create logger
	log := utils.Logger(c.Bool("debug"))

	//create handler
	handler := router.Load(
		c.Bool("debug"),
		ginrus.Ginrus(log, time.RFC3339, true),
	)

	//run http server
	return http.ListenAndServe(":1212", handler)
}
