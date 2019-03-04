package main

import (
	"crypto/tls"
	"flag"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

var hostname = flag.String("hostname", "localhost", "Host to connect to")
var port = flag.String("port", "443", "Port to connect to")
var insecure = flag.Bool("insecure", false, "Skip TLS verify")

func main() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)
	flag.Parse()
	var clnt Client
	clnt.URL.Scheme = "wss"
	clnt.URL.Host = *hostname + ":" + *port
	if *insecure {
		config := tls.Config{InsecureSkipVerify: true}
		clnt.Dialer.TLSClientConfig = &config
	}
	if err := clnt.Download(); err != nil {
		os.Exit(1)
	}
}
