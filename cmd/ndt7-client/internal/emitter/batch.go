package emitter

import (
	"encoding/json"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go"
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
	spec.Measurement
	Failure string `json:",omitempty"`
	Server  string `json:",omitempty"`
}

// OnStarting emits the starting event
func (b Batch) OnStarting(test spec.TestKind) error {
	return b.emitInterface(batchEvent{
		Key: "starting",
		Value: batchValue{
			Measurement: spec.Measurement{
				Test: test,
			},
		},
	})
}

// OnError emits the error event
func (b Batch) OnError(test spec.TestKind, err error) error {
	return b.emitInterface(batchEvent{
		Key: "error",
		Value: batchValue{
			Measurement: spec.Measurement{
				Test: test,
			},
			Failure: err.Error(),
		},
	})
}

// OnConnected emits the connected event
func (b Batch) OnConnected(test spec.TestKind, fqdn string) error {
	return b.emitInterface(batchEvent{
		Key: "connected",
		Value: batchValue{
			Measurement: spec.Measurement{
				Test: test,
			},
			Server: fqdn,
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
func (b Batch) OnComplete(test spec.TestKind) error {
	return b.emitInterface(batchEvent{
		Key: "complete",
		Value: batchValue{
			Measurement: spec.Measurement{
				Test: test,
			},
		},
	})
}

// OnSummary handles the summary event, emitted after the test is over.
func (b Batch) OnSummary(results map[spec.TestKind]*ndt7.MeasurementPair) error {
	return b.emitInterface(results)
}
