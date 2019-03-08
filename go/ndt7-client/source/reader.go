package source

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

func Reader(conn *websocket.Conn, input <-chan error) <-chan error {
	output := make(chan error)
	go func() {
		defer log.Debug("source.Reader: stop")
		defer close(output)
		defer func() {
			for range input {
				// Just drain the channel
			}
		}()
		log.Debug("source.Reader: start")
		for err := range input {
			if err != nil {
				output <- err
				return
			}
			kind, data, err := conn.ReadMessage()
			if err != nil {
				output <- err
				return
			}
			if kind == websocket.TextMessage {
				logmeasurement(data)
			}
		}
	}()
	return output
}
