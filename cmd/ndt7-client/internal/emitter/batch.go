package emitter

import (
	"encoding/json"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

type Batch struct{
	osStdoutWrite func([]byte)(int, error)
}

func NewBatch() Batch {
	return Batch{
		osStdoutWrite: os.Stdout.Write,
	}
}

func (b Batch) emitData(data []byte) error {
	_, err := b.osStdoutWrite(append(data, byte('\n')))
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

func (b Batch) OnStarting(subtest string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_start",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}

func (b Batch) OnError(subtest string, err error) error {
	return b.emitInterface(batchEvent{
		Key: "failure.measurement",
		Value: batchValue{
			Failure: err.Error(),
			Subtest: subtest,
		},
	})
}

func (b Batch) OnConnected(subtest, fqdn string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_begin",
		Value: batchValue{
			Server:  fqdn,
			Subtest: subtest,
		},
	})
}

func (b Batch) OnDownloadEvent(m *spec.Measurement) error {
	return b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

func (b Batch) OnUploadEvent(m *spec.Measurement) error {
	return b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

func (b Batch) OnComplete(subtest string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_done",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}
