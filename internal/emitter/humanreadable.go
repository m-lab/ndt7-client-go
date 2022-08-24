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
	const summaryFormat = `%15s: %s
%15s: %s
%15s: %7.1f %s
`
	_, err := fmt.Fprintf(h.out, summaryFormat,
		"Server", s.ServerFQDN,
		"Client", s.ClientIP,
		"Latency", s.MinRTT.Value, s.MinRTT.Unit)
	if err != nil {
		return err
	}

	if s.Download.Value != 0.0 {
		_, err := fmt.Fprintf(h.out, "%15s: %7.1f %s\n",
			"Download", s.Download.Value, s.Download.Unit)
		if err != nil {
			return err
		}
	}

	if s.Upload.Value != 0.0 {
		_, err := fmt.Fprintf(h.out, "%15s: %7.1f %s\n",
			"Upload", s.Upload.Value, s.Upload.Unit)
		if err != nil {
			return err
		}
	}

	if s.DownloadRetrans.Value != 0.0 {
		_, err := fmt.Fprintf(h.out, "%15s: %7.2f %s\n",
			"Retransmission", s.DownloadRetrans.Value, s.DownloadRetrans.Unit)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(h.out)

	return nil
}
