package main

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/mlabns"
)

// TestNormalUsage tests ndt7-client w/o any command line arguments.
func TestNormalUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	main()
	osExit = savedFunc
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

// TestBatchUsage tests the -batch use case.
func TestBatchUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	exitval := realmain(*flagTimeout, *flagHostname, true)
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

// TestDownloadError tests the case where download fails.
func TestDownloadError(t *testing.T) {
	ctx := context.Background()
	client := ndt7.NewClient(userAgent)
	mockedError := errors.New("mocked error")
	client.LocateFn = func(ctx context.Context, client *mlabns.Client) (string, error) {
		return "", mockedError
	}
	exitval := download(ctx, client, batch{})
	if exitval == 0 {
		log.Fatal("expected to see a nonzero code here")
	}
}

// TestUploadError tests the case where upload fails.
func TestUploadError(t *testing.T) {
	ctx := context.Background()
	client := ndt7.NewClient(userAgent)
	mockedError := errors.New("mocked error")
	client.LocateFn = func(ctx context.Context, client *mlabns.Client) (string, error) {
		return "", mockedError
	}
	exitval := upload(ctx, client, interactive{})
	if exitval == 0 {
		log.Fatal("expected to see a nonzero code here")
	}
}

// TestBatchJSONMarshalPanic ensures that the code panics
// if we cannot marshal a JSON.
func TestBatchJSONMarshalPanic(t *testing.T) {
	savedFunc := jsonMarshal
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, errors.New("mocked error")
	}
	exitval := realmain(*flagTimeout, *flagHostname, true)
	if exitval == 0 {
		log.Fatal("expected to see a nonzero code here")
	}
	jsonMarshal = savedFunc
}

// TestBatchOSStdoutWritePanic ensures that the code panics
// if we cannot write on the standard output.
func TestBatchOSStdoutWritePanic(t *testing.T) {
	savedFunc := osStdoutWrite
	osStdoutWrite = func(b []byte) (n int, err error) {
		return 0, errors.New("mocked error")
	}
	exitval := realmain(*flagTimeout, *flagHostname, true)
	if exitval == 0 {
		log.Fatal("expected to see a nonzero code here")
	}
	osStdoutWrite = savedFunc
}
