package emitter

import (
	"errors"
	"testing"

	"github.com/m-lab/ndt7-client-go/spec"
)

// TestInteractiveDownloadError is what happens if you
// run a download and it results in an error.
func TestInteractiveDownloadError(t *testing.T) {
	interactive := Interactive{}
	interactive.OnStarting("download")
	interactive.OnError("download", errors.New("mocked error"))
	interactive.OnComplete("download")
}

// TestInteractiveDownloadNormal is what happens if you
// run a download and it there are no errors.
func TestInteractiveDownloadNormal(t *testing.T) {
	interactive := Interactive{}
	interactive.OnStarting("download")
	interactive.OnConnected("download", "FQDN")
	interactive.OnDownloadEvent(&spec.Measurement{})
	interactive.OnComplete("download")
}

// TestInteractiveOnUploadEventError is what happens if you
// run a upload and OnUploadEvent fails.
func TestInteractiveOnUploadEventError(t *testing.T) {
	interactive := Interactive{}
	interactive.OnStarting("upload")
	interactive.OnConnected("upload", "FQDN")
	err := interactive.OnUploadEvent(&spec.Measurement{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	interactive.OnComplete("upload")
}

// TestInteractiveUploadNormal is what happens
// during a normal upload subtest.
func TestInteractiveUploadNormal(t *testing.T) {
	interactive := Interactive{}
	interactive.OnStarting("upload")
	interactive.OnConnected("upload", "FQDN")
	interactive.OnUploadEvent(&spec.Measurement{
		Elapsed: 1.0,
	})
	interactive.OnComplete("upload")
}
