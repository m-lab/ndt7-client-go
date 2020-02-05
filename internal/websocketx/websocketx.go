// Package websocketx contains websocket extensions.
package websocketx

import (
	"io"
	"time"

	"github.com/gorilla/websocket"
)

// Conn is the interface of a websocket.Conn used for mocking.
type Conn interface {
	Close() error
	NextReader() (messageType int, reader io.Reader, err error)
	SetReadLimit(limit int64)
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	WritePreparedMessage(pm *websocket.PreparedMessage) error
}
