// Package protocol implements the ndt7 client-server protocol
package protocol

import (
	"crypto/rand"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// newprepared creates a new random prepared message.
func newprepared() (*websocket.PreparedMessage, error) {
	const size = 1 << 13
	data := make([]byte, size)
	rand.Read(data)
	return websocket.NewPreparedMessage(websocket.BinaryMessage, data)
}

// Writer writes messages that load the network in a background
// goroutine. It returns a channel that tells you whether the
// writer encountered an error performing its task.
func Writer(conn *websocket.Conn) <-chan error {
	out := make(chan error)
	go func() {
		defer close(out)
		prepared, err := newprepared()
		if err != nil {
			out <- err
		}
		timer := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-timer.C:
				return
			default:
				err := conn.SetWriteDeadline(time.Now().Add(7 * time.Second))
				if err != nil {
					log.WithError(err).Warn("SetWriteDeadline failed")
					out <- err
					return
				}
				err = conn.WritePreparedMessage(prepared)
				if err != nil {
					log.WithError(err).Warn("WritePreparedMessage failed")
					out <- err
					return
				}
			}
		}
	}()
	return out
}
