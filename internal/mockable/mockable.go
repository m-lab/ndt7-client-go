// Package mockable implements mocks.
package mockable

import (
	"time"

	"github.com/gorilla/websocket"
)

// Conn is a mockable websocket.Conn
type Conn struct {
	// CloseResult is the result of Conn.Close
	CloseResult error

	// ReadMessageByteArray is the byte array returned by Conn.ReadMessage
	ReadMessageByteArray []byte

	// ReadMessageResult is the result returned by conn.ReadMessage
	ReadMessageResult error

	// ReadMessageType is the type returned by conn.ReadMessage
	ReadMessageType int

	// SetReadDeadlineResult is the result returned by conn.SetReadDeadline
	SetReadDeadlineResult error

	// SetWriteDeadlineResult is the result returned by conn.SetWriteDeadline
	SetWriteDeadlineResult error

	// WritePreparedMessageResult is the result returned by conn.WritePreparedMessage
	WritePreparedMessageResult error
}

// Close closes the mocked connection
func (c *Conn) Close() error {
	return c.CloseResult
}

// ReadMessage reads a message from the mocked connection
func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	return c.ReadMessageType, c.ReadMessageByteArray, c.ReadMessageResult
}

// SetReadLimit sets the read limit of the mocked connection
func (*Conn) SetReadLimit(limit int64) {}

// SetReadDeadline sets the read deadline of the mocked connection
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.SetReadDeadlineResult
}

// SetWriteDeadline sets the write deadline of the mocked connection
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.SetWriteDeadlineResult
}

// WritePreparedMessage writes a prepared message on the mocked connection
func (c *Conn) WritePreparedMessage(pm *websocket.PreparedMessage) error {
	return c.WritePreparedMessageResult
}
