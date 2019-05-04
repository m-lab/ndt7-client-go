package ndt7_test

import (
	"context"
	"log"

	"github.com/m-lab/ndt7-client-go"
)

// This shows how to run a ndt7 test.
func Example() {
	client := ndt7.NewClient(context.Background())
	ch, err := client.StartDownload()
	if err != nil {
		log.Fatal(err)
	}
	for ev := range ch {
		log.Printf("%+v", ev)
	}
	ch, err = client.StartUpload()
	if err != nil {
		log.Fatal(err)
	}
	for ev := range ch {
		log.Printf("%+v", ev)
	}
}
