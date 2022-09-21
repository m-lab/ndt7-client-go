// ndt7-prometheus-exporter is an ndt7 non-interactive prometheus exporting client
//
// Usage:
//
//    ndt7-prometheus-exporter
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
// The `-port` flag starts an HTTP server to export summary results in a form
// that can be consumed by Prometheus (http://prometheus.io).
//
// The `-profile` flag defines the file where to write a CPU profile
// that later you can pass to `go tool pprof`. See https://blog.golang.org/pprof.
//
// Additionally, passing any unrecognized flag, such as `-help`, will
// cause ndt7-client to print a brief help message.
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
	"github.com/m-lab/ndt7-client-go/internal/emitter"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/internal/runner"
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

	flagNoVerify = flag.Bool("no-verify", false, "skip TLS certificate verification")
	flagServer   = flag.String("server", "", "optional ndt7 server hostname")
	flagTimeout  = flag.Duration(
		"timeout", defaultTimeout, "time after which the test is aborted")
	flagQuiet    = flag.Bool("quiet", false, "emit summary and errors only")
	flagService  = flagx.URL{}
	flagUpload   = flag.Bool("upload", true, "perform upload measurement")
	flagDownload = flag.Bool("download", true, "perform download measurement")

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

	e := emitter.NewQuiet(emitter.NewHumanReadable())

	if *flagPort > 0 {
		dlThroughput := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "download_throughput_bps",
				Help: "m-lab ndt7 download speed in bits/s",
			},
			[]string{
				// client IP and remote server
				"client_ip",
				"server_ip",
			})
		prometheus.MustRegister(dlThroughput)
		dlLatency := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "download_latency_seconds",
				Help: "m-lab ndt7 download latency time in seconds",
			},
			[]string{
				// client IP and remote server
				"client_ip",
				"server_ip",
			})
		prometheus.MustRegister(dlLatency)
		ulThroughput := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "upload_throughput_bps",
				Help: "m-lab ndt7 upload speed in bits/s",
			},
			[]string{
				// client IP and remote server
				"client_ip",
				"server_ip",
			})
		prometheus.MustRegister(ulThroughput)
		ulLatency := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ndt7",
				Name: "upload_latency_seconds",
				Help: "m-lab ndt7 upload latency time in seconds",
			},
			[]string{
				// client IP and remote server
				"client_ip",
				"server_ip",
			})
		prometheus.MustRegister(ulLatency)

		// The result gauge captures the result of the last test attempt.
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

		e = emitter.NewPrometheus(e, dlThroughput, dlLatency, ulThroughput, ulLatency, lastResultGauge)
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

	r := runner.New(
		runner.RunnerOptions{
			Download: *flagDownload,
			Upload: *flagUpload,
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

	r.RunTestsInLoop()
}
