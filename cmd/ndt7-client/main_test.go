package main

import (
	"context"
	"errors"
	"testing"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/mlabns"
)

// TestNormalUsage tests ndt7-client w/o any command line arguments.
func TestNormalUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	main()
}

// TestDownloadError tests the case where download fails.
func TestDownloadError(t *testing.T) {
	client := ndt7.NewClient(context.Background())
	mockedError := errors.New("mocked error")
	client.LocateFn = func(client *mlabns.Client) (string, error) {
		return "", mockedError
	}
	download(client, batch{})
}

// TestUploadError tests the case where upload fails.
func TestUploadError(t *testing.T) {
	client := ndt7.NewClient(context.Background())
	mockedError := errors.New("mocked error")
	client.LocateFn = func(client *mlabns.Client) (string, error) {
		return "", mockedError
	}
	upload(client, batch{})
}
