package emitter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/m-lab/ndt7-client-go/spec"
)

// TestBatchDownloadError is what happens if you
// run a download and it results in an error.
func TestBatchDownloadError(t *testing.T) {
	batch := NewBatch()
	err := batch.OnStarting("download")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnError("download", errors.New("mocked error"))
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnComplete("download")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
}

// TestBatchDownloadNormal is what happens if you
// run a download and it there are no errors.
func TestBatchDownloadNormal(t *testing.T) {
	batch := NewBatch()
	err := batch.OnStarting("download")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnConnected("download", "FQDN")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnDownloadEvent(&spec.Measurement{})
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnComplete("download")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
}

// TestBatchUploadNormal is what happens
// during a normal upload subtest.
func TestBatchUploadNormal(t *testing.T) {
	batch := NewBatch()
	err := batch.OnStarting("upload")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnConnected("upload", "FQDN")
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnUploadEvent(&spec.Measurement{Elapsed: 1.0})
	if err != nil {
		t.Fatal("Did not expect an error here")
	}
	err = batch.OnComplete("upload")
	if err != nil {
		t.Fatal("Did not expect an error here")
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

// errMocked is a mocked error
var errMocked = errors.New("mocked error")

// mockedWriter is a mocked writer
type mockedWriter struct{}

// Write always returns a mocked error
func (mockedWriter) Write([]byte) (int, error) {
	return 0, errMocked
}

// TestEmitDataFailure ensures that any function in the batch
// API properly deals with a writing error.
func TestEmitDataFailure(t *testing.T) {
	batch := Batch{mockedWriter{}}
	err := batch.emitData([]byte("abc"))
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.emitInterface(1234)
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.OnStarting("download")
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.OnError("download", errors.New("an error"))
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.OnConnected("download", "server-fqdn")
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.OnDownloadEvent(&spec.Measurement{})
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.OnUploadEvent(&spec.Measurement{})
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
	err = batch.OnComplete("download")
	if err != errMocked {
		t.Fatal("Not the result we expected")
	}
}
