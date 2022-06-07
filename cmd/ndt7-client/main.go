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
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/emitter"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/limiter"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/spec"
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
	flagPeriodMean = flag.Int("period_mean", 21600, "mean period (in seconds) between speed tests, when running in daemon mode")
	flagPeriodMin = flag.Int("period_min", 2160, "minimum period (in seconds) between speed tests, when running in daemon mode")
	flagPeriodMax = flag.Int("period_max", 54000, "maximum period (in seconds) between speed tests, when running in daemon mode")

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

type runnerOptions struct {
	download, upload bool
	daemon bool
	timeout time.Duration
	clientFactory func() *ndt7.Client
}

type runner struct {
	client  *ndt7.Client
	emitter emitter.Emitter
	limiter limiter.Limiter
	opt     runnerOptions
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

// defaultSchemeForArch returns the default WebSocket scheme to use, depending
// on the architecture we are running on. A CPU without native AES instructions
// will perform poorly if TLS is enabled.
func defaultSchemeForArch() string {
	if cpu.ARM64.HasAES || cpu.ARM.HasAES || cpu.X86.HasAES {
		return "wss"
	}
	return "ws"
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
	// The upload rate comes from the receiver (the server). Currently
	// ndt-server only provides network-level throughput via TCPInfo.
	// TODO: Use AppInfo for application-level measurements when available.
	if ul, ok := results[spec.TestUpload]; ok {
		if ul.Server.TCPInfo != nil && ul.Server.TCPInfo.BytesReceived > 0 {
			elapsed := float64(ul.Server.TCPInfo.ElapsedTime) / 1e06
			s.Upload = emitter.ValueUnitPair{
				Value: (8.0 * float64(ul.Server.TCPInfo.BytesReceived)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
	}

	return s
}

func (r runner) runTests() int {
	var code int
	for ;; {
		code = 0

		func() {
			ctx, cancel := context.WithTimeout(context.Background(), r.opt.timeout)
			defer cancel()

			r.client = r.opt.clientFactory()

			if r.opt.download {
				code += r.runDownload(ctx)
			}
			if r.opt.upload {
				code += r.runUpload(ctx)
			}
		}()

		s := makeSummary(r.client.FQDN, r.client.Results())
		r.emitter.OnSummary(s)

		if !r.opt.daemon {
			break
		}

		r.limiter.Wait()
	}

	return code
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

	r := runner{
		opt: runnerOptions{
			download: *flagDownload,
			upload: *flagUpload,
			daemon: *flagDaemon,
			timeout: *flagTimeout,
			clientFactory: func() *ndt7.Client {
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
		downloadGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mlab_ndt7_download",
				Help: "m-lab ndt7 download speed in Mbit/s",
			},
			[]string{"client_ip"})
		prometheus.MustRegister(downloadGauge)
		uploadGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mlab_ndt7_upload",
				Help: "m-lab ndt7 upload speed in Mbit/s",
			},
			[]string{"client_ip"})
		prometheus.MustRegister(uploadGauge)
		rttGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mlab_ndt7_rtt",
				Help: "m-lab ndt7 round-trip time in ms",
			},
			[]string{"client_ip"})
		prometheus.MustRegister(rttGauge)
		// Prometheus query to compute time since last test completion with result
		//
		//     time() - topk(1, mlab_ndt_completion_timestamp) without (result)
		completionTimeGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mlab_ndt7_completion_timestamp",
				Help: "m-lab ndt7 test completion time in seconds since 1970-01-01",
			},
			[]string{
				// which test completed
				"test",
				// test result
				"result",
			})
		prometheus.MustRegister(completionTimeGauge)
		e = emitter.NewPrometheus(e, downloadGauge, uploadGauge, rttGauge, completionTimeGauge)
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *flagPort), nil))
		}()
	}

	r.emitter = e

	r.limiter = limiter.NewPoissonLimiter(
			float64(*flagPeriodMean), float64(*flagPeriodMin), float64(*flagPeriodMax),
			func(d time.Duration) {
				log.Printf("Waiting until %s", (time.Now().Add(d)).Format(time.RFC3339))
			})

	osExit(r.runTests())
}
