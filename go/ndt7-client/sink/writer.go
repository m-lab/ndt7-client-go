package sink

import (
	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// Writer is a reducer that receives measurements ready to be sent
// to the other party and sends them using the connection.
func Writer(conn *websocket.Conn, in <-chan []byte) error {
	log.Debug("sink.Writer: start")
	defer log.Debug("sink.Writer: stop")
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
	return nil
}
