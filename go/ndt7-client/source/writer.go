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

// Writer writes binary messages on the connection for ten seconds, then
// sends a close message and terminates. It will also send on the returned
// channel connection status updates. Those will either be `nil` if all
// is good, or a specififc error. After any I/O error the Writer terminates.
//
// Because the Writer must go as fast as possible, any connection status
// update that is not an error must be sent in nonblocking fashion so that
// we can resume writing ASAP. It does not matter to lose one of these
// updates, because it means that the reader is waiting for messages and
// not reading the channel. So this design is fine. The reason why the
// Reader needs to be after the Writer in the pipeline is explained in the
// documentation of the Reader, so we'll not repeat ourselves here.
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
			}
		}
	}()
	return output
}
