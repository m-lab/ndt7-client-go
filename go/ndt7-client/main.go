package main

import (
	"flag"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/m-lab/ndt7-clients/go/ndt7-client/client"
)

var hostname = flag.String("hostname", "localhost", "Host to connect to")
var port = flag.String("port", "443", "Port to connect to")
var insecure = flag.Bool("insecure", false, "Skip TLS verify")

func main() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)
	flag.Parse()
	var clnt client.Client
	clnt.Hostname = *hostname
	clnt.Port = *port
	clnt.Insecure = *insecure
	if err := clnt.Download(); err != nil {
		log.WithError(err).Warn("Download failed")
	}
	if err := clnt.Upload(); err != nil {
		log.WithError(err).Warn("Download failed")
	}
}
