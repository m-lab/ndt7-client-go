package source

import (
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// logmeasurement logs a measurement received from the server. We trim the
// trailing newlines such that the logs are more compact.
func logmeasurement(data []byte) {
	log.Infof("%s", strings.TrimRight(string(data), "\n"))
}

// Reader runs in a background goroutine and processes all incoming
// websocket messages. Text messages are logged. The Reader will
// post any error that may occur on the returned channel. It will
// then close the channel. It will close the channel without any
// error being posted in case of normal websocket closure.
func Reader(conn *websocket.Conn) <-chan error {
	const timeout = 7 * time.Second
	out := make(chan error)
	go func() {
		log.Debug("source.Reader: start")
		defer log.Debug("source.Reader: stop")
		defer close(out)
		for {
			err := conn.SetReadDeadline(time.Now().Add(timeout))
			if err != nil {
				out <- err
				return
			}
			kind, data, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					out <- err
				}
				return
			}
			if kind == websocket.TextMessage {
				logmeasurement(data)
			}
		}
	}()
	return out
}
