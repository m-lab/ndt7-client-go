package main

import (
	"context"

	"github.com/apex/log"
	"github.com/m-lab/ndt7-client-go"
)

func main() {
	ctx := context.Background()
	fqdn, err := ndt7.DiscoverServer(ctx)
	if err != nil {
		log.WithError(err).Fatal("ndt7.DiscoverServer failed")
	}
	log.Infof("discovered server: %s", fqdn)
	client := ndt7.NewClient(ctx)
	client.FQDN = fqdn
	log.Info("starting download")
	ch, err := client.StartDownload()
	if err != nil {
		log.WithError(err).Fatal("client.StartDownload failed")
	}
	log.Info("download in progress")
	for ev := range ch {
		log.Infof("%+v", ev)
	}
	log.Info("download complete; starting upload")
	ch, err = client.StartUpload()
	if err != nil {
		log.WithError(err).Fatal("client.StartUpload failed")
	}
	log.Info("upload in progress")
	for ev := range ch {
		log.Infof("%+v", ev)
	}
	log.Info("upload complete")
}
