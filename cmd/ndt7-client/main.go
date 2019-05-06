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
	"time"

	"github.com/m-lab/ndt7-client-go"
)

var flagHostname = flag.String("hostname", "", "optional ndt7 server hostname")

var flagTimeout = flag.Int64(
	"timeout", 45, "seconds after which the ndt7 test is aborted",
)

func download(client *ndt7.Client, emitter emitter) {
	emitter.onStarting("download")
	ch, err := client.StartDownload()
	if err != nil {
		emitter.onError("download", err)
		return
	}
	emitter.onConnected("download", client.FQDN)
	for ev := range ch {
		emitter.onDownloadEvent(&ev)
	}
	emitter.onComplete("download")
}

func upload(client *ndt7.Client, emitter emitter) {
	emitter.onStarting("upload")
	ch, err := client.StartUpload()
	if err != nil {
		emitter.onError("upload", err)
		return
	}
	emitter.onConnected("upload", client.FQDN)
	for ev := range ch {
		emitter.onUploadEvent(&ev)
	}
	emitter.onComplete("upload")
}

func main() {
	flag.Parse()
	// TODO(bassosimone): implement -batch
	timeout := time.Duration(*flagTimeout) * time.Second
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(timeout),
	)
	defer cancel()
	client := ndt7.NewClient(ctx)
	client.FQDN = *flagHostname
	emitter := interactive{}
	download(client, emitter)
	upload(client, emitter)
}
