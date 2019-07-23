// ndt7-client is the ndt7 command line client.
//
// Usage:
//
//    ndt7-client [-batch] [-hostname <name>] [-no-verify] [-timeout <string>]
//
// The `-batch` flag causes the command to emit JSON messages on the
// standard output, thus allowing for easy machine parsing. The default
// is to emit user friendly pretty output.
//
// The `-hostname <name>` flag specifies to use the `name` hostname for
// performing the ndt7 test. The default is to auto-discover a suitable
// server by using Measurement Lab's locate service.
//
// The `-no-verify` flag allows to skip TLS certificate verification.
//
// The `-timeout <string>` flag specifies the time after which the
// whole test is interrupted. The `<string>` is a string suitable to
// be passed to time.ParseDuration, e.g., "15s". The default is a large
// enough value that should be suitable for common conditions.
//
// Additionally, passing any unrecognized flag, such as `-help`, will
// cause ndt7-client to print a brief help message.
//
// Events emitted in batch mode
//
// This section describes the events emitted in batch mode. The code
// will always emit a single event per line. In some cases we have
// wrapped long event lines, below, to simplify reading.
//
// When the download subtest starts, this event is emitted:
//
//   {"key":"status.measurement_start","value":{"subtest":"download"}}
//
// After this event is emitted, we discover the server to use (unless it
// has been configured by the user) and we connect to it. If any of these
// operations fail, this event is emitted:
//
//   {"key":"failure.measurement",
//    "value":{"failure":"<failure>","subtest":"download"}}
//
// where `<failure>` is the error that occurred serialized as string. In
// case of failure, the subtest is over and the next event to be emitted is
// `"status.measurement_done"`.
//
// Otherwise, the download subtest starts and we see the following event:
//
//   {"key":"status.measurement_begin",
//    "value":{"server":"<server>","subtest":"download"}}
//
// where `<server>` is the FQDN of the server we're using. Then there
// are zero or more events like:
//
//   {"key": "measurement", "value": <value>}
//
// where `<value>` is a serialized spec.Measurement struct. Note that
// the minimal `<value>` MUST contain a field named `"subtest"` with
// value equal either to `"download"` or `"upload"`.
//
// Finally, this event is always emitted at the end of the subtest:
//
//   {"key":"status.measurement_done","value":{"subtest":"download"}}
//
// The upload subtest is like the download subtest, except for the
// value of the `"subtest"` key.
//
// Exit code
//
// This tool exits with zero on success, nonzero on failure. Under
// some severe internal error conditions, this tool will exit using
// a nonzero exit code without being able to print a diagnostic
// message explaining the error that occurred. In all other cases,
// checking the output should help to understand the error cause.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/emitter"
	"github.com/m-lab/ndt7-client-go/spec"
)

const (
	clientName     = "ndt7-client-go-cmd"
	clientVersion  = "0.1.0"
	defaultTimeout = 55 * time.Second
)

var (
	flagBatch    = flag.Bool("batch", false, "emit JSON events on stdout")
	flagNoVerify = flag.Bool("no-verify", false, "skip TLS certificate verification")
	flagHostname = flag.String("hostname", "", "optional ndt7 server hostname")
	flagTimeout  = flag.Duration(
		"timeout", defaultTimeout, "time after which the test is aborted")
)

type runner struct {
	client  *ndt7.Client
	emitter emitter.Emitter
}

func (r runner) doRunSubtest(
	ctx context.Context, subtest string,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) int {
	ch, err := start(ctx)
	if err != nil {
		r.emitter.OnError(subtest, err)
		return 1
	}
	err = r.emitter.OnConnected(subtest, r.client.FQDN)
	if err != nil {
		return 1
	}
	for ev := range ch {
		err = emitEvent(&ev)
		if err != nil {
			return 1
		}
	}
	return 0
}

func (r runner) runSubtest(
	ctx context.Context, subtest string,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) int {
	// Implementation note: we want to always emit the initial and the
	// final events regardless of how the actual subtest goes. What's more,
	// we want the exit code to be nonzero in case of any error.
	err := r.emitter.OnStarting(subtest)
	if err != nil {
		return 1
	}
	code := r.doRunSubtest(ctx, subtest, start, emitEvent)
	err = r.emitter.OnComplete(subtest)
	if err != nil {
		return 1
	}
	return code
}

func (r runner) runDownload(ctx context.Context) int {
	return r.runSubtest(ctx, "download", r.client.StartDownload,
		r.emitter.OnDownloadEvent)
}

func (r runner) runUpload(ctx context.Context) int {
	return r.runSubtest(ctx, "upload", r.client.StartUpload,
		r.emitter.OnUploadEvent)
}

var osExit = os.Exit

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()
	var r runner
	r.client = ndt7.NewClient(clientName, clientVersion)
	r.client.Dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: *flagNoVerify,
	}
	r.client.FQDN = *flagHostname
	if *flagBatch {
		r.emitter = emitter.NewBatch()
	} else {
		r.emitter = emitter.NewInteractive()
	}
	osExit(r.runDownload(ctx) + r.runUpload(ctx))
}
