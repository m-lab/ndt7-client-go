// Package websocketx contains websocket extensions.
package websocketx

import (
	"time"

	"github.com/gorilla/websocket"
)

// Conn is the interface of a websocket.Conn used for mocking.
type Conn interface {
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	SetReadLimit(limit int64)
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	WritePreparedMessage(pm *websocket.PreparedMessage) error
}
