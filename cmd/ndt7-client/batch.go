package main

import (
	"encoding/json"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

type batch struct{}

func (batch) emitData(data []byte) {
	_, err := os.Stdout.Write(append(data, byte('\n')))
	if err != nil {
		panic(err)
	}
}

func (b batch) emitInterface(any interface{}) {
	data, err := json.Marshal(any)
	if err != nil {
		panic(err)
	}
	b.emitData(data)
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

func (b batch) onStarting(subtest string) {
	b.emitInterface(batchEvent{
		Key: "status.measurement_start",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}

func (b batch) onError(subtest string, err error) {
	b.emitInterface(batchEvent{
		Key: "failure.measurement",
		Value: batchValue{
			Failure: err.Error(),
		},
	})
}

func (b batch) onConnected(subtest, fqdn string) {
	b.emitInterface(batchEvent{
		Key: "status.measurement_begin",
		Value: batchValue{
			Server:  fqdn,
			Subtest: subtest,
		},
	})
}

func (b batch) onDownloadEvent(m *spec.Measurement) {
	b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

func (b batch) onUploadEvent(m *spec.Measurement) {
	b.emitInterface(batchEvent{
		Key:   "measurement",
		Value: m,
	})
}

func (b batch) onComplete(subtest string) {
	b.emitInterface(batchEvent{
		Key: "status.measurement_done",
		Value: batchValue{
			Subtest: subtest,
		},
	})
}
