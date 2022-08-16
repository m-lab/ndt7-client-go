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
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/internal/emitter"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/internal/runner"
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
		nil)

	osExit(len(r.RunTestsOnce()))
}
