package main

import (
	"context"
	"errors"
	"testing"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestNormalUsage tests ndt7-client w/o any command line arguments.
func TestNormalUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	main()
	osExit = savedFunc
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

// TestBatchUsage tests the -batch use case.
func TestBatchUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	*flagBatch = true
	main()
	*flagBatch = false
	osExit = savedFunc
	if exitval != 0 {
		t.Fatal("expected zero return code here")
	}
}

// TestDownloadError tests the case where a subtest fails.
func TestDownloadError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	exitval := 0
	savedFunc := osExit
	osExit = func(code int) {
		exitval = code
	}
	*flagHostname = "\t" // fail parsing
	main()
	*flagHostname = ""
	osExit = savedFunc
	if exitval == 0 {
		t.Fatal("expected nonzero return code here")
	}
}

type mockedEmitter struct {
	StartingError  error
	ConnectedError error
	CompleteError  error
}

func (me mockedEmitter) OnStarting(subtest string) error {
	return me.StartingError
}

func (mockedEmitter) OnError(subtest string, err error) error {
	return nil
}

func (me mockedEmitter) OnConnected(subtest, fqdn string) error {
	return me.ConnectedError
}

func (mockedEmitter) OnDownloadEvent(m *spec.Measurement) error {
	return nil
}

func (mockedEmitter) OnUploadEvent(m *spec.Measurement) error {
	return nil
}

func (me mockedEmitter) OnComplete(subtest string) error {
	return me.CompleteError
}

// TestRunSubtestOnStartingError deals with the case where
// the emitter.OnStarting function fails.
func TestRunSubtestOnStartingError(t *testing.T) {
	runner := runner{
		client: ndt7.NewClient(userAgent),
		emitter: mockedEmitter{
			StartingError: errors.New("mocked error"),
		},
	}
	code := runner.runSubtest(
		context.Background(),
		"subtest",
		func(context.Context) (<-chan spec.Measurement, error) {
			out := make(chan spec.Measurement)
			close(out)
			return out, nil
		},
		func(m *spec.Measurement) error {
			return nil
		},
	)
	if code == 0 {
		t.Fatal("expected nonzero return code here")
	}
}

// TestRunSubtestOnConnectedError deals with the case where
// the emitter.OnConnected function fails.
func TestRunSubtestOnConnectedError(t *testing.T) {
	runner := runner{
		client: ndt7.NewClient(userAgent),
		emitter: mockedEmitter{
			ConnectedError: errors.New("mocked error"),
		},
	}
	code := runner.runSubtest(
		context.Background(),
		"subtest",
		func(context.Context) (<-chan spec.Measurement, error) {
			out := make(chan spec.Measurement)
			close(out)
			return out, nil
		},
		func(m *spec.Measurement) error {
			return nil
		},
	)
	if code == 0 {
		t.Fatal("expected nonzero return code here")
	}
}

// TestRunSubtestOnCompleteError deals with the case where
// the emitter.OnComplete function fails.
func TestRunSubtestOnCompleteError(t *testing.T) {
	runner := runner{
		client: ndt7.NewClient(userAgent),
		emitter: mockedEmitter{
			CompleteError: errors.New("mocked error"),
		},
	}
	code := runner.runSubtest(
		context.Background(),
		"subtest",
		func(context.Context) (<-chan spec.Measurement, error) {
			out := make(chan spec.Measurement)
			close(out)
			return out, nil
		},
		func(m *spec.Measurement) error {
			return nil
		},
	)
	if code == 0 {
		t.Fatal("expected nonzero return code here")
	}
}

// TestRunSubtestEmitEventError deals with the case where
// the emitEvent function fails.
func TestRunSubtestEmitEventError(t *testing.T) {
	runner := runner{
		client:  ndt7.NewClient(userAgent),
		emitter: mockedEmitter{},
	}
	code := runner.runSubtest(
		context.Background(),
		"subtest",
		func(context.Context) (<-chan spec.Measurement, error) {
			out := make(chan spec.Measurement)
			go func() {
				defer close(out)
				out <- spec.Measurement{}
			}()
			return out, nil
		},
		func(m *spec.Measurement) error {
			return errors.New("mocked error")
		},
	)
	if code == 0 {
		t.Fatal("expected nonzero return code here")
	}
}
