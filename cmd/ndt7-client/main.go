// ndt7-client is the ndt7 command line client.
//
// Usage:
//
//    ndt7-client [-batch] [-hostname <hostname>] [-timeout <seconds>]
//
// ndt7-client performs a ndt7 nettest.
//
// The `-batch` flag causes the command to emit JSON messages on the
// standard output, thus allowing for easy machine parsing. The default
// is to emit user friendly pretty output.
//
// The `-hostname <hostname>` flag specifies the hostname to use for
// performing the ndt7 test. The default is to auto-discover a suitable
// server by using Measurement Lab's locate service.
//
// The `-timeout <timeout>` flag specifies after how many seconds a
// running ndt7 test should timeout. The default is a large enough
// value that should be suitable for common conditions.
package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

var flagBatch = flag.Bool("batch", false, "emit JSON events on stdout")

var flagHostname = flag.String("hostname", "", "optional ndt7 server hostname")

var flagTimeout = flag.Int64(
	"timeout", 45, "seconds after which the ndt7 test is aborted",
)

func downloadUpload(
	client *ndt7.Client, emitter emitter, subtest string,
	start func() (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement),
) (code int) {
	code = 0
	defer func() {
		if err := recover(); err != nil {
			code = 1
		}
	}()
	emitter.onStarting(subtest)
	ch, err := start()
	if err != nil {
		emitter.onError(subtest, err)
		code = 2
		return
	}
	emitter.onConnected(subtest, client.FQDN)
	for ev := range ch {
		emitEvent(&ev)
	}
	emitter.onComplete(subtest)
	return
}

func download(client *ndt7.Client, emitter emitter) int {
	return downloadUpload(
		client, emitter, "download", client.StartDownload,
		emitter.onDownloadEvent,
	)
}

func upload(client *ndt7.Client, emitter emitter) int {
	return downloadUpload(
		client, emitter, "upload", client.StartUpload,
		emitter.onUploadEvent,
	)
}

func realmain(timeoutSec int64, hostname string, batchmode bool) int {
	timeout := time.Duration(timeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(timeout),
	)
	defer cancel()
	client := ndt7.NewClient(ctx)
	client.FQDN = hostname
	var emitter emitter = interactive{}
	if batchmode {
		emitter = batch{}
	}
	return download(client, emitter) + upload(client, emitter)
}

var osExit = os.Exit

func main() {
	flag.Parse()
	rv := realmain(*flagTimeout, *flagHostname, *flagBatch)
	osExit(rv)
}
