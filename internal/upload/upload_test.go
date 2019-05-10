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
	conn := mocks.Conn{}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestSetReadDealindError ensures that we deal with
// the case where SetReadDeadline fails.
func TestSetReadDeadlineError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		SetReadDeadlineResult: errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestReadMessageError ensures that we deal with
// the case where ReadMessage fails.
func TestReadMessageError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		ReadMessageResult: errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestMakePreparedMessageError ensures that we deal with
// the case where makePreparedMessage fails.
func TestMakePreparedMessageError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	savedFunc := makePreparedMessage
	makePreparedMessage = func(size int) (*websocket.PreparedMessage, error) {
		return nil, errors.New("mocked error")
	}
	conn := mocks.Conn{}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
	makePreparedMessage = savedFunc
}

// TestSetWriteDeadlineError ensures that we deal with
// the case where SetWriteDeadline fails.
func TestSetWriteDeadlineError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		SetWriteDeadlineResult: errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestWritePreparedMessageError ensures that we deal with
// the case where WritePreparedMessage fails.
func TestWritePreparedMessageError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		WritePreparedMessageResult: errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}
