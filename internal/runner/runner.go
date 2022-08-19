package runner

import (
	"context"
<<<<<<< HEAD
=======
	"fmt"
>>>>>>> upstream/main
	"net"
	"time"

	"github.com/m-lab/go/memoryless"
	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/internal/emitter"
	"github.com/m-lab/ndt7-client-go/spec"
)

type RunnerOptions struct {
	Download, Upload bool
	Timeout time.Duration
	ClientFactory func() *ndt7.Client
}

type Runner struct {
	client  *ndt7.Client
	emitter emitter.Emitter
	ticker  *memoryless.Ticker
	opt     RunnerOptions
}

func New(opt RunnerOptions, emitter emitter.Emitter, ticker *memoryless.Ticker) *Runner {
	return &Runner{
		opt: opt,
		emitter: emitter,
		ticker: ticker,
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
		<- r.ticker.C
	}
}

func makeSummary(FQDN string, results map[spec.TestKind]*ndt7.LatestMeasurements) *emitter.Summary {

	s := emitter.NewSummary(FQDN)

	if results[spec.TestDownload] != nil &&
		results[spec.TestDownload].ConnectionInfo != nil {
		// Get UUID, ClientIP and ServerIP from ConnectionInfo.
		s.DownloadUUID = results[spec.TestDownload].ConnectionInfo.UUID

		clientIP, _, err := net.SplitHostPort(results[spec.TestDownload].ConnectionInfo.Client)
		if err == nil {
			s.ClientIP = clientIP
		}

		serverIP, _, err := net.SplitHostPort(results[spec.TestDownload].ConnectionInfo.Server)
		if err == nil {
			s.ServerIP = serverIP
		}
	}

	// Download comes from the client-side Measurement during the download
	// test. DownloadRetrans and MinRTT come from the server-side Measurement,
	// if it includes a TCPInfo object.
	if dl, ok := results[spec.TestDownload]; ok {
		if dl.Client.AppInfo != nil && dl.Client.AppInfo.ElapsedTime > 0 {
			elapsed := float64(dl.Client.AppInfo.ElapsedTime) / 1e06
			s.Download = emitter.ValueUnitPair{
				Value: (8.0 * float64(dl.Client.AppInfo.NumBytes)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
		if dl.Server.TCPInfo != nil {
			if dl.Server.TCPInfo.BytesSent > 0 {
				s.DownloadRetrans = emitter.ValueUnitPair{
					Value: float64(dl.Server.TCPInfo.BytesRetrans) / float64(dl.Server.TCPInfo.BytesSent) * 100,
					Unit:  "%",
				}
			}
			s.MinRTT = emitter.ValueUnitPair{
				Value: float64(dl.Server.TCPInfo.MinRTT) / 1000,
				Unit:  "ms",
			}
		}
	}
	// The upload rate comes from the receiver (the server). Currently
	// ndt-server only provides network-level throughput via TCPInfo.
	// TODO: Use AppInfo for application-level measurements when available.
	if ul, ok := results[spec.TestUpload]; ok {
		if ul.Server.TCPInfo != nil && ul.Server.TCPInfo.BytesReceived > 0 {
			elapsed := float64(ul.Server.TCPInfo.ElapsedTime) / 1e06
			s.Upload = emitter.ValueUnitPair{
				Value: (8.0 * float64(ul.Server.TCPInfo.BytesReceived)) /
					elapsed / (1000.0 * 1000.0),
				Unit: "Mbit/s",
			}
		}
	}

	return s
}

