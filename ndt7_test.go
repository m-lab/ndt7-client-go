package ndt7

import (
	"context"
	"fmt"
	"testing"
)

// TestIntegration runs a ndt7 test.
func TestIntegration(t *testing.T) {
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
