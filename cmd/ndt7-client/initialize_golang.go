// +build !ndt7openssl

package main

import (
	"crypto/tls"

	"github.com/m-lab/ndt7-client-go"
)

// initialize initializes |clnt| to use golang's standard library
// for doing TLS, which is obviously nice and clean.
func initialize(clnt *ndt7.Client) {
	clnt.Scheme = flagScheme.Value
	clnt.Dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: *flagNoVerify,
	}
	clnt.FQDN = *flagHostname
}
