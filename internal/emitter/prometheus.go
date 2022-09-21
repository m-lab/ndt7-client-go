package emitter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Prometheus tees summary metrics as prometheus metrics.
// The message is actually emitted by the embedded Emitter.
type Prometheus struct {
	emitter Emitter
	// Download throughput
	// Value: throughput in bits/s
	// Labels: client_ip, server_ip
	dlTp *prometheus.GaugeVec
	// Download latency
	// Value: latency in secs
	// Labels: client_ip, server_ip
	dlLat *prometheus.GaugeVec
	// Upload throughput
	// Value: throughput in bits/s
	// Labels: client_ip, server_ip
	ulTp *prometheus.GaugeVec
	// Upload latency
	// Value: latency in secs
	// Labels: client_ip, server_ip
	ulLat *prometheus.GaugeVec
	// Last results
	// Value: time in seconds since unix epoch
	// labels: test, result
	lastResult *prometheus.GaugeVec
}

// NewPrometheus returns a Summary emitter which emits messages
// via the passed Emitter.
func NewPrometheus(e Emitter, dlThroughput, dlLatency, ulThroughput, ulLatency, lastResult *prometheus.GaugeVec) Emitter {
	return &Prometheus{e, dlThroughput, dlLatency, ulThroughput, ulLatency, lastResult}
}

// OnStarting emits the starting event
func (p Prometheus) OnStarting(test spec.TestKind) error {
	return p.emitter.OnStarting(test)
}

// OnError emits the error event
func (p Prometheus) OnError(test spec.TestKind, err error) error {
	g := p.lastResult.WithLabelValues(string(test), "ERROR")
	g.Set(float64(time.Now().Unix()))
	return p.emitter.OnError(test, err)
}

// OnConnected emits the connected event
func (p Prometheus) OnConnected(test spec.TestKind, fqdn string) error {
	return p.emitter.OnConnected(test, fqdn)
}

// OnDownloadEvent handles an event emitted during the download
func (p Prometheus) OnDownloadEvent(m *spec.Measurement) error {
	return p.emitter.OnDownloadEvent(m)
}

// OnUploadEvent handles an event emitted during the upload
func (p Prometheus) OnUploadEvent(m *spec.Measurement) error {
	return p.emitter.OnUploadEvent(m)
}

// OnComplete is the event signalling the end of the test
func (p Prometheus) OnComplete(test spec.TestKind) error {
	g := p.lastResult.WithLabelValues(string(test), "OK")
	g.Set(float64(time.Now().Unix()))
	return p.emitter.OnComplete(test)
}

// OnSummary handles the summary event, emitted after the test is over.
func (p *Prometheus) OnSummary(s *Summary) error {
	// Note this assumes download and upload throughput units are Mbit/s
	// and latency units are msecs.
	p.dlTp.Reset()
	p.dlTp.WithLabelValues(s.ClientIP, s.ServerIP).Set(s.Download.Throughput.Value * 1000.0 * 1000.0)
	p.dlLat.Reset()
	p.dlLat.WithLabelValues(s.ClientIP, s.ServerIP).Set(s.Download.Latency.Value / 1000.0)
	p.ulTp.Reset()
	p.ulTp.WithLabelValues(s.ClientIP, s.ServerIP).Set(s.Upload.Throughput.Value * 1000.0 * 1000.0)
	p.ulLat.Reset()
	p.ulLat.WithLabelValues(s.ClientIP, s.ServerIP).Set(s.Upload.Latency.Value / 1000.0)

	return p.emitter.OnSummary(s)
}
