package emitter

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/m-lab/ndt7-client-go/cmd/ndt7-client/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

func TestHumanReadableOnStarting(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnStarting("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\rstarting download")) {
		t.Fatal("unexpected output")
	}
}

func TestHumanReadableOnStartingFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnStarting("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnError(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnError("download", errors.New("mocked error"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\rdownload failed: mocked error\n")) {
		t.Fatal("unexpected output")
	}
}

func TestHumanReadableOnErrorFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnError("download", errors.New("some error"))
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnConnected(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnConnected("download", "FQDN")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\rdownload in progress with FQDN\n")) {
		t.Fatal("unexpected output")
	}
}

func TestHumanReadableOnConnectedFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnConnected("download", "FQDN")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnDownloadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnDownloadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 3000000,
			NumBytes:    100000000,
		},
		Origin: spec.OriginClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(
		sw.Data[0],
		[]byte("\rAvg. speed  :   266.7 Mbit/s"),
	) {
		t.Fatal("unexpected output")
	}
}

func TestHumanReadableOnDownloadEventFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnDownloadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234,
		},
		Origin: spec.OriginClient,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableIgnoresServerData(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
		Origin: spec.OriginServer,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 0 {
		t.Fatal("invalid length")
	}
}

func TestHumanReadableOnUploadEvent(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 3000000,
			NumBytes:    100000000,
		},
		Origin: spec.OriginClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(
		sw.Data[0],
		[]byte("\rAvg. speed  :   266.7 Mbit/s"),
	) {
		t.Fatal("unexpected output")
	}
}

func TestHumanReadableOnUploadEventSafetyCheck(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
		Origin: spec.OriginClient,
	})
	if err == nil {
		t.Fatal("We did expect an error here")
	}
	if len(sw.Data) != 0 {
		t.Fatal("Some data was written and it shouldn't have")
	}
}

func TestHumanReadableOnUploadEventDivideByZero(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{},
		Origin:  spec.OriginClient,
	})
	if err == nil {
		t.Fatal("We did expect an error here")
	}
	if len(sw.Data) != 0 {
		t.Fatal("Some data was written and it shouldn't have")
	}
}

func TestHumanReadableOnUploadEventFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnUploadEvent(&spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234,
		},
		Origin: spec.OriginClient,
	})
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnComplete(t *testing.T) {
	sw := &mocks.SavingWriter{}
	hr := HumanReadable{sw}
	err := hr.OnComplete("download")
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if !reflect.DeepEqual(sw.Data[0], []byte("\ndownload: complete\n")) {
		t.Fatal("unexpected output")
	}
}

func TestHumanReadableOnCompleteFailure(t *testing.T) {
	hr := HumanReadable{&mocks.FailingWriter{}}
	err := hr.OnComplete("download")
	if err != mocks.ErrMocked {
		t.Fatal("Not the error we expected")
	}
}

func TestHumanReadableOnSummary(t *testing.T) {
	expected := `         Server: test
         Client: test
        Latency:    10.0 ms
       Download:   100.0 Mbit/s
         Upload:   100.0 Mbit/s
 Retransmission:    1.00 %
`
	summary := &Summary{
		ClientIP:   "test",
		ServerFQDN: "test",
		Download: ValueUnitPair{
			Value: 100.0,
			Unit:  "Mbit/s",
		},
		Upload: ValueUnitPair{
			Value: 100.0,
			Unit:  "Mbit/s",
		},
		DownloadRetrans: ValueUnitPair{
			Value: 1.0,
			Unit:  "%",
		},
		MinRTT: ValueUnitPair{
			Value: 10.0,
			Unit:  "ms",
		},
	}
	sw := &mocks.SavingWriter{}
	j := HumanReadable{sw}
	err := j.OnSummary(summary)
	if err != nil {
		t.Fatal(err)
	}

	if len(sw.Data) != 1 {
		t.Fatal("invalid length")
	}
	if string(sw.Data[0]) != expected {
		fmt.Println(string(sw.Data[0]))
		fmt.Println(expected)
		t.Fatal("OnSummary(): unexpected data")
	}
}

func TestHumanReadableOnSummaryFailure(t *testing.T) {
	sw := &mocks.FailingWriter{}
	j := HumanReadable{sw}
	err := j.OnSummary(&Summary{})
	if err == nil {
		t.Fatal("OnSummary(): expected err, got nil")
	}
}

func TestNewHumanReadableConstructor(t *testing.T) {
	hr := NewHumanReadable()
	if hr == nil {
		t.Fatal("NewHumanReadable() did not return a HumanReadable")
	}
}

func TestNewHumanReadableWithWriter(t *testing.T) {
	hr := NewHumanReadableWithWriter(&mocks.SavingWriter{})
	if hr == nil {
		t.Fatal("NewHumanReadableWithWriter() did not return a HumanReadable")
	}
}
