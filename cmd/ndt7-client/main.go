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
	"fmt"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

var flagHostname = flag.String("hostname", "", "optional ndt7 server hostname")

var flagTimeout = flag.Int64(
	"timeout", 45, "seconds after which the ndt7 test is aborted",
)

func onStarting(subtest string) {
	fmt.Printf("\rstarting %s", subtest)
}

func onError(subtest string, err error) {
	fmt.Printf("\r%s failed: %s\n", subtest, err.Error())
}

func onConnected(subtest, fqdn string) {
	fmt.Printf("\r%s in progress with %s\n", subtest, fqdn)
}

func onDownloadEvent(m *spec.Measurement) {
	fmt.Printf(
		"\rMaxBandwidth: %7.1f Mbit/s - RTT: %4.0f/%4.0f/%4.0f (min/smoothed/var) ms",
		float64(m.BBRInfo.MaxBandwidth)/(1000.0*1000.0),
		m.BBRInfo.MinRTT,
		m.TCPInfo.SmoothedRTT,
		m.TCPInfo.RTTVar,
	)
}

func onUploadEvent(m *spec.Measurement) {
	if m.Elapsed > 0.0 {
		v := (8.0 * float64(m.AppInfo.NumBytes)) / m.Elapsed / (1000.0 * 1000.0)
		fmt.Printf("\rAvg. speed  : %7.1f Mbit/s", v)
	}
}

func onComplete(subtest string) {
	fmt.Printf("\n%s: complete\n", subtest)
}

func download(client *ndt7.Client) {
	onStarting("download")
	ch, err := client.StartDownload()
	if err != nil {
		onError("download", err)
		return
	}
	onConnected("download", client.FQDN)
	for ev := range ch {
		onDownloadEvent(&ev)
	}
	onComplete("download")
}

func upload(client *ndt7.Client) {
	onStarting("upload")
	ch, err := client.StartUpload()
	if err != nil {
		onError("upload", err)
		return
	}
	onConnected("upload", client.FQDN)
	for ev := range ch {
		onUploadEvent(&ev)
	}
	onComplete("upload")
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
	download(client)
	upload(client)
}
