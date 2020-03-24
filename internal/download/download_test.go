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
	"github.com/m-lab/tcp-info/inetdiag"
	"github.com/m-lab/tcp-info/tcp"
)

func TestReadText(t *testing.T) {
	orig := spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: 1234000,
			NumBytes:    1234,
		},
		BBRInfo: &spec.BBRInfo{
			BBRInfo: inetdiag.BBRInfo{
				BW:     12345,
				MinRTT: 12345,
			},
		},
		TCPInfo: &spec.TCPInfo{
			ElapsedTime: 1234000,
			LinuxTCPInfo: tcp.LinuxTCPInfo{
				RTT:    12345000,
				RTTVar: 12345000,
			},
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
		if m.Origin == spec.OriginClient {
			// We test this specific case in another test. Just do not
			// fallthrough because we're gonna fail otherwise.
			continue
		}
		if m.Origin != spec.OriginServer {
			t.Fatal("The origin is invalid")
		}
		if m.Test != spec.TestDownload {
			t.Fatal("The test is invalid")
		}
		// clear origin and direction for DeepEqual to work
		m.Origin = ""
		m.Test = ""
		if !reflect.DeepEqual(orig, m) {
			t.Fatal("The two structs differ")
		}
	}
	if tot <= 0 {
		t.Fatal("Expected at least one measurement")
	}
}

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
	prev := spec.Measurement{
		AppInfo: &spec.AppInfo{},
	}
	for m := range outch {
		if m.Origin != spec.OriginClient {
			t.Fatal("unexpected measurement origin")
		}
		if m.Test != spec.TestDownload {
			t.Fatal("unexpected test name")
		}
		if m.AppInfo == nil {
			t.Fatal("m.AppInfo is nil")
		}
		if m.AppInfo.ElapsedTime <= prev.AppInfo.ElapsedTime {
			t.Fatal("the time is not increasing")
		}
		if m.AppInfo.NumBytes <= prev.AppInfo.NumBytes {
			t.Fatal("the number of bytes is not increasing")
		}
		prev = m
	}
}

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
