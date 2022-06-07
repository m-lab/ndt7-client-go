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
	download, upload, rtt, completionTime *prometheus.GaugeVec
}

// NewPrometheus returns a Summary emitter which emits messages
// via the passed Emitter.
func NewPrometheus(e Emitter, download, upload, rtt, completionTime *prometheus.GaugeVec) Emitter {
	return &Prometheus{e, download, upload, rtt, completionTime}
}

// OnStarting emits the starting event
func (p Prometheus) OnStarting(test spec.TestKind) error {
	return p.emitter.OnStarting(test)
}

// OnError emits the error event
func (p Prometheus) OnError(test spec.TestKind, err error) error {
	g := p.completionTime.WithLabelValues(string(test), "ERROR")
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
	g := p.completionTime.WithLabelValues(string(test), "OK")
	g.Set(float64(time.Now().Unix()))
	return p.emitter.OnComplete(test)
}

// OnSummary handles the summary event, emitted after the test is over.
func (p *Prometheus) OnSummary(s *Summary) error {
	p.download.WithLabelValues(s.ClientIP).Set(s.Download.Value)
	p.upload.WithLabelValues(s.ClientIP).Set(s.Upload.Value)
	p.rtt.WithLabelValues(s.ClientIP).Set(s.MinRTT.Value)
	return p.emitter.OnSummary(s)
}
