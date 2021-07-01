package emitter

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
)

// HumanReadable is a human readable emitter. It emits the events generated
// by running a ndt7 test as pleasant stdout messages.
type HumanReadable struct {
	out io.Writer
}

// NewHumanReadable returns a new human readable emitter.
func NewHumanReadable() Emitter {
	return HumanReadable{os.Stdout}
}

// NewHumanReadableWithWriter returns a new human readable emitter using the
// specified writer.
func NewHumanReadableWithWriter(w io.Writer) Emitter {
	return HumanReadable{w}
}

// OnStarting handles the start event
func (h HumanReadable) OnStarting(test spec.TestKind) error {
	_, err := fmt.Fprintf(h.out, "\rstarting %s", test)
	return err
}

// OnError handles the error event
func (h HumanReadable) OnError(test spec.TestKind, err error) error {
	_, failure := fmt.Fprintf(h.out, "\r%s failed: %s\n", test, err.Error())
	return failure
}

// OnConnected handles the connected event
func (h HumanReadable) OnConnected(test spec.TestKind, fqdn string) error {
	_, err := fmt.Fprintf(h.out, "\r%s in progress with %s\n", test, fqdn)
	return err
}

// OnDownloadEvent handles an event emitted by the download test
func (h HumanReadable) OnDownloadEvent(m *spec.Measurement) error {
	return h.onSpeedEvent(m)
}

// OnUploadEvent handles an event emitted during the upload test
func (h HumanReadable) OnUploadEvent(m *spec.Measurement) error {
	return h.onSpeedEvent(m)
}

func (h HumanReadable) onSpeedEvent(m *spec.Measurement) error {
	// The specification recommends that we show application level
	// measurements. Let's just do that in interactive mode. To this
	// end, we ignore any measurement coming from the server.
	if m.Origin != spec.OriginClient {
		return nil
	}
	if m.AppInfo == nil || m.AppInfo.ElapsedTime <= 0 {
		return errors.New("Missing m.AppInfo or invalid m.AppInfo.ElapsedTime")
	}
	elapsed := float64(m.AppInfo.ElapsedTime) / 1e06
	v := (8.0 * float64(m.AppInfo.NumBytes)) / elapsed / (1000.0 * 1000.0)
	_, err := fmt.Fprintf(h.out, "\rAvg. speed  : %7.1f Mbit/s", v)
	return err
}

// OnComplete handles the complete event
func (h HumanReadable) OnComplete(test spec.TestKind) error {
	_, err := fmt.Fprintf(h.out, "\n%s: complete\n", test)
	return err
}

// OnSummary handles the summary event.
func (h HumanReadable) OnSummary(s *Summary) error {
	const summaryFormat = `%15s: %s
%15s: %s
%15s: %7.1f %s
%15s: %7.1f %s
%15s: %7.1f %s
%15s: %7.2f %s
`
	_, err := fmt.Fprintf(h.out, summaryFormat,
		"Server", s.ServerFQDN,
		"Client", s.ClientIP,
		"Latency", s.MinRTT.Value, s.MinRTT.Unit,
		"Download", s.Download.Value, s.Upload.Unit,
		"Upload", s.Upload.Value, s.Upload.Unit,
		"Retransmission", s.DownloadRetrans.Value, s.DownloadRetrans.Unit)
	if err != nil {
		return err
	}

	return nil
}
