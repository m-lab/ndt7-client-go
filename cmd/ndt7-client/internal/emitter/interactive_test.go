package emitter

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

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
		t.Fatal("unexpected output")
	}
}

func TestInteractiveOnStartingFailure(t *testing.T) {
	interactive := Interactive{&mocks.FailingWriter{}}
	err := interactive.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

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
		t.Fatal("unexpected output")
	}
}

func TestInteractiveOnErrorFailure(t *testing.T) {
	interactive := Interactive{&mocks.FailingWriter{}}
	err := interactive.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

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
		t.Fatal("unexpected output")
	}
}

func TestInteractiveOnConnectedFailure(t *testing.T) {
	interactive := Interactive{&mocks.FailingWriter{}}
	err := interactive.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestInteractiveOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnDownloadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 3000000,
			NumBytes:    100000000,
		},
		Origin: spec.OriginClient,
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
		t.Fatal("unexpected output")
	}
}

func TestInteractiveOnDownloadEventFailure(t *testing.T) {
	interactive := Interactive{&mocks.FailingWriter{}}
	err := interactive.OnDownloadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234,
		},
		Origin: spec.OriginClient,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestInteractiveIgnoresServerData(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{
		Origin: spec.OriginServer,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("invalid length")
	}
}

func TestInteractiveOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 3000000,
			NumBytes:    100000000,
		},
		Origin: spec.OriginClient,
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
		t.Fatal("unexpected output")
	}
}

func TestInteractiveOnUploadEventSafetyCheck(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{
		Origin: spec.OriginClient,
	})
	if err == nil {
		t.Fatal("We did expect an error here")
	}
	if len(sw.Data) != 0 {
		t.Fatal("Some data was written and it shouldn't have")
	}
}

func TestInteractiveOnUploadEventDivideByZero(t *testing.T) {
	sw := &mocks.SavingWriter{}
	interactive := Interactive{sw}
	err := interactive.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{},
		Origin:  spec.OriginClient,
	})
	if err == nil {
		t.Fatal("We did expect an error here")
	}
	if len(sw.Data) != 0 {
		t.Fatal("Some data was written and it shouldn't have")
	}
}

func TestInteractiveOnUploadEventFailure(t *testing.T) {
	interactive := Interactive{&mocks.FailingWriter{}}
	err := interactive.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234,
		},
		Origin: spec.OriginClient,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

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
		t.Fatal("unexpected output")
	}
}

func TestInteractiveOnCompleteFailure(t *testing.T) {
	interactive := Interactive{&mocks.FailingWriter{}}
	err := interactive.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestNewInteractiveConstructor(t *testing.T) {
	interactive := NewInteractive()
	if interactive.out != os.Stdout {
		t.Fatal("Interactive is not using stdout")
	}
}
