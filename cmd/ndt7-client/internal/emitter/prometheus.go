package emitter

import (
	"fmt"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

// PrometheusExporter is a human readable emitter. It emits the events generated
// by running a ndt7 test as pleasant stdout messages.
type PrometheusExporter struct {
	out io.Writer
}

// NewPrometheusExporter returns a new human readable emitter.
func NewPrometheusExporter() Emitter {
	return PrometheusExporter{os.Stdout}
}

// NewPrometheusExporterWithWriter returns a new human readable emitter using the
// specified writer.
func NewPrometheusExporterWithWriter(w io.Writer) Emitter {
	return PrometheusExporter{w}
}

// OnStarting handles the start event
func (h PrometheusExporter) OnStarting(test spec.TestKind) error {
	// _, err := fmt.Fprintf(h.out, "\rstarting %s", test)
	// return err
	return nil
}

// OnError handles the error event
func (h PrometheusExporter) OnError(test spec.TestKind, err error) error {
	_, failure := fmt.Fprintf(h.out, "\r%s failed: %s\n", test, err.Error())
	return failure
}

// OnConnected handles the connected event
func (h PrometheusExporter) OnConnected(test spec.TestKind, fqdn string) error {
	// _, err := fmt.Fprintf(h.out, "\r%s in progress with %s\n", test, fqdn)
	// return err
	return nil
}

// OnDownloadEvent handles an event emitted by the download test
func (h PrometheusExporter) OnDownloadEvent(m *spec.Measurement) error {
	return h.onSpeedEvent(m)
}

// OnUploadEvent handles an event emitted during the upload test
func (h PrometheusExporter) OnUploadEvent(m *spec.Measurement) error {
	return h.onSpeedEvent(m)
}

func (h PrometheusExporter) onSpeedEvent(m *spec.Measurement) error {
	// The specification recommends that we show application level
	// measurements. Let's just do that in interactive mode. To this
	// end, we ignore any measurement coming from the server.
	// if m.Origin != spec.OriginClient {
	// 	return nil
	// }
	// if m.AppInfo == nil || m.AppInfo.ElapsedTime <= 0 {
	// 	return errors.New("Missing m.AppInfo or invalid m.AppInfo.ElapsedTime")
	// }
	// elapsed := float64(m.AppInfo.ElapsedTime) / 1e06
	// v := (8.0 * float64(m.AppInfo.NumBytes)) / elapsed / (1000.0 * 1000.0)
	// _, err := fmt.Fprintf(h.out, "\rAvg. speed  : %7.1f Mbit/s", v)
	// return err
	return nil
}

// OnComplete handles the complete event
func (h PrometheusExporter) OnComplete(test spec.TestKind) error {
	// _, err := fmt.Fprintf(h.out, "\n%s: complete\n", test)
	// return err
	return nil
}

// # HELP m-lab_download Download bandwidth (Mbps).
// # TYPE m-lab_download gauge
// m-lab_download 51.77325110921138
// # HELP m-lab_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which m-lab_exporter was built.
// # TYPE m-lab_exporter_build_info gauge
// m-lab_exporter_build_info{branch="",goversion="go1.12.9",revision="",version=""} 1.0
// # HELP m-lab_ping Latency (ms)
// # TYPE m-lab_ping gauge
// m-lab_ping 33.138479
// # HELP m-lab_upload Upload bandwidth (Mbps).
// # TYPE m-lab_upload gauge
// m-lab_upload 21.660442480282065

// OnSummary handles the summary event.
func (h PrometheusExporter) OnSummary(s *Summary) error {
	const summaryFormat = `# HELP m-lab_download Download bandwidth (%s).
# TYPE m-lab_download gauge
m-lab_download %7.1f
# HELP m-lab_ping Latency (%s)
# TYPE m-lab_ping gauge
m-lab_ping %7.1f
# HELP m-lab_upload Upload bandwidth (%s).
# TYPE m-lab_upload gauge
m-lab_upload %7.1f
# HELP m-lab_servername server name.
# TYPE m-lab_servername text
m-lab_servername %s
# HELP m-lab_clientip client IP.
# TYPE m-lab_clientip text
m-lab_clientip %s
`

	_, err := fmt.Fprintf(h.out, summaryFormat,
		s.Download.Unit,
		s.Download.Value,
		s.RTT.Unit,
		s.RTT.Value,
		s.Upload.Unit,
		s.Upload.Value,
		s.ServerFQDN,
		s.ClientIP)

	if err != nil {
		return err
	}

	return nil
}
