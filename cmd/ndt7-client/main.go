// ndt7-client is the ndt7 command line client.
//
// Usage:
//
//    ndt7-client [-format <human|json>] [-server <name>] [-no-verify]
//                [-scheme <scheme>] [-timeout <string>] [-service-url <url>]
//                [-upload <bool>] [-download <bool>]
//
// The `-format` flag defines how the output should be emitter. Possible
// values are "human", which is the default, and "json", where each message
// is a valid JSON object.
//
// The (DEPRECATED) `-batch` flag is equivalent to `-format json`, and the
// latter should be used instead.
//
// The default behavior is for ndt7-client to discover a suitable server using
// Measurement Lab's locate service. This behavior may be overridden using
// either the `-server` or `-service-url` flags.
//
// The `-server <name>` flag specifies the server `name` for performing
// the ndt7 test. This option overrides `-service-url`.
//
// The `-service-url <url>` flag specifies a complete URL that specifies the
// scheme (e.g. "ws"), server name and port, protocol (e.g. /ndt/v7/download),
// and HTTP parameters. By default, upload and download measurements are run
// automatically. The `-service-url` specifies only one measurement direction.
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
// The `-upload` and `-download` flags are boolean options that default to true,
// but may be set to false on the command line to run only upload or only
// download.
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
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/emitter"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/spec"
)

const (
	clientName     = "ndt7-client-go-cmd"
	clientVersion  = "0.4.1"
	defaultTimeout = 55 * time.Second
)

var (
	flagScheme = flagx.Enum{
		Options: []string{"wss", "ws"},
		Value:   "wss",
	}
	flagFormat = flagx.Enum{
		Options: []string{"human", "json"},
		Value:   "human",
	}

	flagBatch = flag.Bool("batch", false, "emit JSON events on stdout "+
		"(DEPRECATED, please use -format=json)")
	flagNoVerify = flag.Bool("no-verify", false, "skip TLS certificate verification")
	flagServer   = flag.String("server", "", "optional ndt7 server hostname")
	flagTimeout  = flag.Duration(
		"timeout", defaultTimeout, "time after which the test is aborted")
	flagQuiet    = flag.Bool("quiet", false, "emit summary and errors only")
	flagService  = flagx.URL{}
	flagUpload   = flag.Bool("upload", true, "perform upload measurement")
	flagDownload = flag.Bool("download", true, "perform download measurement")
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
		"output format to use: 'human' or 'json' for batch processing",
	)
	flag.Var(
		&flagService,
		"service-url",
		"Service URL specifies target hostname and other URL fields like access token. Overrides -server.",
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
		r.emitter.OnError(test, err)
		return 1
	}
	err = r.emitter.OnConnected(test, r.client.FQDN)
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
		return 1
	}
	code := r.doRunTest(ctx, test, start, emitEvent)
	err = r.emitter.OnComplete(test)
	if err != nil {
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
	// test. DownloadRetrans and MinRTT come from the server-side Measurement,
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
			s.MinRTT = emitter.ValueUnitPair{
				Value: float64(dl.Server.TCPInfo.MinRTT) / 1000,
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

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()
	var r runner

	// If a service URL is given, then only one direction is possible.
	if flagService.URL != nil && strings.Contains(flagService.URL.Path, params.DownloadURLPath) {
		*flagUpload = false
		*flagDownload = true
	} else if flagService.URL != nil && strings.Contains(flagService.URL.Path, params.UploadURLPath) {
		*flagUpload = true
		*flagDownload = false
	} else if flagService.URL != nil {
		fmt.Println("WARNING: ignoring unsupported service url")
		flagService.URL = nil
	}

	r.client = ndt7.NewClient(clientName, clientVersion)
	r.client.ServiceURL = flagService.URL
	r.client.Server = *flagServer
	r.client.Scheme = flagScheme.Value
	r.client.Dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: *flagNoVerify,
	}

	var e emitter.Emitter

	// If -batch, force -format=json.
	if *flagBatch || flagFormat.Value == "json" {
		e = emitter.NewJSON(os.Stdout)
	} else {
		e = emitter.NewHumanReadable()
	}
	if *flagQuiet {
		e = emitter.NewQuiet(e)
	}
	r.emitter = e

	var code int
	if *flagDownload {
		code += r.runDownload(ctx)
	}
	if *flagUpload {
		code += r.runUpload(ctx)
	}
	if code != 0 {
		osExit(code)
	}

	s := makeSummary(r.client.FQDN, r.client.Results())
	r.emitter.OnSummary(s)
}
