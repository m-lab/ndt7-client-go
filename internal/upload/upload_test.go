package upload

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

func TestNormal(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	conn := mocks.Conn{}
	go func() {
		err := Run(ctx, &conn, outch)
		if err != nil {
			t.Fatal(err)
		}
	}()
	prev := spec.Measurement{
		AppInfo: &spec.AppInfo{},
	}
	tot := 0
	for m := range outch {
		tot++
		if m.Origin != spec.OriginClient {
			t.Fatal("The origin is wrong")
		}
		if m.Test != spec.TestUpload {
			t.Fatal("The test is wrong")
		}
		if m.AppInfo == nil {
			t.Fatal("m.AppInfo is nil")
		}
		if m.AppInfo.ElapsedTime <= prev.AppInfo.ElapsedTime {
			t.Fatal("Time is not increasing")
		}
		// Note: it can stay constant when we're servicing
		// a TCP timeout longer than the update interval
		if m.AppInfo.NumBytes < prev.AppInfo.NumBytes {
			t.Fatal("Number of bytes is decreasing")
		}
		prev = m
	}
	if tot <= 0 {
		t.Fatal("Expected at least one message")
	}
}

func TestSetReadDeadlineError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	conn := mocks.Conn{
		SetReadDeadlineResult: mockedErr,
	}
	ch := make(chan spec.Measurement, 128)
	errs := make(chan error)
	go readcounterflow(context.Background(), &conn, ch, errs)
	err := <-errs
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}

func TestReadMessageError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	conn := mocks.Conn{
		ReadMessageResult: mockedErr,
	}
	ch := make(chan spec.Measurement, 128)
	errs := make(chan error)
	go readcounterflow(context.Background(), &conn, ch, errs)
	err := <-errs
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}

func TestReadNonTextMessageError(t *testing.T) {
	conn := mocks.Conn{
		ReadMessageType:  websocket.BinaryMessage,
		MessageByteArray: []byte("abcdef"),
	}
	ch := make(chan spec.Measurement, 128)
	errs := make(chan error)
	go readcounterflow(context.Background(), &conn, ch, errs)
	err := <-errs
	if err != errNonTextMessage {
		t.Fatal("Not the error we expected")
	}
}

func TestReadNonJSONError(t *testing.T) {
	conn := mocks.Conn{
		ReadMessageType:  websocket.TextMessage,
		MessageByteArray: []byte("{"),
	}
	ch := make(chan spec.Measurement, 128)
	errs := make(chan error)
	go readcounterflow(context.Background(), &conn, ch, errs)
	err := <-errs
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatal("Not the error we expected")
	}
}

func TestReadGoodMessage(t *testing.T) {
	conn := mocks.Conn{
		ReadMessageType:  websocket.TextMessage,
		MessageByteArray: []byte("{}"),
	}
	ch := make(chan spec.Measurement, 128)
	var count int64
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for range ch {
			count++
			cancel()
		}
	}()
	errs := make(chan error)
	go readcounterflow(ctx, &conn, ch, errs)
	err := <-errs
	if err != nil {
		t.Fatal(err)
	}
}

func TestMakePreparedMessageError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	outch := make(chan int64)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	savedFunc := makePreparedMessage
	makePreparedMessage = func(size int) (*websocket.PreparedMessage, error) {
		return nil, mockedErr
	}
	conn := mocks.Conn{}
	go func() {
		for range outch {
			t.Fatal("Did not expect messages here")
		}
	}()
	err := upload(ctx, &conn, outch)
	makePreparedMessage = savedFunc
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}

func TestSetWriteDeadlineError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	outch := make(chan int64)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	conn := mocks.Conn{
		SetWriteDeadlineResult: mockedErr,
	}
	go func() {
		for range outch {
			t.Fatal("Did not expect messages here")
		}
	}()
	err := upload(ctx, &conn, outch)
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}

func TestWritePreparedMessageError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	outch := make(chan int64)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	defer cancel()
	conn := mocks.Conn{
		WritePreparedMessageResult: mockedErr,
	}
	go func() {
		for range outch {
			t.Fatal("Did not expect messages here")
		}
	}()
	err := upload(ctx, &conn, outch)
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}
