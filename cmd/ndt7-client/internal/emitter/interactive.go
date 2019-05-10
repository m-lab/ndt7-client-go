package emitter

import (
	"errors"
	"fmt"

	"github.com/m-lab/ndt7-client-go/spec"
)

type Interactive struct{}

func (Interactive) OnStarting(subtest string) error {
	_, err := fmt.Printf("\rstarting %s", subtest)
	return err
}

func (Interactive) OnError(subtest string, err error) error {
	_, failure := fmt.Printf("\r%s failed: %s\n", subtest, err.Error())
	return failure
}

func (Interactive) OnConnected(subtest, fqdn string) error {
	_, err := fmt.Printf("\r%s in progress with %s\n", subtest, fqdn)
	return err
}

func (Interactive) OnDownloadEvent(m *spec.Measurement) error {
	_, err := fmt.Printf(
		"\rMaxBandwidth: %7.1f Mbit/s - RTT: %4.0f/%4.0f/%4.0f (min/smoothed/var) ms",
		float64(m.BBRInfo.MaxBandwidth)/(1000.0*1000.0),
		m.BBRInfo.MinRTT,
		m.TCPInfo.SmoothedRTT,
		m.TCPInfo.RTTVar,
	)
	return err
}

func (Interactive) OnUploadEvent(m *spec.Measurement) error {
	if m.Elapsed <= 0.0 {
		return errors.New("Negative or zero m.Elapsed")
	}
	v := (8.0 * float64(m.AppInfo.NumBytes)) / m.Elapsed / (1000.0 * 1000.0)
	_, err := fmt.Printf("\rAvg. speed  : %7.1f Mbit/s", v)
	return err
}

func (Interactive) OnComplete(subtest string) error {
	_, err := fmt.Printf("\n%s: complete\n", subtest)
	return err
}
