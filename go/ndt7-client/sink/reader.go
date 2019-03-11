package sink

import (
	"strings"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// logmeasurement logs a measurement received from the server. We trim the
// trailing newlines such that the logs are more compact.
func logmeasurement(data []byte) {
	log.Infof("%s", strings.TrimRight(string(data), "\n"))
}

// ReadResult is the result emitted by Reader on its output channel.
type ReadResult struct {
	// Err is the error that may have occurred.
	Err error

	// Count is the number of read bytes.
	Count int64
}

// Reader reads messages from the websocket connection and posts the
// amount of read bytes on the output channel. If an error occurs,
// the error is posted on the output channel and we terminate. Also,
// we log all the received measurement messages.
func Reader(conn *websocket.Conn) <-chan ReadResult {
	output := make(chan ReadResult)
	go func() {
		conn.SetCloseHandler(func(int, string) error {
			log.Debug("sink.Reader: got close message; deferring response")
			return nil
		})
		log.Debug("sink.Reader: start")
		defer log.Debug("sink.Reader: stop")
		defer close(output)
		for {
			kind, data, err := conn.ReadMessage()
			if err != nil {
				output <- ReadResult{Err: err}
				return
			}
			if kind == websocket.TextMessage {
				logmeasurement(data)
			}
			output <- ReadResult{Count: int64(len(data))}
		}
	}()
	return output
}
