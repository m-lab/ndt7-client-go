package download

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/mocks"
	"github.com/m-lab/ndt7-client-go/spec"
)

// TestReadText is the case where we read text messages.
func TestReadText(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		ReadMessageType:      websocket.TextMessage,
		ReadMessageByteArray: []byte("{}"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestReadBinary is the case where we read binary messages.
func TestReadBinary(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		ReadMessageType:      websocket.BinaryMessage,
		ReadMessageByteArray: []byte("{}"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestSetReadDeadlineError is the case where we get
// an error when setting the read deadline.
func TestSetReadDeadlineError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		ReadMessageType:       websocket.TextMessage,
		ReadMessageByteArray:  []byte("{}"),
		SetReadDeadlineResult: errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestReadMessageError is the case where the
// ReadMessage method fails
func TestReadMessageError(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		ReadMessageType:      websocket.TextMessage,
		ReadMessageByteArray: []byte("{}"),
		ReadMessageResult:    errors.New("mocked error"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}

// TestReadInvalidJSON is the case where we read
// an invalid JSON message.
func TestReadInvalidJSON(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mocks.Conn{
		ReadMessageType:      websocket.TextMessage,
		ReadMessageByteArray: []byte("{"),
	}
	defer cancel()
	go Run(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}
