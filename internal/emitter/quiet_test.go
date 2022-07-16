package emitter

import (
	"os"
	"testing"

	"github.com/m-lab/ndt7-client-go/internal/mocks"

	"github.com/m-lab/ndt7-client-go/spec"
)

func TestNewQuiet(t *testing.T) {
	e := jsonEmitter{os.Stdout}
	if NewQuiet(e) == nil {
		t.Fatal("NewQuiet() did not return an Emitter")
	}
}

func TestQuiet_OnStarting(t *testing.T) {
	sw := &mocks.SavingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnStarting("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("OnStarting(): unexpected data")
	}
}

func TestQuiet_OnError(t *testing.T) {
	// The only thing to test here is that errors from the underlying emitter
	// are passed back to the caller.
	sw := &mocks.FailingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnError("download", mocks.ErrMocked)
	if err != mocks.ErrMocked {
		t.Fatal("OnError(): unexpected error type or nil")
	}
}

func TestQuiet_OnConnected(t *testing.T) {
	sw := &mocks.SavingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnConnected("download", "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("OnConnected(): unexpected data")
	}
}

func TestQuiet_OnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnDownloadEvent(&spec.Measurement{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("OnDownloadEvent(): unexpected data")
	}
}

func TestQuiet_OnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnUploadEvent(&spec.Measurement{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("OnUploadEvent(): unexpected data")
	}
}

func TestQuiet_OnComplete(t *testing.T) {
	sw := &mocks.SavingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnComplete("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("OnComplete(): unexpected data")
	}
}

func TestQuiet_OnSummary(t *testing.T) {
	// The only thing to test here is that errors from the underlying emitter
	// are passed back to the caller.
	sw := &mocks.FailingWriter{}
	e := jsonEmitter{sw}
	quiet := Quiet{e}
	err := quiet.OnSummary(&Summary{})
	if err != mocks.ErrMocked {
		t.Fatal("OnSummary(): unexpected error type or nil")
	}
}
