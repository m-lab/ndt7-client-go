// +build ndt7openssl

package main

import (
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/openssl"
)

// initialize initializes |clnt| for using OpenSSL. Because this is a
// rather non standard config, this is also not super clean.
func initialize(clnt *ndt7.Client) {
	clnt.Scheme = "ws" // even with TLS force websocket to not do TLS
	dialer := openssl.NewDialer()
	switch flagScheme.Value {
	case "ws":
	case "wss":
		dialer.InsecureSkipVerify = *flagNoVerify
		clnt.Dialer.NetDial = dialer.Dial
		clnt.Dialer.NetDialContext = dialer.DialContext
	}
	clnt.FQDN = *flagHostname
}
