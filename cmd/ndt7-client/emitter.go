package main

import "github.com/m-lab/ndt7-client-go/spec"

type emitter interface {
	onStarting(subtest string)
	onError(subtest string, err error)
	onConnected(subtest, fqdn string)
	onDownloadEvent(m *spec.Measurement)
	onUploadEvent(m *spec.Measurement)
	onComplete(subtest string)
}
