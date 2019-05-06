// Package mockable implements mocks.
package mockable

import (
	"bytes"
	"io"
	"net/http"
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

// reponseBody is a fake HTTP response body.
type reponseBody struct {
	reader io.Reader
}

// newResponseBody creates a new response body.
func newResponseBody(data []byte) io.ReadCloser {
	return &reponseBody{
		reader: bytes.NewReader(data),
	}
}

// Read reads the response body.
func (r *reponseBody) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

// Close closes the response body.
func (r *reponseBody) Close() error {
	return nil
}

// HTTPRequestor is a mockable HTTP requestor
type HTTPRequestor struct {
	// Response is the response bound to this requestor.
	Response *http.Response

	// Error is the error to return.
	Error error
}

// NewHTTPRequestor returns a mockable HTTP requestor.
func NewHTTPRequestor(code int, body []byte, err error) *HTTPRequestor {
	return &HTTPRequestor{
		Error: err,
		Response: &http.Response{
			Body:       newResponseBody(body),
			StatusCode: code,
		},
	}
}

// Do executes the request to return a response or an error.
func (r *HTTPRequestor) Do(req *http.Request) (*http.Response, error) {
	return r.Response, r.Error
}
