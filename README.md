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

You need Go >= 1.11. We use modules. Make sure that (1) there is
no `GOPATH` and (2) Go < 1.13 uses modules by default. To do
that, do the following in your shell:

```bash
unset GOPATH           # 1.
export GO111MODULE=on  # 2.
```

### Cloning the sources and developing

Make sure you environment is correctly set as explained above. To clone
the sources in the current working directory, run:

```bash
git clone https://github.com/m-lab/ndt7-client-go
```

Then you can enter the directory and do usual Go development.

### Compiling the ndt7-client command on the fly

Make sure you environment is correctly set as explained above. Then
compile on the fly the ndt7-client command with:


```bash
go get -v github.com/m-lab/ndt7-client-go/cmd/ndt7-client
```

This will compile the `ndt7-client` binary and install it into
the `$HOME/go/bin` directory.
