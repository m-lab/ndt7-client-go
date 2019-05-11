package emitter

import (
	"errors"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestInteractiveDownloadError is what happens if you
// run a download and it results in an error.
func TestInteractiveDownloadError(t *testing.T) {
	interactive := NewInteractive()
	err := interactive.OnStarting("download")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnError("download", errors.New("mocked error"))
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnComplete("download")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
}

// TestInteractiveDownloadNormal is what happens if you
// run a download and it there are no errors.
func TestInteractiveDownloadNormal(t *testing.T) {
	interactive := NewInteractive()
	err := interactive.OnStarting("download")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnConnected("download", "FQDN")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnDownloadEvent(&spec.Measurement{})
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnComplete("download")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
}

// TestInteractiveOnUploadEventError is what happens if you
// run a upload and OnUploadEvent fails.
func TestInteractiveOnUploadEventError(t *testing.T) {
	interactive := NewInteractive()
	err := interactive.OnStarting("upload")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnConnected("upload", "FQDN")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}

	// Note: the following should fail because we're passing
	// a zero elapsed, hence we should avoid dividing by
	// zero and we should return an error to the caller.
	err = interactive.OnUploadEvent(&spec.Measurement{})
	if err == nil {
		t.Fatal("expected an error here")
	}

	err = interactive.OnComplete("upload")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
}

// TestInteractiveUploadNormal is what happens
// during a normal upload subtest.
func TestInteractiveUploadNormal(t *testing.T) {
	interactive := NewInteractive()
	err := interactive.OnStarting("upload")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnConnected("upload", "FQDN")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnUploadEvent(&spec.Measurement{
		Elapsed: 1.0,
	})
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
	err = interactive.OnComplete("upload")
	if err != nil {
		t.Fatal("We were not expecting an error here")
	}
}

// TestInteractiveFprintfFailure ensures that the interactive
// emitter APIs all deal with a Fprintf failure.
func TestInteractiveFprintfFailure(t *testing.T) {
	interactive := Interactive{mocks.FailingWriter{}}
	err := interactive.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the result we expected")
	}
	err = interactive.OnError("download", errors.New("an error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the result we expected")
	}
	err = interactive.OnConnected("download", "server-fqdn")
	if err != mocks.ErrMocked {
		t.Fatal("Not the result we expected")
	}
	err = interactive.OnDownloadEvent(&spec.Measurement{})
	if err != mocks.ErrMocked {
		t.Fatal("Not the result we expected")
	}
	err = interactive.OnUploadEvent(&spec.Measurement{
		Elapsed: 1.0,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the result we expected")
	}
	err = interactive.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the result we expected")
	}
}
