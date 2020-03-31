[![GoDoc](https://godoc.org/github.com/m-lab/ndt7-client-go?status.svg)](https://godoc.org/github.com/m-lab/ndt7-client-go) [![Build Status](https://travis-ci.org/m-lab/ndt7-client-go.svg?branch=master)](https://travis-ci.org/m-lab/ndt7-client-go) [![Coverage Status](https://coveralls.io/repos/github/m-lab/ndt7-client-go/badge.svg?branch=master)](https://coveralls.io/github/m-lab/ndt7-client-go?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/m-lab/ndt7-client-go)](https://goreportcard.com/report/github.com/m-lab/ndt7-client-go)

# ndt7 Go client

Reference ndt7 Go client implementation. Useful resources:

- [API exposed by this library](
    https://godoc.org/github.com/m-lab/ndt7-client-go
);

- [Manual for the ndt7-client CLI program](
    https://godoc.org/github.com/m-lab/ndt7-client-go/cmd/ndt7-client
);

- [ndt7 protocol specification](
    https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md
).

The master branch contains stable code. We don't promise we won't break
the API, but we'll try not to.

## Installing

You need Go >= 1.12. We use modules. Make sure Go knows that:

```bash
export GO111MODULE=on
```

Clone the repository wherever you want with

```bash
git clone https://github.com/m-lab/ndt7-client-go
```

From inside the repository, use `go get ./cmd/ndt7-client` to
build the client. Binaries will be placed in `$GOPATH/bin`, if
`GOPATH` is set, and in `$HOME/go/bin` otherwise.

If you're into a one-off install, this

```bash
go get -v github.com/m-lab/ndt7-client-go/cmd/ndt7-client
```

is equivalent to cloning the repository, running `go get ./cmd/ndt7-client`,
and then cancelling the repository directory.


### Prometheus exporter

If you start the client using the flag `-format=prometheus`, then a http server will be started that runs a speed test evertime the exposed handler on http://localhost/metrics is called. The results will be shown in a format that is readable by [Prometheus](https://prometheus.io), so that you can run the tests freqeuently, automated and collect the results. This can be visualized using [Grafana](https://grafana.com), for example.
