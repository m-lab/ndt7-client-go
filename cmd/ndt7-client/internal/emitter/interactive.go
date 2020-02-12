package emitter

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal"

	"github.com/m-lab/ndt7-client-go/spec"
)

// Interactive is an interactive emitter. It emits the events generated
// by running a ndt7 test as pleasant stdout messages.
type Interactive struct {
	out io.Writer
}

// NewInteractive returns a new interactive emitter.
func NewInteractive() Interactive {
	return Interactive{os.Stdout}
}

// OnStarting handles the start event
func (i Interactive) OnStarting(test spec.TestKind) error {
	_, err := fmt.Fprintf(i.out, "\rstarting %s", test)
	return err
}

// OnError handles the error event
func (i Interactive) OnError(test spec.TestKind, err error) error {
	_, failure := fmt.Fprintf(i.out, "\r%s failed: %s\n", test, err.Error())
	return failure
}

// OnConnected handles the connected event
func (i Interactive) OnConnected(test spec.TestKind, fqdn string) error {
	_, err := fmt.Fprintf(i.out, "\r%s in progress with %s\n", test, fqdn)
	return err
}

// OnDownloadEvent handles an event emitted by the download test
func (i Interactive) OnDownloadEvent(m *spec.Measurement) error {
	return i.onSpeedEvent(m)
}

// OnUploadEvent handles an event emitted during the upload test
func (i Interactive) OnUploadEvent(m *spec.Measurement) error {
	return i.onSpeedEvent(m)
}

func (i Interactive) onSpeedEvent(m *spec.Measurement) error {
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
	_, err := fmt.Fprintf(i.out, "\rAvg. speed  : %7.1f Mbit/s", v)
	return err
}

// OnComplete handles the complete event
func (i Interactive) OnComplete(test spec.TestKind) error {
	_, err := fmt.Fprintf(i.out, "\n%s: complete\n", test)
	return err
}

// OnSummary handles the summary event.
func (i Interactive) OnSummary(s *internal.Summary) error {
	_, err := fmt.Fprintf(i.out, "%15s: %s\n", "Server", s.Server)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(i.out, "%15s: %s\n", "Client", s.Client)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(i.out, "%15s: %7.1f %s\n", "Latency", s.RTT.Value, s.RTT.Unit)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(i.out, "%15s: %7.1f %s\n", "Download",
		s.Download.Value, s.Upload.Unit)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(i.out, "%15s: %7.1f %s\n", "Upload", s.Upload.Value,
		s.Upload.Unit)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(i.out, "%15s: %7.2f %s\n", "Retransmission",
		s.DownloadRetrans.Value, s.DownloadRetrans.Unit)
	if err != nil {
		return err
	}

	return nil
}
