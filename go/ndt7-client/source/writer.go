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

// Writer writes messages that load the network in a background
// goroutine. It also keeps an eye on an input channel where
// the Reader may post any error. The writer will return early
// with an error if the Reader reports an error or received
// a clean CloseMessage. Otherwise, it will continue writing
// until the test duration has expired. At that point, it will
// send a clean CloseMessage to indicate it is done.
func Writer(conn *websocket.Conn, in <-chan error) <-chan error {
	output := make(chan error)
	go func() {
		defer close(output)
		defer func() {
			for range in {
				// drain
			}
		}()
		defer log.Debug("source.Writer: stop")
		log.Debug("source.Writer: start")
		data := makedata()
		timer := time.NewTimer(maxSendTime)
		for {
			select {
			case err := <-in:
				output <- err // Forward error coming from the reader
				return
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
			}
		}
	}()
	return output
}
