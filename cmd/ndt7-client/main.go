// ndt7-client is the ndt7 command line client.
//
// Usage:
//
//    ndt7-client [-format <human|json>] [-hostname <name>] [-no-verify]
//                [-scheme <scheme>] [-timeout <string>]
//
// The `-format` flag defines how the output should be emitter. Possible
// values are "human", which is the default, and "json", where each message
// is a valid JSON object.
//
// The (DEPRECATED) `-batch` flag is equivalent to `-format json`, and the
// latter should be used instead.
//
// The `-hostname <name>` flag specifies to use the `name` hostname for
// performing the ndt7 test. The default is to auto-discover a suitable
// server by using Measurement Lab's locate service.
//
// The `-no-verify` flag allows to skip TLS certificate verification.
//
// The `-scheme <scheme>` flag allows to override the default scheme, i.e.,
// "wss", with another scheme. The only other supported scheme is "ws"
// and causes ndt7 to run unencrypted.
//
// The `-timeout <string>` flag specifies the time after which the
// whole test is interrupted. The `<string>` is a string suitable to
// be passed to time.ParseDuration, e.g., "15s". The default is a large
// enough value that should be suitable for common conditions.
//
// Additionally, passing any unrecognized flag, such as `-help`, will
// cause ndt7-client to print a brief help message.
//
// JSON events emitted when -format=json
//
// This section describes the events emitted when using the json output format.
// The code will always emit a single event per line.
//
// When the download test starts, this event is emitted:
//
//   {"Key":"starting","Value":{"Test":"download"}}
//
// After this event is emitted, we discover the server to use (unless it
// has been configured by the user) and we connect to it. If any of these
// operations fail, this event is emitted:
//
//   {"Key":"error","Value":{"Failure":"<failure>","Test":"download"}}
//
// where `<failure>` is the error that occurred serialized as string. In
// case of failure, the test is over and the next event to be emitted is
// `"complete"`
//
// Otherwise, the download test starts and we see the following event:
//
//   {"Key":"connected","Value":{"Server":"<server>","Test":"download"}}
//
// where `<server>` is the FQDN of the server we're using. Then there
// are zero or more events like:
//
//   {"Key": "measurement","Value": <value>}
//
// where `<value>` is a serialized spec.Measurement struct.
//
// Finally, this event is always emitted at the end of the test:
//
//   {"Key":"complete","Value":{"Test":"download"}}
//
// The upload test is like the download test, except for the
// value of the `"Test"` key.
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
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/m-lab/go/flagx"
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
	flagScheme = flagx.Enum{
		Options: []string{"wss", "ws"},
		Value:   "wss",
	}
	flagFormat = flagx.Enum{
		Options: []string{"human", "json", "prometheus"},
		Value:   "human",
	}

	flagBatch = flag.Bool("batch", false, "emit JSON events on stdout "+
		"(DEPRECATED, please use -format=json)")
	flagNoVerify = flag.Bool("no-verify", false, "skip TLS certificate verification")
	flagHostname = flag.String("hostname", "", "optional ndt7 server hostname")
	flagTimeout  = flag.Duration(
		"timeout", defaultTimeout, "time after which the test is aborted")
	flagQuiet     = flag.Bool("quiet", false, "emit summary and errors only")
	listenAddress = flag.String("listen-address", ":9122", "Address to listen to server prometheus metrics.")
	metricsPath   = flag.String("metrics-path", "/metrics", "Path under which to expose prometheus metrics.")
)

func init() {
	flag.Var(
		&flagScheme,
		"scheme",
		`WebSocket scheme to use: either "wss" (the default) or "ws"`,
	)
	flag.Var(
		&flagFormat,
		"format",
		"output format to use: 'human', 'json' or 'prometheus' for batch processing",
	)
}

type runner struct {
	client  *ndt7.Client
	emitter emitter.Emitter
}

func (r runner) doRunTest(
	ctx context.Context, test spec.TestKind,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) int {
	ch, err := start(ctx)
	if err != nil {
		log.Printf("doRunTest start error: %v", err)
		r.emitter.OnError(test, err)
		return 1
	}
	err = r.emitter.OnConnected(test, r.client.FQDN)
	if err != nil {
		log.Printf("doRunTest OnConnected error: %v", err)
		return 1
	}
	for ev := range ch {
		err = emitEvent(&ev)
		if err != nil {
			log.Printf("doRunTest emitEvent error: %v", err)
			return 1
		}
	}
	return 0
}

func (r runner) runTest(
	ctx context.Context, test spec.TestKind,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) int {
	// Implementation note: we want to always emit the initial and the
	// final events regardless of how the actual test goes. What's more,
	// we want the exit code to be nonzero in case of any error.
	err := r.emitter.OnStarting(test)
	if err != nil {
		log.Printf("runTest got error %s:  %s", r.client.ClientName, err)
		return 1
	}
	code := r.doRunTest(ctx, test, start, emitEvent)
	err = r.emitter.OnComplete(test)
	if err != nil {
		log.Printf("runTest got error in client %s: %s", r.client.ClientName, err)
		return 1
	}
	return code
}

