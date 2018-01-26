package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
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

	httpsrv := &http.Server{
		Addr:    ":1212",
		Handler: handler,
	}

	// serve connections
	go func() {
		if err := httpsrv.ListenAndServe(); err != nil {
			logrus.WithError(err).Panicln("server start failed")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt) // subscribe on interrupt event
	<-quit                            // wait for event
	logrus.Infoln("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return httpsrv.Shutdown(ctx)
}
