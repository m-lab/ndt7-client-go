package emitter

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

// Interactive is an interactive emitter
type Interactive struct {
	out io.Writer
}

// NewInteractive returns a new interactive emitter.
func NewInteractive() Interactive {
	return Interactive{os.Stdout}
}

// OnStarting handles the start event
func (i Interactive) OnStarting(subtest string) error {
	_, err := fmt.Fprintf(i.out, "\rstarting %s", subtest)
	return err
}

// OnError handles the error event
func (i Interactive) OnError(subtest string, err error) error {
	_, failure := fmt.Fprintf(i.out, "\r%s failed: %s\n", subtest, err.Error())
	return failure
}

// OnConnected handles the connected event
func (i Interactive) OnConnected(subtest, fqdn string) error {
	_, err := fmt.Fprintf(i.out, "\r%s in progress with %s\n", subtest, fqdn)
	return err
}

// OnDownloadEvent handles an event emitted by the download subtest
func (i Interactive) OnDownloadEvent(m *spec.Measurement) error {
	_, err := fmt.Fprintf(i.out,
		"\rMaxBandwidth: %7.1f Mbit/s - RTT: %4.0f/%4.0f/%4.0f (min/smoothed/var) ms",
		float64(m.BBRInfo.MaxBandwidth)/(1000.0*1000.0),
		m.BBRInfo.MinRTT,
		m.TCPInfo.SmoothedRTT,
		m.TCPInfo.RTTVar,
	)
	return err
}

// OnUploadEvent handles an event emittd during the upload subtest
func (i Interactive) OnUploadEvent(m *spec.Measurement) error {
	if m.Elapsed <= 0.0 {
		return errors.New("Negative or zero m.Elapsed")
	}
	v := (8.0 * float64(m.AppInfo.NumBytes)) / m.Elapsed / (1000.0 * 1000.0)
	_, err := fmt.Fprintf(i.out, "\rAvg. speed  : %7.1f Mbit/s", v)
	return err
}

// OnComplete handles the complete event
func (i Interactive) OnComplete(subtest string) error {
	_, err := fmt.Fprintf(i.out, "\n%s: complete\n", subtest)
	return err
}
