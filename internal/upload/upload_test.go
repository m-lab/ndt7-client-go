package upload

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/mockable"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestNormal is the normal test case
func TestNormal(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mockable.Conn{}
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
	conn := mockable.Conn{
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
	conn := mockable.Conn{
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
	conn := mockable.Conn{}
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
	conn := mockable.Conn{
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
	conn := mockable.Conn{
		WritePreparedMessageResult: errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}
