// ndt7-client is the ndt7 command line client.
//
// Usage:
//
//    ndt7-client [flags]
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
// The `-port` flag starts an HTTP server to export summary results in a form
// that can be consumed by Prometheus (http://prometheus.io).
//
// The `-daemon` flag runs tests repeatedly with rate limiting. It is intended
// to be used together with the `-port` flag.
//
// The `-profile` flag defines the file where to write a CPU profile
// that later you can pass to `go tool pprof`. See https://blog.golang.org/pprof.
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
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/memoryless"
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/emitter"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/runner"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sys/cpu"
)

const (
	defaultTimeout = 55 * time.Second
)

var (
	ClientName    = "ndt7-client-go-cmd"
	ClientVersion = "0.7.0"
	flagProfile   = flag.String("profile", "",
		"file where to store pprof profile (see https://blog.golang.org/pprof)")

	flagScheme = flagx.Enum{
		Options: []string{"wss", "ws"},
		Value:   defaultSchemeForArch(),
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

	flagDaemon = flag.Bool("daemon", false, "run tests in a (rate limited) loop")
	// The flag values below implement rate limiting at the recommended rate
	flagPeriodMean = flag.Duration("period_mean", 6 * time.Hour, "mean period, e.g. 6h, between speed tests, when running in daemon mode")
	flagPeriodMin = flag.Duration("period_min", 36 * time.Minute, "minimum period, e.g. 36m, between speed tests, when running in daemon mode")
	flagPeriodMax = flag.Duration("period_max", 15 * time.Hour, "maximum period, e.g. 15h, between speed tests, when running in daemon mode")

	flagPort = flag.Int("port", 0, "if non-zero, start an HTTP server on this port to export prometheus metrics")
)

func init() {
	flag.Var(
		&flagScheme,
		"scheme",
		`WebSocket scheme to use: either "wss" or "ws"`,
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

// defaultSchemeForArch returns the default WebSocket scheme to use, depending
// on the architecture we are running on. A CPU without native AES instructions
// will perform poorly if TLS is enabled.
func defaultSchemeForArch() string {
	if cpu.ARM64.HasAES || cpu.ARM.HasAES || cpu.X86.HasAES {
		return "wss"
	}
	return "ws"
}

var osExit = os.Exit

func main() {
	flag.Parse()

	if *flagProfile != "" {
		log.Printf("warning: using -profile will reduce the performance")
		fp, err := os.Create(*flagProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(fp)
		defer pprof.StopCPUProfile()
	}

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

	if *flagPort > 0 {
		downloadGauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "download_rate_bps",
				Help: "m-lab ndt7 download speed in bits/s",
			})
		prometheus.MustRegister(downloadGauge)
		uploadGauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "upload_rate_bps",
				Help: "m-lab ndt7 upload speed in bits/s",
			})
		prometheus.MustRegister(uploadGauge)
		rttGauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "rtt_seconds",
				Help: "m-lab ndt7 round-trip time in seconds",
			})
		prometheus.MustRegister(rttGauge)

		// The result gauge captures the result of the last test attemp.
		//
		// Since its value is a timestamp, the following PromQL expression will
		// give the most recent result for each upload and download test.
		//
		//     time() - topk(1, ndt7_result_timestamp_seconds) without (result)
		lastResultGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "result_timestamp_seconds",
				Help: "m-lab ndt7 test completion time in seconds since 1970-01-01",
			},
			[]string{
				// which test completed
				"test",
				// test result
				"result",
			})
		prometheus.MustRegister(lastResultGauge)

		// The success gauge captures both the client IP and the server FQDN.
		//
		// Since its value is a timestamp, we can use it to determine the last
		// client-server pair that successfully ran the tests. With some PromQL
		// trickery, it is possible to join the client-server labels with test
		// results.
		//
		// For example:
		//
		// - last download test result with client-server labels
		//
		//   ndt7_download_rate_bps + on () group_left(client, server)
		//   0 * topk(1, ndt7_last_success_timestamp_seconds) without (client, server)
		//
		// - last upload test result with client-server labels
		//
		//   ndt7_upload_rate_bps + on () group_left(client, server)
		//   0 * topk(1, ndt7_last_success_timestamp_seconds) without (client, server)
		//
		// - last rtt test result with client-server labels
		//
		//   ndt7_rtt_seconds + on () group_left(client, server)
		//   0 * topk(1, ndt7_last_success_timestamp_seconds) without (client, server)
		//
		lastSuccessGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "last_success_timestamp_seconds",
				Help: "last successful m-lab ndt7 test completion time in seconds since 1970-01-01",
			},
			[]string{
				// client IP and remote server
				"client",
				"server",
			})
		prometheus.MustRegister(lastSuccessGauge)
		e = emitter.NewPrometheus(e, downloadGauge, uploadGauge, rttGauge, lastResultGauge, lastSuccessGauge)
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *flagPort), nil))
		}()
	}

	ticker, err := memoryless.NewTicker(
		context.Background(),
		memoryless.Config{
			Expected: *flagPeriodMean,
			Min: *flagPeriodMin,
			Max: *flagPeriodMax,
		})
	if err != nil {
		log.Fatalf("Failed to create memoryless.Ticker: %v", err)
	}
	defer ticker.Stop()

	r := runner.NewRunner(
		runner.RunnerOptions{
			Download: *flagDownload,
			Upload: *flagUpload,
			Daemon: *flagDaemon,
			Timeout: *flagTimeout,
			ClientFactory: func() *ndt7.Client {
				c := ndt7.NewClient(ClientName, ClientVersion)
				c.ServiceURL = flagService.URL
				c.Server = *flagServer
				c.Scheme = flagScheme.Value
				c.Dialer.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: *flagNoVerify,
				}

				return c
			},
		},
		e,
		ticker)

	osExit(r.RunTestsInLoop())
}
