package emitter

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestInteractiveOnStarting verifies that OnStarting works correctly
func TestInteractiveOnStarting(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnStarting("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\rstarting download")) {
		t.Fatal("unexpected ouput")
	}
}

// TestInteractiveOnStartingFailure verifies that OnStarting
// fails if we cannot write.
func TestInteractiveOnStartingFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestInteractiveOnError verifies that OnError works correctly
func TestInteractiveOnError(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnError("download", errors.New("mocked error"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\rdownload failed: mocked error\n")) {
		t.Fatal("unexpected ouput")
	}
}

// TestInteractiveOnErrorFailure verifies that OnError
// fails if we cannot write.
func TestInteractiveOnErrorFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestInteractiveOnConnected verifies that OnConnected works correctly
func TestInteractiveOnConnected(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnConnected("download", "FQDN")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\rdownload in progress with FQDN\n")) {
		t.Fatal("unexpected ouput")
	}
}

// TestInteractiveOnConnectedFailure verifies that OnConnected
// fails if we cannot write.
func TestInteractiveOnConnectedFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestInteractiveOnDownloadEvent verifies that OnDownloadEvent
// works correctly.
func TestInteractiveOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnDownloadEvent(&spec.Measurement{
		BBRInfo: spec.BBRInfo{
			MaxBandwidth: 6400000,
			MinRTT:       71,
		},
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
	if !reflect.DeepEqual(
		sw.Data[0],
		[]byte("\rMaxBandwidth:     6.4 Mbit/s - RTT:   71/ 150/  11 (min/smoothed/var) ms"),
	) {
		t.Fatal("unexpected ouput")
	}
}

// TestInteractiveOnDownloadEventFailure verifies that OnDownloadEvent
// fails if we cannot write.
func TestInteractiveOnDownloadEventFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnDownloadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestInteractiveOnUploadEvent verifies that OnUploadEvent
// works correctly.
func TestInteractiveOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{
		AppInfo: spec.AppInfo{
			NumBytes: 100000000,
		},
		Elapsed: 3.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(
		sw.Data[0],
		[]byte("\rAvg. speed  :   266.7 Mbit/s"),
	) {
		t.Fatal("unexpected ouput")
	}
}

// TestInteractiveOnUploadEventDivideByZero verifies that
// OnUploadEvent punts if we try to divide by zero.
func TestInteractiveOnUploadEventDivideByZero(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{})
	if err == nil {
		t.Fatal("We did expect an error here")
	}
	if len(sw.Data) != 0 {
		t.Fatal("Some data was written and it shouldn't have")
	}
}

// TestInteractiveOnUploadEventFailure verifies that OnUploadEvent
// fails if we cannot write.
func TestInteractiveOnUploadEventFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{
		Elapsed: 1.0,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestInteractiveOnComplete verifies that OnComplete works correctly
func TestInteractiveOnComplete(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnComplete("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\ndownload: complete\n")) {
		t.Fatal("unexpected ouput")
	}
}

// TestInteractiveOnCompleteFailure verifies that OnComplete
// fails if we cannot write.
func TestInteractiveOnCompleteFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

// TestNewInteractiveConstructor verifies that we are
// constructing an interactive bound to stdout.
func TestNewInteractiveConstructor(t *testing.T) {
	interactive := NewInteractive()
	if interactive.out != os.Stdout {
		t.Fatal("Interactive is not using stdout")
	}
}
