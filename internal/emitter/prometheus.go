package emitter

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Prometheus tees summary metrics as prometheus metrics.
// The message is actually emitted by the embedded Emitter.
type Prometheus struct {
	emitter Emitter
	// Last download, upload speed in bits/s
	download, upload prometheus.Gauge
	// Last RTT in seconds
	rtt prometheus.Gauge
	// Last results
	// Value: time in seconds since unix epoch
	// labels: test, result
	lastResult *prometheus.GaugeVec
	// Last successful test
	// Value: time in seconds since unix epoch
	// labels: client_ip, server
	lastSuccess *prometheus.GaugeVec
}

// NewPrometheus returns a Summary emitter which emits messages
// via the passed Emitter.
func NewPrometheus(e Emitter, download, upload, rtt prometheus.Gauge, lastResult, lastSuccess *prometheus.GaugeVec) Emitter {
	return &Prometheus{e, download, upload, rtt, lastResult, lastSuccess}
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
	// Note this assumes download and upload test result units are Mbit/s.
	p.download.Set(s.Download.Value * 1000.0 * 1000.0)
	p.upload.Set(s.Upload.Value * 1000.0 * 1000.0)

	// Note this assumes RTT units are millisecs
	p.rtt.Set(s.MinRTT.Value / 1000.0)

	success := p.lastSuccess.WithLabelValues(
		s.ClientIP,
		fmt.Sprintf("%s:%s", s.ServerIP, s.ServerPort))
	success.Set(float64(time.Now().Unix()))

	return p.emitter.OnSummary(s)
}
