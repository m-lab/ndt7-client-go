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

// Reader is the first stage of the source pipeline. It runs in a background
// goroutine and returns a channel. It will drain messages from the websocket
// connection, discard binary messages, and log text messages. When there is
// a read error (which includes receiving a CloseMessage control message on
// the websocket connection), the reader goroutine will post such error on the
// channel, close the channel, and then terminate.
func Reader(conn *websocket.Conn) <-chan error {
	output := make(chan error)
	go func() {
		log.Debug("source.Reader: start")
		defer log.Debug("source.Reader: stop")
		defer close(output)
		for {
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
