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
	conn := mocks.Conn{
		MessageByteArray: []byte("{}"),
		ReadMessageType:  websocket.TextMessage,
	}
	go func() {
		err := Run(ctx, &conn, outch)
		if err != nil {
			t.Fatal(err)
		}
	}()
	tot := 0
	// Drain the channel and count the number of Measurements read.
	for _ = range outch {
		tot++
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
