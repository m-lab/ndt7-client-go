package main

import (
	"fmt"

	"github.com/m-lab/ndt7-client-go/spec"
)

type interactive struct{}

func (interactive) onStarting(subtest string) {
	fmt.Printf("\rstarting %s", subtest)
}

func (interactive) onError(subtest string, err error) {
	fmt.Printf("\r%s failed: %s\n", subtest, err.Error())
}

func (interactive) onConnected(subtest, fqdn string) {
	fmt.Printf("\r%s in progress with %s\n", subtest, fqdn)
}

func (interactive) onDownloadEvent(m *spec.Measurement) {
	fmt.Printf(
		"\rMaxBandwidth: %7.1f Mbit/s - RTT: %4.0f/%4.0f/%4.0f (min/smoothed/var) ms",
		float64(m.BBRInfo.MaxBandwidth)/(1000.0*1000.0),
		m.BBRInfo.MinRTT,
		m.TCPInfo.SmoothedRTT,
		m.TCPInfo.RTTVar,
	)
}

func (interactive) onUploadEvent(m *spec.Measurement) {
	if m.Elapsed > 0.0 {
		v := (8.0 * float64(m.AppInfo.NumBytes)) / m.Elapsed / (1000.0 * 1000.0)
		fmt.Printf("\rAvg. speed  : %7.1f Mbit/s", v)
	}
}

func (interactive) onComplete(subtest string) {
	fmt.Printf("\n%s: complete\n", subtest)
}
