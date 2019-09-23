package emitter

import (
	"encoding/json"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

// Batch is a batch emitter. It emits messages consistent with
// the cmd/ndt7-client/main.go documentation for `-batch`.
type Batch struct {
	io.Writer
}

// NewBatch creates a new batch emitter
func NewBatch() Batch {
	return Batch{
		Writer: os.Stdout,
	}
}

func (b Batch) emitData(data []byte) error {
	_, err := b.Write(append(data, byte('\n')))
	return err
}

func (b Batch) emitInterface(any interface{}) error {
	data, err := json.Marshal(any)
	if err != nil {
		return err
	}
	return b.emitData(data)
}

type batchEvent struct {
	Key   string
	Value interface{}
}

type batchValue struct {
	Failure string
	Server  string
	Test    string
}

// OnStarting emits the starting event
func (b Batch) OnStarting(test string) error {
	return b.emitInterface(batchEvent{
		Key: "starting",
		Value: batchValue{
			Test: test,
		},
	})
}

// OnError emits the error event
func (b Batch) OnError(test string, err error) error {
	return b.emitInterface(batchEvent{
		Key: "error",
		Value: batchValue{
			Failure: err.Error(),
			Test:    test,
		},
	})
}

// OnConnected emits the connected event
func (b Batch) OnConnected(test, fqdn string) error {
	return b.emitInterface(batchEvent{
		Key: "connected",
		Value: batchValue{
			Server: fqdn,
			Test:   test,
		},
	})
}

// OnDownloadEvent handles an event emitted during the download
func (b Batch) OnDownloadEvent(m *spec.Measurement) error {
	return b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

// OnUploadEvent handles an event emitted during the upload
func (b Batch) OnUploadEvent(m *spec.Measurement) error {
	return b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

// OnComplete is the event signalling the end of the test
func (b Batch) OnComplete(test string) error {
	return b.emitInterface(batchEvent{
		Key: "complete",
		Value: batchValue{
			Test: test,
		},
	})
}
