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
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type batchValue struct {
	Failure string `json:"failure,omitempty"`
	Server  string `json:"server,omitempty"`
	Subtest string `json:"subtest"`
}

// OnStarting emits the starting event
func (b Batch) OnStarting(subtest string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_start",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}

// OnError emits the error event
func (b Batch) OnError(subtest string, err error) error {
	return b.emitInterface(batchEvent{
		Key: "failure.measurement",
		Value: batchValue{
			Failure: err.Error(),
			Subtest: subtest,
		},
	})
}

// OnConnected emits the connected event
func (b Batch) OnConnected(subtest, fqdn string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_begin",
		Value: batchValue{
			Server:  fqdn,
			Subtest: subtest,
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

// OnComplete is the event signalling the end of the subtest
func (b Batch) OnComplete(subtest string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_done",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}
