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

type ReadResult struct {
	Err error

	Count int64
}

// Reader reads messages from the websocket connection in a background
// goroutine. The length of messages will be posted on the returned
// channel. Additionally, measurement messages will be logged. In case
// on any error, the reader will close the returned channel.
func Reader(conn *websocket.Conn) <-chan ReadResult {
	output := make(chan ReadResult)
	go func() {
		log.Debug("sink.Reader: start")
		defer log.Debug("sink.Reader: stop")
		defer close(output)
		conn.SetCloseHandler(func (int, string) error {
			log.Debug("sink.Reader: got close message; deferring response")
			return nil
		})
		for {
			kind, data, err := conn.ReadMessage()
			if err != nil {
				log.WithError(err).Debug("sink.Reader error")
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
