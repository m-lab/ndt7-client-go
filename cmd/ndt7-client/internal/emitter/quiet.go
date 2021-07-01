package emitter

import (
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
func (q Quiet) OnStarting(test spec.TestKind) error {
	return nil
}

// OnError emits the error event
func (q Quiet) OnError(test spec.TestKind, err error) error {
	return q.emitter.OnError(test, err)
}

// OnConnected emits the connected event
func (q Quiet) OnConnected(test spec.TestKind, fqdn string) error {
	return nil
}

// OnDownloadEvent handles an event emitted during the download
func (q Quiet) OnDownloadEvent(m *spec.Measurement) error {
	return nil
}

// OnUploadEvent handles an event emitted during the upload
func (q Quiet) OnUploadEvent(m *spec.Measurement) error {
	return nil
}

// OnComplete is the event signalling the end of the test
func (q Quiet) OnComplete(test spec.TestKind) error {
	return nil
}

// OnSummary handles the summary event, emitted after the test is over.
func (q Quiet) OnSummary(s *Summary) error {
	return q.emitter.OnSummary(s)
}
