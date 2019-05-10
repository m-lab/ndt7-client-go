package emitter

import (
	"encoding/json"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

type batch struct{}

var osStdoutWrite = os.Stdout.Write

func (batch) emitData(data []byte) error {
	_, err := osStdoutWrite(append(data, byte('\n')))
	return err
}

var jsonMarshal = json.Marshal

func (b batch) emitInterface(any interface{}) error {
	data, err := jsonMarshal(any)
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

func (b batch) OnStarting(subtest string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_start",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}

func (b batch) OnError(subtest string, err error) error {
	return b.emitInterface(batchEvent{
		Key: "failure.measurement",
		Value: batchValue{
			Failure: err.Error(),
			Subtest: subtest,
		},
	})
}

func (b batch) OnConnected(subtest, fqdn string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_begin",
		Value: batchValue{
			Server:  fqdn,
			Subtest: subtest,
		},
	})
}

func (b batch) OnDownloadEvent(m *spec.Measurement) error {
	return b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

func (b batch) OnUploadEvent(m *spec.Measurement) error {
	return b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

func (b batch) OnComplete(subtest string) error {
	return b.emitInterface(batchEvent{
		Key: "status.measurement_done",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}
