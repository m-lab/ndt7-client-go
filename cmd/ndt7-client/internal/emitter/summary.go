package emitter

import (
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Summary is a summary emitter. It only emits a summary at the end of
// the tests, if they were successful, and errors otherwise. The exact output
// format is delegated to the embedded Emitter.
type Summary struct {
	emitter Emitter
}

func NewSummary(e Emitter) Emitter {
	return &Summary{
		emitter: e,
	}
}

// OnStarting emits the starting event
func (s Summary) OnStarting(test spec.TestKind) error {
	return nil
}

// OnError emits the error event
func (s Summary) OnError(test spec.TestKind, err error) error {
	return s.emitter.OnError(test, err)
}

// OnConnected emits the connected event
func (s Summary) OnConnected(test spec.TestKind, fqdn string) error {
	return nil
}

// OnDownloadEvent handles an event emitted during the download
func (s Summary) OnDownloadEvent(m *spec.Measurement) error {
	return nil
}

// OnUploadEvent handles an event emitted during the upload
func (s Summary) OnUploadEvent(m *spec.Measurement) error {
	return nil
}

// OnComplete is the event signalling the end of the test
func (s Summary) OnComplete(test spec.TestKind) error {
	return nil
}

// OnSummary handles the summary event, emitted after the test is over.
func (s Summary) OnSummary(results map[spec.TestKind]*ndt7.MeasurementPair) error {
	return s.emitter.OnSummary(results)
}
