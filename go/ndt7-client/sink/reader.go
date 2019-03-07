package sink

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

// Reader reads messages from the websocket connection in a background
// goroutine. The length of messages will be posted on the returned
// channel. Additionally, measurement messages will be logged. In case
// on any error, the reader will close the returned channel.
func Reader(conn *websocket.Conn) <-chan int64 {
	const timeout = 7 * time.Second
	out := make(chan int64)
	go func() {
		log.Debug("sink.Reader: start")
		defer log.Debug("sink.Reader: stop")
		defer close(out)
		for {
			err := conn.SetReadDeadline(time.Now().Add(timeout))
			if err != nil {
				log.WithError(err).Warn("SetReadDeadline failed")
				return
			}
			kind, data, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.WithError(err).Warn("ReadMessage failed")
				}
				return
			}
			if kind == websocket.TextMessage {
				logmeasurement(data)
			}
			out <- int64(len(data))
		}
	}()
	return out
}
