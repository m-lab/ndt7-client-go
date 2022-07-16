package runner

import (
	"context"
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

func NewRunner(opt RunnerOptions, emitter emitter.Emitter, ticker *memoryless.Ticker) *Runner {
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
) int {
	ch, err := start(ctx)
	if err != nil {
		r.emitter.OnError(test, err)
		return 1
	}
	err = r.emitter.OnConnected(test, r.client.FQDN)
	if err != nil {
		return 1
	}
	for ev := range ch {
		err = emitEvent(&ev)
		if err != nil {
			return 1
		}
	}
	return 0
}

func (r Runner) runTest(
	ctx context.Context, test spec.TestKind,
	start func(context.Context) (<-chan spec.Measurement, error),
	emitEvent func(m *spec.Measurement) error,
) int {
	// Implementation note: we want to always emit the initial and the
	// final events regardless of how the actual test goes. What's more,
	// we want the exit code to be nonzero in case of any error.
	err := r.emitter.OnStarting(test)
	if err != nil {
		return 1
	}
	code := r.doRunTest(ctx, test, start, emitEvent)
	err = r.emitter.OnComplete(test)
	if err != nil {
		return 1
	}
	return code
}

func (r Runner) runDownload(ctx context.Context) int {
	return r.runTest(ctx, spec.TestDownload, r.client.StartDownload,
		r.emitter.OnDownloadEvent)
}

func (r Runner) runUpload(ctx context.Context) int {
	return r.runTest(ctx, spec.TestUpload, r.client.StartUpload,
		r.emitter.OnUploadEvent)
}

func (r Runner) RunTestsOnce() int {
	var code int
	
	ctx, cancel := context.WithTimeout(context.Background(), r.opt.Timeout)
	defer cancel()

	r.client = r.opt.ClientFactory()

	if r.opt.Download {
		code += r.runDownload(ctx)
	}
	if r.opt.Upload {
		code += r.runUpload(ctx)
	}

	s := makeSummary(r.client.FQDN, r.client.Results())
	r.emitter.OnSummary(s)

	return code
}

func (r Runner) RunTestsInLoop() int {
	var code int
	for ;; {
		code = r.RunTestsOnce()

		// Wait
		<- r.ticker.C
	}

	return code
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

