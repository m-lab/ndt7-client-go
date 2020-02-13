package emitter

import (
	"encoding/json"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal"
	"github.com/m-lab/ndt7-client-go/spec"
)

// JSON is a JSON emitter. It emits messages consistent with
// the cmd/ndt7-client/main.go documentation for `-format=json`.
type JSON struct {
	io.Writer
}

// NewJSON creates a new JSON emitter
func NewJSON() JSON {
	return JSON{
		Writer: os.Stdout,
	}
}

func (j JSON) emitData(data []byte) error {
	_, err := j.Write(append(data, byte('\n')))
	return err
}

func (j JSON) emitInterface(any interface{}) error {
	data, err := json.Marshal(any)
	if err != nil {
		return err
	}
	return j.emitData(data)
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
func (j JSON) OnStarting(test spec.TestKind) error {
	return j.emitInterface(batchEvent{
		Key: "starting",
		Value: batchValue{
			Measurement: spec.Measurement{
				Test: test,
			},
		},
	})
}

// OnError emits the error event
func (j JSON) OnError(test spec.TestKind, err error) error {
	return j.emitInterface(batchEvent{
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
func (j JSON) OnConnected(test spec.TestKind, fqdn string) error {
	return j.emitInterface(batchEvent{
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
func (j JSON) OnDownloadEvent(m *spec.Measurement) error {
	return j.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

// OnUploadEvent handles an event emitted during the upload
func (j JSON) OnUploadEvent(m *spec.Measurement) error {
	return j.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

// OnComplete is the event signalling the end of the test
func (j JSON) OnComplete(test spec.TestKind) error {
	return j.emitInterface(batchEvent{
		Key: "complete",
		Value: batchValue{
			Measurement: spec.Measurement{
				Test: test,
			},
		},
	})
}

// OnSummary handles the summary event, emitted after the test is over.
func (j JSON) OnSummary(s *internal.Summary) error {
	return j.emitInterface(s)
}
