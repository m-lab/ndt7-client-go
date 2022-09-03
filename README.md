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

### Building with a custom client name

In case you are integrating an ndt7-client binary into a third-party
application, it may be useful to build it with a custom client name. Since this
value is passed to the server as metadata, doing so will allow you to retrieve
measurements coming from your custom integration in Measurement Lab's data
easily.

To set a custom client name at build time:

```bash
CLIENTNAME=my-custom-client-name

go build -ldflags "-X main.ClientName=$CLIENTNAME" ./cmd/ndt7-client
```

### Prometheus Exporter

While `ndt7-client` is a "single shot" ndt7 client, there is also a
non-interactive periodic test runner `ndt7-prometheus-exporter`.

#### Build and Run using Docker

```bash
git clone https://github.com/m-lab/ndt7-client-go
docker build -t ndt7-prometheus-exporter .
```

To run tests repeatedly

```bash
PORT=9191
docker run -d -p ${PORT}:8080 ndt7-prometheus-exporter
```

#### Sample Prometheus config

```
# scrape ndt7 test metrics
  - job_name: ndt7
    metrics_path: /metrics
    static_configs:
	  - targets:
	    # host:port of the exporter
	    - localhost:9191

# scrape ndt7-prometheus-exporter itself
  - job_name: ndt7-prometheus-exporter
    static_configs:
	  - targets:
	    # host:port of the exporter
		- localhost:9191
```
