package emitter

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/m-lab/ndt7-client-go/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

func TestJSONOnStarting(t *testing.T) {
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnStarting("download")
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

func TestJSONOnStartingFailure(t *testing.T) {
	j := NewJSON(&mocks.FailingWriter{})
	err := j.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestJSONOnError(t *testing.T) {
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnError("download", errors.New("mocked error"))
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

func TestJSONOnErrorFailure(t *testing.T) {
	j := NewJSON(&mocks.FailingWriter{})
	err := j.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestJSONOnConnected(t *testing.T) {
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnConnected("download", "FQDN")
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

func TestJSONOnConnectedFailure(t *testing.T) {
	j := NewJSON(&mocks.FailingWriter{})
	err := j.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestJSONOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnDownloadEvent(&spec.Measurement{
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

func TestJSONOnDownloadEventFailure(t *testing.T) {
	j := NewJSON(&mocks.FailingWriter{})
	err := j.OnDownloadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestJSONOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnUploadEvent(&spec.Measurement{
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

func TestJSONOnUploadEventFailure(t *testing.T) {
	j := NewJSON(&mocks.FailingWriter{})
	err := j.OnUploadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestJSONOnComplete(t *testing.T) {
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnComplete("download")
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

func TestJSONOnCompleteFailure(t *testing.T) {
	j := NewJSON(&mocks.FailingWriter{})
	err := j.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestNewJSON(t *testing.T) {
	if NewJSON(&mocks.SavingWriter{}) == nil {
		t.Fatal("NewJSON did not return an Emitter")
	}
}

func TestEmitInterfaceFailure(t *testing.T) {
	j := jsonEmitter{Writer: os.Stdout}
	// See https://stackoverflow.com/a/48901259
	x := map[string]interface{}{
		"foo": make(chan int),
	}
	err := j.emitInterface(x)
	switch err.(type) {
	case *json.UnsupportedTypeError:
		// nothing
	default:
		t.Fatal("Expected a json.UnsupportedTypeError here")
	}
}

func TestJSONOnSummary(t *testing.T) {
	summary := &Summary{}
	sw := &mocks.SavingWriter{}
	j := NewJSON(sw)
	err := j.OnSummary(summary)
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}

	var output Summary
	err = json.Unmarshal(sw.Data[0], &output)
	if err != nil {
		t.Fatal(err)
	}
	if output.ClientIP != summary.ClientIP ||
		output.ServerFQDN != summary.ServerFQDN ||
		output.ServerIP != summary.ServerIP ||
		output.Download != summary.Download ||
		output.Upload != summary.Upload {
		t.Fatal("OnSummary(): unexpected output")
	}

}
