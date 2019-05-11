package emitter

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestBatchOnStarting verifies that OnStarting works correctly
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
		Key   string `json:"key"`
		Value struct {
			Subtest string `json:"subtest"`
		} `json:"value"`
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "status.measurement_start" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Subtest != "download" {
		t.Fatal("Unexpected subtest field value")
	}
}

// TestBatchOnStartingFailure verifies that OnStarting
// fails if we cannot write.
func TestBatchOnStartingFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestBatchOnError verifies that OnError works correctly
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
		Key   string `json:"key"`
		Value struct {
			Failure string `json:"failure"`
			Subtest string `json:"subtest"`
		} `json:"value"`
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "failure.measurement" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Subtest != "download" {
		t.Fatal("Unexpected subtest field value")
	}
	if event.Value.Failure != "mocked error" {
		t.Fatal("Unexpected failure field value")
	}
}

// TestBatchOnErrorFailure verifies that OnError
// fails if we cannot write.
func TestBatchOnErrorFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestBatchOnConnected verifies that OnConnected works correctly
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
		Key   string `json:"key"`
		Value struct {
			Server  string `json:"server"`
			Subtest string `json:"subtest"`
		} `json:"value"`
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "status.measurement_begin" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Subtest != "download" {
		t.Fatal("Unexpected subtest field value")
	}
	if event.Value.Server != "FQDN" {
		t.Fatal("Unexpected failure field value")
	}
}

// TestBatchOnConnectedFailure verifies that OnConnected
// fails if we cannot write.
func TestBatchOnConnectedFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestBatchOnDownloadEvent verifies that OnDownloadEvent
// works correctly.
func TestBatchOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnDownloadEvent(&spec.Measurement{
		BBRInfo: spec.BBRInfo{
			MaxBandwidth: 6400000,
			MinRTT:       71,
		},
		Direction: "download",
		Elapsed:   4,
		Origin:    "server",
		TCPInfo: spec.TCPInfo{
			RTTVar:      11,
			SmoothedRTT: 150,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string `json:"key"`
		Value struct {
			BBRInfo struct {
				MaxBandwidth int64   `json:"max_bandwidth"`
				MinRTT       float64 `json:"min_rtt"`
			} `json:"bbr_info"`
			Direction string  `json:"direction"`
			Elapsed   float64 `json:"elapsed"`
			Origin    string  `json:"origin"`
			TCPInfo   struct {
				SmoothedRTT float64 `json:"smoothed_rtt"`
				RTTVar      float64 `json:"rtt_var"`
			} `json:"tcp_info"`
		} `json:"value"`
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "measurement" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.BBRInfo.MaxBandwidth != 6400000 {
		t.Fatal("Unexpected max bandwidth field value")
	}
	if event.Value.BBRInfo.MinRTT != 71 {
		t.Fatal("Unexpected min rtt field value")
	}
	if event.Value.Direction != "download" {
		t.Fatal("Unexpected direction field value")
	}
	if event.Value.Elapsed != 4.0 {
		t.Fatal("Unexpected elapsed field value")
	}
	if event.Value.Origin != "server" {
		t.Fatal("Unexpected origin field value")
	}
	if event.Value.TCPInfo.SmoothedRTT != 150.0 {
		t.Fatal("Unexpected smoothed rtt field value")
	}
	if event.Value.TCPInfo.RTTVar != 11.0 {
		t.Fatal("Unexpected rtt var value")
	}
}

// TestBatchOnDownloadEventFailure verifies that OnDownloadEvent
// fails if we cannot write.
func TestBatchOnDownloadEventFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnDownloadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestBatchOnUploadEvent verifies that OnUploadEvent
// works correctly.
func TestBatchOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	batch := Batch{sw}
	err := batch.OnUploadEvent(&spec.Measurement{
		AppInfo: spec.AppInfo{
			NumBytes: 100000000,
		},
		Direction: "upload",
		Elapsed:   3.0,
		Origin:    "client",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	var event struct {
		Key   string `json:"key"`
		Value struct {
			AppInfo struct {
				NumBytes int64 `json:"num_bytes"`
			} `json:"app_info"`
			Direction string  `json:"direction"`
			Elapsed   float64 `json:"elapsed"`
			Origin    string  `json:"origin"`
		} `json:"value"`
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
	if event.Value.Direction != "upload" {
		t.Fatal("Unexpected direction field value")
	}
	if event.Value.Elapsed != 3.0 {
		t.Fatal("Unexpected elapsed field value")
	}
	if event.Value.Origin != "client" {
		t.Fatal("Unexpected elapsed field value")
	}
}

// TestBatchOnUploadEventFailure verifies that OnUploadEvent
// fails if we cannot write.
func TestBatchOnUploadEventFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnUploadEvent(&spec.Measurement{
		Elapsed: 1.0,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestBatchOnComplete verifies that OnComplete works correctly
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
		Key   string `json:"key"`
		Value struct {
			Subtest string `json:"subtest"`
		} `json:"value"`
	}
	err = json.Unmarshal(sw.Data[0], &event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Key != "status.measurement_done" {
		t.Fatal("Unexpected event key")
	}
	if event.Value.Subtest != "download" {
		t.Fatal("Unexpected subtest field value")
	}
}

// TestBatchOnCompleteFailure verifies that OnComplete
// fails if we cannot write.
func TestBatchOnCompleteFailure(t *testing.T) {
	batch := Batch{&mocks.FailingWriter{}}
	err := batch.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestNewBatchConstructor verifies that we are
// constructing a batch bound to stdout.
func TestNewBatchConstructor(t *testing.T) {
	batch := NewBatch()
	if batch.Writer != os.Stdout {
		t.Fatal("Batch is not using stdout")
	}
}

// TestEmitInterfaceFailure makes sure that emitInterface
// correctly deals with a non serializable type.
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
