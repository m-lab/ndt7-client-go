package ndt7_test

import (
	"context"
	"log"
	"time"

	"github.com/m-lab/ndt7-client-go"
)

// This shows how to run a ndt7 test.
func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client := ndt7.NewClient("ndt7-client-go-example", "0.1.0")
	ch, err := client.StartDownload(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for ev := range ch {
		log.Printf("%+v", ev)
	}
	ch, err = client.StartUpload(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for ev := range ch {
		log.Printf("%+v", ev)
	}
}
