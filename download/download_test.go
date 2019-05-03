package download

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/spec"
)

type mockedConn struct {
	Message        []byte
	ReadErr        error
	SetDeadlineErr error
	Type           int
}

func (*mockedConn) Close() error {
	return nil
}

func (c *mockedConn) ReadMessage() (messageType int, p []byte, err error) {
	return c.Type, c.Message, c.ReadErr
}

func (*mockedConn) SetReadLimit(limit int64) {}

func (c *mockedConn) SetReadDeadline(t time.Time) error {
	return c.SetDeadlineErr
}

// TestReadText is the case where we read text messages.
func TestReadText(t *testing.T) {
	outch := make(chan spec.Measurement)
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(time.Second),
	)
	conn := mockedConn{
		Type:    websocket.TextMessage,
		Message: []byte("{}"),
	}
	defer cancel()
	go mockableRun(ctx, &conn, outch)
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
	conn := mockedConn{
		Type:    websocket.BinaryMessage,
		Message: []byte("{}"),
	}
	defer cancel()
	go mockableRun(ctx, &conn, outch)
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
	conn := mockedConn{
		Type:           websocket.TextMessage,
		Message:        []byte("{}"),
		SetDeadlineErr: errors.New("mocked error"),
	}
	defer cancel()
	go mockableRun(ctx, &conn, outch)
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
	conn := mockedConn{
		Type:    websocket.TextMessage,
		Message: []byte("{}"),
		ReadErr: errors.New("mocked error"),
	}
	defer cancel()
	go mockableRun(ctx, &conn, outch)
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
	conn := mockedConn{
		Type:    websocket.TextMessage,
		Message: []byte("{"),
	}
	defer cancel()
	go mockableRun(ctx, &conn, outch)
	for range outch {
		// ignore
	}
}
