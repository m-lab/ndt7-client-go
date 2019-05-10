package emitter

import (
	"errors"
	"testing"

	"github.com/m-lab/ndt7-client-go/spec"
)

// TestBatchDownloadError is what happens if you
// run a download and it results in an error.
func TestBatchDownloadError(t *testing.T) {
	batch := NewBatch()
	batch.OnStarting("download")
	batch.OnError("download", errors.New("mocked error"))
	batch.OnComplete("download")
}

// TestBatchDownloadNormal is what happens if you
// run a download and it there are no errors.
func TestBatchDownloadNormal(t *testing.T) {
	batch := NewBatch()
	batch.OnStarting("download")
	batch.OnConnected("download", "FQDN")
	batch.OnDownloadEvent(&spec.Measurement{})
	batch.OnComplete("download")
}

// TestBatchOnUploadEventError is what happens if you
// run a upload and OnUploadEvent fails.
func TestBatchOnUploadEventError(t *testing.T) {
	batch := NewBatch()
	batch.OnStarting("upload")
	batch.OnConnected("upload", "FQDN")
	batch.OnUploadEvent(&spec.Measurement{})
	batch.OnComplete("upload")
}

// TestBatchUploadNormal is what happens
// during a normal upload subtest.
func TestBatchUploadNormal(t *testing.T) {
	batch := NewBatch()
	batch.OnStarting("upload")
	batch.OnConnected("upload", "FQDN")
	batch.OnUploadEvent(&spec.Measurement{
		Elapsed: 1.0,
	})
	batch.OnComplete("upload")
}
