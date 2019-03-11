package sink

import (
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// Writer reads measurement results from the input channel and submits
// them to the peer using the websocket connection. In case of any error,
// this function will post in on the returned channel and terminate.
func Writer(conn *websocket.Conn, input <-chan MeasureResult) <-chan error {
	output := make(chan error)
	go func() {
		defer close(output)
		defer log.Debug("sink.Writer: stop")
		defer func() {
			for range input {
				// Just drain the channel
			}
		}()
		log.Debug("sink.Writer: start")
		for mr := range input {
			if mr.Err != nil {
				if !websocket.IsCloseError(mr.Err, websocket.CloseNormalClosure) {
					output <- mr.Err
					return
				}
				break
			}
			err := conn.WriteMessage(websocket.TextMessage, mr.Measurement)
			if err != nil {
				log.WithError(err).Warn("WriteMessage failed")
				output <- err
				return
			}
		}
		log.Debug("sink.Writer: sending CLOSE message to peer")
		msg := websocket.FormatCloseMessage(
			websocket.CloseNormalClosure, "Subtest complete")
		output <- conn.WriteControl(websocket.CloseMessage, msg, time.Time{})
	}()
	return output
}
