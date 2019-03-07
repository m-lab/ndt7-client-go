package sink

import (
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// Writer is a reducer that receives measurements ready to be sent
// to the other party and sends them using the connection. When it is
// done, Writer will send a websocket.CloseMessage message so that
// it is clear we are done with sending counterflow measurements.
func Writer(conn *websocket.Conn, in <-chan []byte) error {
	defer func() {
		for range in {
			// make sure we drain the channel
		}
	}()
	for data := range in {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.WithError(err).Warn("WriteMessage failed")
			return err
		}
	}
	msg := websocket.FormatCloseMessage(
		websocket.CloseNormalClosure, "Done with sending counterflow measurements")
	deadline := time.Now().Add(3 * time.Second)
	return conn.WriteControl(websocket.CloseMessage, msg, deadline)
}
