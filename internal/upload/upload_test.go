package upload

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestNormal is the normal test case
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
	prev := spec.Measurement{}
	tot := 0
	for m := range outch {
		tot++
		if m.Origin != spec.OriginClient {
			t.Fatal("The origin is wrong")
		}
		if m.Direction != spec.DirectionUpload {
			t.Fatal("The direction is wrong")
		}
		if m.Elapsed <= prev.Elapsed {
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

// TestSetReadDealineError ensures that we deal with
// the case where SetReadDeadline fails.
func TestSetReadDeadlineError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	conn := mocks.Conn{
		SetReadDeadlineResult: mockedErr,
	}
	err := ignoreIncoming(&conn)
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}

// TestReadMessageError ensures that we deal with
// the case where ReadMessage fails.
func TestReadMessageError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	conn := mocks.Conn{
		ReadMessageResult: mockedErr,
	}
	err := ignoreIncoming(&conn)
	if err != mockedErr {
		t.Fatal("Not the error we expected")
	}
}

// TestReadNonTextMessageError ensures that we deal with the
// case where ReadMessage returns a non text message.
func TestReadNonTextMessageError(t *testing.T) {
	conn := mocks.Conn{
		ReadMessageType:      websocket.BinaryMessage,
		ReadMessageByteArray: []byte("abcdef"),
	}
	err := ignoreIncoming(&conn)
	if err != errNonTextMessage {
		t.Fatal("Not the error we expected")
	}
}

// TestMakePreparedMessageError ensures that we deal with
// the case where makePreparedMessage fails.
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

// TestSetWriteDeadlineError ensures that we deal with
// the case where SetWriteDeadline fails.
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

// TestWritePreparedMessageError ensures that we deal with
// the case where WritePreparedMessage fails.
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
