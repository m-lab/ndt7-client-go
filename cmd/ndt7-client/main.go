package main

import (
	"context"

	"github.com/apex/log"
	"github.com/m-lab/ndt7-client-go"
)

func main() {
	client := ndt7.NewClient(context.Background())
	log.Info("starting download")
	ch, err := client.StartDownload()
	if err != nil {
		log.WithError(err).Fatal("client.StartDownload failed")
	}
	log.Infof("download in progress with %s", client.FQDN)
	for ev := range ch {
		log.Infof("%+v", ev)
	}
	log.Info("download complete; starting upload")
	ch, err = client.StartUpload()
	if err != nil {
		log.WithError(err).Fatal("client.StartUpload failed")
	}
	log.Infof("upload in progress with %s", client.FQDN)
	for ev := range ch {
		log.Infof("%+v", ev)
	}
	log.Info("upload complete")
}
