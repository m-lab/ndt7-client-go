package ndt7

import (
	"context"
	"fmt"
	"testing"
)

// TestIntegration runs a ndt7 test.
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	client := NewClient(context.Background())
	ch, err := client.StartDownload()
	if err != nil {
		t.Fatal(err)
	}
	for ev := range ch {
		fmt.Printf("%+v\n", ev)
	}
	ch, err = client.StartUpload()
	if err != nil {
		t.Fatal(err)
	}
	for ev := range ch {
		fmt.Printf("%+v\n", ev)
	}
}

// TestStartDiscoverServerError ensures that we deal
// with an error when discovering a server.
func TestStartDiscoverServerError(t *testing.T) {
	client := NewClient(context.Background())
	client.MlabNSBaseURL = "\t" // cause URL parse to fail
	_, err := client.start(nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}

// TestStartConnectError ensures that we deal
// with an error when connecting.
func TestStartConnectError(t *testing.T) {
	client := NewClient(context.Background())
	client.FQDN = "\t" // cause URL parse to fail
	_, err := client.start(nil, "")
	if err == nil {
		t.Fatal("We expected an error here")
	}
}
