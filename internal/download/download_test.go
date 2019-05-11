package download

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestReadText is the case where we read text messages.
func TestReadText(t *testing.T) {
	orig := spec.Measurement{
		AppInfo: spec.AppInfo{
			NumBytes: 1234,
		},
		BBRInfo: spec.BBRInfo{
			MaxBandwidth: 12345,
			MinRTT:       1.2345,
		},
		Elapsed: 1.234,
		TCPInfo: spec.TCPInfo{
			SmoothedRTT: 1.2345,
			RTTVar:      1.2345,
		},
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	conn := mocks.Conn{
		ReadMessageType:      websocket.TextMessage,
		ReadMessageByteArray: data,
	}
	go func() {
		err := Run(ctx, &conn, outch)
		if err != nil {
			log.Fatal(err)
		}
	}()
	tot := 0
	for m := range outch {
		tot++
		if m.Origin != spec.OriginServer {
			t.Fatal("The origin is invalid")
		}
		if m.Direction != spec.DirectionDownload {
			t.Fatal("The direction is invalid")
		}
		// clear origin and direction for DeepEqual to work
		m.Origin = ""
		m.Direction = ""
		if !reflect.DeepEqual(orig, m) {
			t.Fatal("The two structs differ")
		}
	}
	if tot <= 0 {
		t.Fatal("Expected at least one measurement")
	}
}

// TestReadBinary is the case where we read binary messages.
func TestReadBinary(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	conn := mocks.Conn{
		ReadMessageType:      websocket.BinaryMessage,
		ReadMessageByteArray: []byte("12345678"),
	}
	go func() {
		err := Run(ctx, &conn, outch)
		if err != nil {
			t.Fatal(err)
		}
	}()
	for range outch {
		t.Fatal("We didn't expect a measurement here")
	}
}

// TestSetReadDeadlineError is the case where we get
// an error when setting the read deadline.
func TestSetReadDeadlineError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	mockedErr := errors.New("mocked error")
	conn := mocks.Conn{
		ReadMessageType:       websocket.TextMessage,
		ReadMessageByteArray:  []byte("{}"),
		SetReadDeadlineResult: mockedErr,
	}
	go func() {
		for range outch {
			t.Fatal("We didn't expect measurements here")
		}
	}()
	err := Run(ctx, &conn, outch)
	if err != mockedErr {
		t.Fatal("Not the error that we were expecting")
	}
}

// TestReadMessageError is the case where the
// ReadMessage method fails
func TestReadMessageError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	mockedErr := errors.New("mocked error")
	conn := mocks.Conn{
		ReadMessageType:      websocket.TextMessage,
		ReadMessageByteArray: []byte("{}"),
		ReadMessageResult:    mockedErr,
	}
	go func() {
		for range outch {
			t.Fatal("We didn't expect measurements here")
		}
	}()
	err := Run(ctx, &conn, outch)
	if err != mockedErr {
		t.Fatal("Not the error that we were expecting")
	}
}

// TestReadInvalidJSON is the case where we read
// an invalid JSON message.
func TestReadInvalidJSON(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	conn := mocks.Conn{
		ReadMessageType:      websocket.TextMessage,
		ReadMessageByteArray: []byte("{"),
	}
	go func() {
		for range outch {
			t.Fatal("We didn't expect measurements here")
		}
	}()
	err := Run(ctx, &conn, outch)
	if err == nil {
		t.Fatal("We expected to have an error here")
	}
}
