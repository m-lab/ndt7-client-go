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

// Reader reads the connection status from the input channel. In case we
// see an error, we will immediately post it on the output channel and
// leave. Otherwise, it will read the next counter-flow message carrying
// a measurement when activated. If we fail to read, we drain the input
// channel and then return the error to the caller.
//
// Note that Reader is not performance critical because we expect to
// receive just a few counter-flow measurements every second. As such,
// it's fine for it to be periodically activated by the Writer that,
// instead, must go as fast as possible. I tried initially to put the
// Reader before the writer in the pipeline, but that led to a race
// because the shutdown sequence is that the source.Writer sends a
// CloseMessage and leaves, then the source.Reader needs to wait for
// the final counter-flow measurements followed by the peer sending
// us its close message. As such, the Reader needs to come after
// the Writer otherwise we need to play with time.Sleep and in the
// general case we will always make some clients unhappy.
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
		for {
			select {
			case err, ok := <-input:
				if ok && err != nil {
					output <- err
					return
				}
			default:
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
