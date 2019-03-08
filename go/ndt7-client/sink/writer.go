package sink

import (
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
				output <- mr.Err
				return
			}
			err := conn.WriteMessage(websocket.TextMessage, mr.Measurement)
			if err != nil {
				log.WithError(err).Warn("WriteMessage failed")
				output <- err
				return
			}
		}
	}()
	return output
}
