package runner

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/m-lab/go/memoryless"
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/internal/emitter"
	"github.com/m-lab/ndt7-client-go/spec"
)

type RunnerOptions struct {
	Download, Upload bool
	Timeout          time.Duration
	ClientFactory    func() *ndt7.Client
}

type Runner struct {
	client  *ndt7.Client
	emitter emitter.Emitter
	ticker  *memoryless.Ticker
	opt     RunnerOptions
}

func New(opt RunnerOptions, emitter emitter.Emitter, ticker *memoryless.Ticker) *Runner {
	return &Runner{
		opt:     opt,
		emitter: emitter,
		ticker:  ticker,
	}
}

func (r Runner) doRunTest(
	ctx context.Context, test spec.TestKind,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) error {
	ch, err := start(ctx)
	if err != nil {
		r.emitter.OnError(test, err)
		return fmt.Errorf("Failed to start test %v: %v", test, err)
	}
	err = r.emitter.OnConnected(test, r.client.FQDN)
	if err != nil {
		return fmt.Errorf("Failed to emit connection event for test %v: %v", test, err)
	}
	for ev := range ch {
		err = emitEvent(&ev)
		if err != nil {
			return fmt.Errorf("Failed to emit event for test %v: %v", test, err)
		}
	}
	return nil
}

func (r Runner) runTest(
	ctx context.Context, test spec.TestKind,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) error {
	// Implementation note: we want to always emit the initial and the
	// final events regardless of how the actual test goes. What's more,
	// we want the exit code to be nonzero in case of any error.
	if err := r.emitter.OnStarting(test); err != nil {
		return fmt.Errorf("Failed to start test %v: %v", test, err)
	}
	err := r.doRunTest(ctx, test, start, emitEvent)
	if emitErr := r.emitter.OnComplete(test); emitErr != nil {
		return fmt.Errorf("Failed to emit completion event for test %v: %v", test, emitErr)
	}
	if err != nil {
		return fmt.Errorf("Failed to run test %v: %v", test, err)
	}
	return nil
}

func (r Runner) runDownload(ctx context.Context) error {
	return r.runTest(ctx, spec.TestDownload, r.client.StartDownload,
		r.emitter.OnDownloadEvent)
}

func (r Runner) runUpload(ctx context.Context) error {
	return r.runTest(ctx, spec.TestUpload, r.client.StartUpload,
		r.emitter.OnUploadEvent)
}

func (r Runner) RunTestsOnce() []error {
	errs := make([]error, 0)

	ctx, cancel := context.WithTimeout(context.Background(), r.opt.Timeout)
	defer cancel()

	r.client = r.opt.ClientFactory()

	if r.opt.Download {
		err := r.runDownload(ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if r.opt.Upload {
		err := r.runUpload(ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}

	s := makeSummary(r.client.FQDN, r.client.Results())
	r.emitter.OnSummary(s)

	return errs
}

func (r Runner) RunTestsInLoop() {
	for {
		// We ignore the return value here since we rely on the emiiters
		// to report that the measurement failed. We want to continue
		// even when there is an error.
		_ = r.RunTestsOnce()

		// Wait
		<-r.ticker.C
	}
}

func makeSummary(FQDN string, results map[spec.TestKind]*ndt7.LatestMeasurements) *emitter.Summary {

	s := emitter.NewSummary(FQDN)

	var server, client string

	// If there is a download result, populate the summary.
	if dl, ok := results[spec.TestDownload]; ok {
		s.Download = &emitter.SubtestSummary{}
		if dl.ConnectionInfo != nil {
			connInfo := dl.ConnectionInfo
			s.Download.UUID = connInfo.UUID
			client = connInfo.Client
			server = connInfo.Server
		}
		// Read the throughput at the receiver (i.e. the client).
		if dl.Client.AppInfo != nil &&
			dl.Client.AppInfo.ElapsedTime > 0 {
			appInfo := dl.Client.AppInfo
			elapsed := float64(appInfo.ElapsedTime) / 1e06
			s.Download.Throughput = emitter.ValueUnitPair{
				Value: (8.0 * float64(appInfo.NumBytes)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
		if dl.Server.TCPInfo != nil {
			tcpInfo := dl.Server.TCPInfo
			// Read the retransmission rate at the sender.
			if tcpInfo.BytesSent > 0 {
				s.Download.Retransmission = emitter.ValueUnitPair{
					Value: float64(tcpInfo.BytesRetrans) /
						float64(tcpInfo.BytesSent) * 100,
					Unit: "%",
				}
			}
			// Read the latency at the sender.
			s.Download.Latency = emitter.ValueUnitPair{
				Value: float64(tcpInfo.MinRTT) / 1000,
				Unit:  "ms",
			}
		}
	}

	if ul, ok := results[spec.TestUpload]; ok {
		s.Upload = &emitter.SubtestSummary{}
		if ul.ConnectionInfo != nil {
			connInfo := ul.ConnectionInfo
			s.Upload.UUID = connInfo.UUID
			client = connInfo.Client
			server = connInfo.Server
		}
		if ul.Server.TCPInfo != nil {
			tcpInfo := ul.Server.TCPInfo
			// Read the throughput at the receiver (i.e. the server).
			if tcpInfo.ElapsedTime > 0 {
				elapsed := float64(tcpInfo.ElapsedTime) / 1e06
				s.Upload.Throughput = emitter.ValueUnitPair{
					Value: (8.0 * float64(tcpInfo.BytesReceived)) /
						elapsed / (1000.0 * 1000.0),
					Unit: "Mbit/s",
				}
			}
			// Read the latency at the receiver.
			s.Upload.Latency = emitter.ValueUnitPair{
				Value: float64(tcpInfo.MinRTT) / 1000,
				Unit:  "ms",
			}
		}
	}

	clientIP, _, err := net.SplitHostPort(client)
	if err == nil {
		s.ClientIP = clientIP
	}
	serverIP, _, err := net.SplitHostPort(server)
	if err == nil {
		s.ServerIP = serverIP
	}

	return s
}
