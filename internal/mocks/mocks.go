// Package mocks implements mocks.
package mocks

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

type httpTransport struct {
	Response *http.Response
	Error    error
}

// NewHTTPClient returns a mocked *http.Client.
func NewHTTPClient(code int, body []byte, err error) *http.Client {
	return &http.Client{
		Transport: &httpTransport{
			Error: err,
			Response: &http.Response{
				Body:       newResponseBody(body),
				StatusCode: code,
			},
		},
	}
}

func (r *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Cannot be more concise than this (i.e. `return r.Error, r.Response`) because
	// http.Client.Do warns if both Error and Response are non nil
	if r.Error != nil {
		return nil, r.Error
	}
	return r.Response, nil
}