func (r runner) runDownload(ctx context.Context) int {
	return r.runTest(ctx, spec.TestDownload, r.client.StartDownload,
		r.emitter.OnDownloadEvent)
}

func (r runner) runUpload(ctx context.Context) int {
	return r.runTest(ctx, spec.TestUpload, r.client.StartUpload,
		r.emitter.OnUploadEvent)
}

func makeSummary(FQDN string, results map[spec.TestKind]*ndt7.LatestMeasurements) *emitter.Summary {

	s := emitter.NewSummary(FQDN)

	if results[spec.TestDownload] != nil &&
		results[spec.TestDownload].ConnectionInfo != nil {
		// Get UUID, ClientIP and ServerIP from ConnectionInfo.
		s.DownloadUUID = results[spec.TestDownload].ConnectionInfo.UUID

		clientIP, _, err := net.SplitHostPort(results[spec.TestDownload].ConnectionInfo.Client)
		if err == nil {
			s.ClientIP = clientIP
		}

		serverIP, _, err := net.SplitHostPort(results[spec.TestDownload].ConnectionInfo.Server)
		if err == nil {
			s.ServerIP = serverIP
		}
	}

	// Download comes from the client-side Measurement during the download
	// test. DownloadRetrans and RTT come from the server-side Measurement,
	// if it includes a TCPInfo object.
	if dl, ok := results[spec.TestDownload]; ok {
		if dl.Client.AppInfo != nil && dl.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(dl.Client.AppInfo.ElapsedTime) / 1e06
			s.Download = emitter.ValueUnitPair{
				Value: (8.0 * float64(dl.Client.AppInfo.NumBytes)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
		if dl.Server.TCPInfo != nil {
			if dl.Server.TCPInfo.BytesSent > 0 {
				s.DownloadRetrans = emitter.ValueUnitPair{
					Value: float64(dl.Server.TCPInfo.BytesRetrans) / float64(dl.Server.TCPInfo.BytesSent) * 100,
					Unit:  "%",
				}
			}
			s.RTT = emitter.ValueUnitPair{
				Value: float64(dl.Server.TCPInfo.RTT) / 1000,
				Unit:  "ms",
			}
		}
	}
	// Upload comes from the client-side Measurement during the upload test.
	if ul, ok := results[spec.TestUpload]; ok {
		if ul.Client.AppInfo != nil && ul.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(ul.Client.AppInfo.ElapsedTime) / 1e06
			s.Upload = emitter.ValueUnitPair{
				Value: (8.0 * float64(ul.Client.AppInfo.NumBytes)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
	}

	return s
}

var osExit = os.Exit

func runWithRetry(ctx context.Context, f func(c context.Context) int) int {
	result := -1
	max := 10
	for i := 0; result != 0 && i < max; i++ {
		result = f(ctx)
		if i > 0 && result != 0 {
			log.Printf("Retry #%d during '%v', error code: %d", i, runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), result)
			time.Sleep(1 * time.Second)
		}
	}
	return result
}

func prommain() {

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
		defer cancel()

		var r runner
		r.client = ndt7.NewClient(clientName, clientVersion)
		r.client.Scheme = flagScheme.Value
		r.client.Dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: *flagNoVerify,
		}
		r.client.FQDN = *flagHostname
		r.emitter = emitter.NewPrometheusExporterWithWriter(w)
		log.Printf("Got request to %s from %s, starting speed test", req.RequestURI, req.RemoteAddr)
		codedown := runWithRetry(ctx, r.runDownload)
		codeup := runWithRetry(ctx, r.runUpload)
		if codedown+codeup != 0 {
			if codedown != 0 {
				log.Printf("Got error during download test of code %d", codedown)
			} else {
				log.Printf("Got error during upload test of code %d", codeup)
			}
			osExit(codeup + codedown)
		}
		s := makeSummary(r.client.FQDN, r.client.Results())
		r.emitter.OnSummary(s)
		log.Printf("Speed test finished %0.2f / %0.2f at %s", s.Download.Value, s.Upload.Value, s.ServerFQDN)
	})

	log.Printf("Starting server at %s", *listenAddress)
	log.Print(http.ListenAndServe(*listenAddress, nil))
}

func main() {
	flag.Parse()

	if flagFormat.Value == "prometheus" {
		prommain()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
		defer cancel()
		var r runner
		r.client = ndt7.NewClient(clientName, clientVersion)
		r.client.Scheme = flagScheme.Value
		r.client.Dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: *flagNoVerify,
		}
		r.client.FQDN = *flagHostname

		var e emitter.Emitter

		// If -batch, force -format=json.
		if *flagBatch || flagFormat.Value == "json" {
			e = emitter.NewJSON(os.Stdout)
		} else if flagFormat.Value == "prometheus" {
			e = emitter.NewPrometheusExporter()
		} else {
			e = emitter.NewHumanReadable()
		}
		if *flagQuiet {
			e = emitter.NewQuiet(e)
		}
		r.emitter = e

		code := r.runDownload(ctx) + r.runUpload(ctx)
		if code != 0 {
			osExit(code)
		}

		s := makeSummary(r.client.FQDN, r.client.Results())
		r.emitter.OnSummary(s)
	}
}
