package protocol

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
// on any error, the reader will close the returned channel. A side
// effect of Reader is that it disables automatically sending the
// websocket.CloseMessage to the other party. This is because sending
// such message when the other party sends a CloseMessage to us is
// in general too early and we should do that when we're done.
func Reader(conn *websocket.Conn) <-chan int64 {
	conn.SetCloseHandler(func(code int, message string) error {
		log.Debugf("Reader: got close %s: %d", message, code)
		log.Debugf("Reader: we are not replying now because it's too early")
		return nil
	})
	const timeout = 7 * time.Second
	out := make(chan int64)
	go func() {
		log.Debug("Reader: start")
		defer log.Debug("Reader: stop")
		defer close(out)
		for {
			err := conn.SetReadDeadline(time.Now().Add(timeout))
			if err != nil {
				log.WithError(err).Warn("SetReadDeadline failed")
				return
			}
			kind, data, err := conn.ReadMessage()
			if err != nil {
				log.WithError(err).Warn("ReadMessage failed")
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
