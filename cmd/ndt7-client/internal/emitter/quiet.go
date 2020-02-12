package emitter

import (
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Quiet acts as a filter allowing summary and error messages only, and
// doesn't perform any formatting.
// The message is actually emitted by the embedded Emitter.
type Quiet struct {
	emitter Emitter
}

// NewQuiet returns a Summary emitter which emits messages
// via the passed Emitter.
func NewQuiet(e Emitter) Emitter {
	return &Quiet{
		emitter: e,
	}
}

// OnStarting emits the starting event
func (s Quiet) OnStarting(test spec.TestKind) error {
	return nil
}

// OnError emits the error event
func (s Quiet) OnError(test spec.TestKind, err error) error {
	return s.emitter.OnError(test, err)
}

// OnConnected emits the connected event
func (s Quiet) OnConnected(test spec.TestKind, fqdn string) error {
	return nil
}

// OnDownloadEvent handles an event emitted during the download
func (s Quiet) OnDownloadEvent(m *spec.Measurement) error {
	return nil
}

// OnUploadEvent handles an event emitted during the upload
func (s Quiet) OnUploadEvent(m *spec.Measurement) error {
	return nil
}

// OnComplete is the event signalling the end of the test
func (s Quiet) OnComplete(test spec.TestKind) error {
	return nil
}

// OnSummary handles the summary event, emitted after the test is over.
func (s Quiet) OnSummary(results map[spec.TestKind]*ndt7.MeasurementPair) error {
	return s.emitter.OnSummary(results)
}
