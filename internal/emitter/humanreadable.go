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
	switch m.Test {
	case spec.TestDownload:
		if m.Origin == spec.OriginClient {
			if m.AppInfo == nil || m.AppInfo.ElapsedTime <= 0 {
				return errors.New("missing AppInfo or invalid ElapsedTime")
			}
			elapsed := float64(m.AppInfo.ElapsedTime)
			v := 8.0 * float64(m.AppInfo.NumBytes) / elapsed
			_, err := fmt.Fprintf(h.out, "\rAvg. speed  : %7.1f Mbit/s", v)
			return err
		}
	case spec.TestUpload:
		if m.Origin == spec.OriginServer {
			if m.TCPInfo == nil || m.TCPInfo.ElapsedTime <= 0 {
				return errors.New("missing TCPInfo or invalid ElapsedTime")
			}
			elapsed := float64(m.TCPInfo.ElapsedTime)
			v := 8.0 * float64(m.TCPInfo.BytesReceived) / elapsed
			_, err := fmt.Fprintf(h.out, "\rAvg. speed  : %7.1f Mbit/s", v)
			return err
		}
	}
	return nil
}

// OnComplete handles the complete event
func (h HumanReadable) OnComplete(test spec.TestKind) error {
	_, err := fmt.Fprintf(h.out, "\n%s: complete\n", test)
	return err
}

// OnSummary handles the summary event.
func (h HumanReadable) OnSummary(s *Summary) error {
	const summaryHeaderFormat = `
Test results

%10s: %s
%10s: %s
`
	const downloadFormat = `
%22s
%15s: %7.1f %s
%15s: %7.1f %s
%15s: %7.1f %s
`
	const uploadFormat = `
%20s
%15s: %7.1f %s
%15s: %7.1f %s
`
	_, err := fmt.Fprintf(h.out, summaryHeaderFormat,
		"Server", s.ServerFQDN,
		"Client", s.ClientIP)
	if err != nil {
		return err
	}

	if s.Download != nil {
		_, err := fmt.Fprintf(h.out, downloadFormat, "Download",
			"Throughput", s.Download.Throughput.Value, s.Download.Throughput.Unit,
			"Latency", s.Download.Latency.Value, s.Download.Latency.Unit,
			"Retransmission", s.Download.Retransmission.Value, s.Download.Retransmission.Unit)
		if err != nil {
			return err
		}
	}

	if s.Upload != nil {
		_, err := fmt.Fprintf(h.out, uploadFormat, "Upload",
			"Throughput", s.Upload.Throughput.Value, s.Upload.Throughput.Unit,
			"Latency", s.Upload.Latency.Value, s.Upload.Latency.Unit)
		if err != nil {
			return err
		}
	}

	return nil
}
