package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

//go:generate protoc --go_out=../../proto -I../../proto exec.proto
//go:generate swagger flatten ../../swagger-basic.yml -o ../../swagger-basic.json
//go:generate swagger generate spec -m -i ../../swagger-basic.json -o ../../swagger.json
//go:generate swagger flatten ../../swagger.json -o ../../swagger.json
//go:generate swagger validate ../../swagger.json

func main() {
	app := cli.NewApp()
	app.Name = "ch-kube-api-server"
	app.Usage = "Kube api server for Container Hosting"
	app.Flags = flags

	fmt.Printf("Starting %v %v\n", app.Name, app.Version)

	app.Action = initServer

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
