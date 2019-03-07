package source

import (
	"crypto/rand"
	"errors"
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
// goroutine. It also keeps an eye on an input channel where
// the Reader may post any error. The writer will return early
// with an error if the Reader reports an error or received
// a clean CloseMessage. Otherwise, it will continue writing
// until the test duration has expired. At that point, it will
// send a clean CloseMessage to indicate it is done.
func Writer(conn *websocket.Conn, in <-chan error) error {
	defer func() {
		log.Debug("source.Writer: draining reader output channel")
		for range in {
			// drain
		}
	}()
	log.Debug("source.Writer: start")
	defer log.Debug("source.Writer: stop")
	prepared, err := newprepared()
	if err != nil {
		return err
	}
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case err := <-in:
			if err == nil {
				err = errors.New("source.Writer: the reader disconnected early")
			}
			return err
		case <-timer.C:
			log.Debug("source.Writer: closing the connection cleanly")
			msg := websocket.FormatCloseMessage(
				websocket.CloseNormalClosure, "Measurement complete")
			deadline := time.Now().Add(7 * time.Second)
			return conn.WriteControl(websocket.CloseMessage, msg, deadline)
		default:
			err := conn.SetWriteDeadline(time.Now().Add(7 * time.Second))
			if err != nil {
				log.WithError(err).Warn("SetWriteDeadline failed")
				return err
			}
			err = conn.WritePreparedMessage(prepared)
			if err != nil {
				log.WithError(err).Warn("WritePreparedMessage failed")
				return err
			}
		}
	}
}
