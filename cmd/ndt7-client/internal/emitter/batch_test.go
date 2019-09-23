package emitter

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

func TestBatchOnStarting(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnStarting("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string
		Value struct {
			Test string
		}
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "starting" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Test != "download" {
		t.Fatal("Unexpected test field value")
	}
}

func TestBatchOnStartingFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestBatchOnError(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnError("download", errors.New("mocked error"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string
		Value struct {
			Failure string
			Test    string
		}
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "error" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Test != "download" {
		t.Fatal("Unexpected test field value")
	}
	if event.Value.Failure != "mocked error" {
		t.Fatal("Unexpected failure field value")
	}
}

func TestBatchOnErrorFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestBatchOnConnected(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnConnected("download", "FQDN")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string
		Value struct {
			Server string
			Test   string
		}
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "connected" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Test != "download" {
		t.Fatal("Unexpected test field value")
	}
	if event.Value.Server != "FQDN" {
		t.Fatal("Unexpected failure field value")
	}
}

func TestBatchOnConnectedFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestBatchOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnDownloadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 7100000,
			NumBytes:    41000,
		},
		Test:   spec.TestDownload,
		Origin: spec.OriginClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string
		Value struct {
			AppInfo struct {
				ElapsedTime int64
				NumBytes    int64
			}
			Origin string
			Test   string
		}
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "measurement" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.AppInfo.ElapsedTime != 7100000 {
		t.Fatal("Unexpected max bandwidth field value")
	}
	if event.Value.AppInfo.NumBytes != 41000 {
		t.Fatal("Unexpected min rtt field value")
	}
	if event.Value.Test != "download" {
		t.Fatal("Unexpected direction field value")
	}
	if event.Value.Origin != "client" {
		t.Fatal("Unexpected origin field value")
	}
}

func TestBatchOnDownloadEventFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnDownloadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestBatchOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 3000000,
			NumBytes:    100000000,
		},
		Test:   spec.TestUpload,
		Origin: spec.OriginClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string
		Value struct {
			AppInfo struct {
				ElapsedTime int64
				NumBytes    int64
			}
			Origin string
			Test   string
		}
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "measurement" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.AppInfo.NumBytes != 100000000 {
		t.Fatal("Unexpected num bytes field value")
	}
	if event.Value.Test != "upload" {
		t.Fatal("Unexpected direction field value")
	}
	if event.Value.AppInfo.ElapsedTime != 3000000 {
		t.Fatal("Unexpected elapsed field value")
	}
	if event.Value.Origin != "client" {
		t.Fatal("Unexpected elapsed field value")
	}
}

func TestBatchOnUploadEventFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnUploadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestBatchOnComplete(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnComplete("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string
		Value struct {
			Test string
		}
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "complete" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Test != "download" {
		t.Fatal("Unexpected test field value")
	}
}

func TestBatchOnCompleteFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestNewBatchConstructor(t *testing.T) {
	batch := NewBatch()
	if batch.Writer != os.Stdout {
		t.Fatal("Batch is not using stdout")
	}
}

func TestEmitInterfaceFailure(t *testing.T) {
	batch := NewBatch()
	// See https://stackoverflow.com/a/48901259
	x := map[string]interface{}{
		"foo": make(chan int),
	}
	err := batch.emitInterface(x)
	switch err.(type) {
	case *json.UnsupportedTypeError:
		// nothing
	default:
		t.Fatal("Expected a json.UnsupportedTypeError here")
	}
}
