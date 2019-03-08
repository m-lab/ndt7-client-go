package source

import (
	"crypto/rand"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// datasize is the size of the messages we send to load the network
const datasize = 1 << 13

// makedata fills a byte slice with random data
func makedata() []byte {
	const size = datasize
	data := make([]byte, size)
	rand.Read(data)
	return data
}

// maxSendTime is the max time for which we are supposed to send
const maxSendTime = 10 * time.Second

func Writer(conn *websocket.Conn) <-chan error {
	output := make(chan error)
	go func() {
		defer close(output)
		defer log.Debug("source.Writer: stop")
		log.Debug("source.Writer: start")
		data := makedata()
		timer := time.NewTimer(maxSendTime)
		for {
			select {
			case <-timer.C:
				log.Debug("source.Writer: sending CloseMessage")
				msg := websocket.FormatCloseMessage(
					websocket.CloseNormalClosure, "Subtest complete")
				output <- conn.WriteControl(websocket.CloseMessage, msg, time.Time{})
				return // Finished running subtest
			default:
				err := conn.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					output <- err // Forward I/O error when writing
					return
				}
				select {
				case output <- nil:
				default:
				}
			}
		}
	}()
	return output
}
