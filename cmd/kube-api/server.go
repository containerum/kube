package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/router"
	log "github.com/sirupsen/logrus"
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
	if c.Bool("debug") {
		//log.SetFormatter(&log.TextFormatter{})
		log.SetLevel(log.DebugLevel)
	}

	kube := kubernetes.Kube{}
	kube.RegisterClient(c.String("kubeconf"))

	app := router.CreateRouter(&kube, c.Bool("debug"))

	// for graceful shutdown
	httpsrv := &http.Server{
		Addr:    ":1212",
		Handler: app,
	}

	// serve connections
	go func() {
		if err := httpsrv.ListenAndServe(); err != nil {
			log.WithError(err).Error("http server start failed")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt) // subscribe on interrupt event
	<-quit                            // wait for event
	log.Infoln("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return httpsrv.Shutdown(ctx)
}
