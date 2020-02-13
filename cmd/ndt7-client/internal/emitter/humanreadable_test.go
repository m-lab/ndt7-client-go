package emitter

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

func TestHumanReadableOnStarting(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnStarting("download")
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

func TestHumanReadableOnStartingFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnError(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnError("download", errors.New("mocked error"))
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

func TestHumanReadableOnErrorFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnConnected(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnConnected("download", "FQDN")
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

func TestHumanReadableOnConnectedFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnDownloadEvent(&spec.Measurement{
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

func TestHumanReadableOnDownloadEventFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnDownloadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234,
		},
		Origin: spec.OriginClient,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableIgnoresServerData(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
		Origin: spec.OriginServer,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("invalid length")
	}
}

func TestHumanReadableOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
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

func TestHumanReadableOnUploadEventSafetyCheck(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
		Origin: spec.OriginClient,
	})
	if err == nil {
		t.Fatal("We did expect an error here")
	}
	if len(sw.Data) != 0 {
		t.Fatal("Some data was written and it shouldn't have")
	}
}

func TestHumanReadableOnUploadEventDivideByZero(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
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

func TestHumanReadableOnUploadEventFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234,
		},
		Origin: spec.OriginClient,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnComplete(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnComplete("download")
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

func TestHumanReadableOnCompleteFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestNewInteractiveConstructor(t *testing.T) {
	hr := NewHumanReadable()
	if hr.out != os.Stdout {
		t.Fatal("Interactive is not using stdout")
	}
}
