package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/emitter"
	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

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

func (me mockedEmitter) OnStarting(test spec.TestKind) error {
	return me.StartingError
}

func (mockedEmitter) OnError(test spec.TestKind, err error) error {
	return nil
}

func (me mockedEmitter) OnConnected(test spec.TestKind, fqdn string) error {
	return me.ConnectedError
}

func (mockedEmitter) OnDownloadEvent(m *spec.Measurement) error {
	return nil
}

func (mockedEmitter) OnUploadEvent(m *spec.Measurement) error {
	return nil
}

func (me mockedEmitter) OnComplete(test spec.TestKind) error {
	return me.CompleteError
}

func TestRunTestOnStartingError(t *testing.T) {
	runner := runner{
		client: ndt7.NewClient(clientName, clientVersion),
		emitter: mockedEmitter{
			StartingError: errors.New("mocked error"),
		},
	}
	code := runner.runTest(
		context.Background(),
		"download",
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

func TestRunTestOnConnectedError(t *testing.T) {
	runner := runner{
		client: ndt7.NewClient(clientName, clientVersion),
		emitter: mockedEmitter{
			ConnectedError: errors.New("mocked error"),
		},
	}
	code := runner.runTest(
		context.Background(),
		"download",
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

func TestRunTestOnCompleteError(t *testing.T) {
	runner := runner{
		client: ndt7.NewClient(clientName, clientVersion),
		emitter: mockedEmitter{
			CompleteError: errors.New("mocked error"),
		},
	}
	code := runner.runTest(
		context.Background(),
		"download",
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

func TestRunTestEmitEventError(t *testing.T) {
	runner := runner{
		client:  ndt7.NewClient(clientName, clientVersion),
		emitter: mockedEmitter{},
	}
	code := runner.runTest(
		context.Background(),
		"download",
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

func TestBatchEmitterEventsOrderNormal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	writer := &mocks.SavingWriter{}
	runner := runner{
		client:  ndt7.NewClient(clientName, clientVersion),
		emitter: emitter.Batch{Writer: writer},
	}
	code := runner.runTest(
		context.Background(),
		"download",
		runner.client.StartDownload,
		runner.emitter.OnDownloadEvent,
	)
	if code != 0 {
		t.Fatal("expected zero return code here")
	}
	numLines := len(writer.Data)
	if numLines < 4 {
		t.Fatal("expected at least four lines")
	}
	for lineno, data := range writer.Data {
		var m struct {
			Key string
		}
		err := json.Unmarshal(data, &m)
		if err != nil {
			t.Fatal(err)
		}
		if lineno == 0 {
			if m.Key != "starting" {
				t.Fatal("unexpected first key")
			}
		} else if lineno == 1 {
			if m.Key != "connected" {
				t.Fatal("unexpected second key")
			}
		} else if lineno < numLines-1 {
			if m.Key != "measurement" {
				t.Fatalf("expected measurement key at line: %d; found %s",
					lineno, m.Key)
			}
		} else if lineno == numLines-1 {
			if m.Key != "complete" {
				t.Fatal("unexpected last key")
			}
		} else {
			t.Fatal("invalid index")
		}
	}
}

func TestBatchEmitterEventsOrderFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	writer := &mocks.SavingWriter{}
	runner := runner{
		client:  ndt7.NewClient(clientName, clientVersion),
		emitter: emitter.Batch{Writer: writer},
	}
	runner.client.MLabNSClient.BaseURL = "\t" // URL parser error
	code := runner.runTest(
		context.Background(),
		"download",
		runner.client.StartDownload,
		runner.emitter.OnDownloadEvent,
	)
	if code == 0 {
		t.Fatal("expected nonzero return code here")
	}
	numLines := len(writer.Data)
	if numLines != 3 {
		t.Fatal("expected at exactly three lines")
	}
	for lineno, data := range writer.Data {
		var m struct {
			Key string
		}
		err := json.Unmarshal(data, &m)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%d - %s\n", lineno, m.Key)
		if lineno == 0 {
			if m.Key != "starting" {
				t.Fatal("unexpected first key")
			}
		} else if lineno == 1 {
			if m.Key != "error" {
				t.Fatal("unexpected second key")
			}
		} else if lineno == 2 {
			if m.Key != "complete" {
				t.Fatal("unexpected third key")
			}
		} else {
			t.Fatal("invalid index")
		}
	}
}
