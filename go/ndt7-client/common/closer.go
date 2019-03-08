package common

import (
	"errors"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// ErrTimeout indicates that this subtest has timed out.
var ErrTimeout = errors.New("Timeout while sending data to peer")

// maxDuration is the time after which we timeout.
const maxDuration = 15 * time.Second

// Closer should always be the last stage of a subtest pipeline. It will
// make sure that we stop after fifteen seconds. It will report to the
// caller the error that occurred, if any. In case no error occurred, the
// subtest completed successfully. Otherwise, it may have timed out or
// terminated earlier than expected because of some network error.
func Closer(conn *websocket.Conn, in <-chan error) error {
	defer func() {
		for range in {
			// drain
		}
	}()
	defer log.Debug("common.Closer: stop")
	log.Debug("common.Closer: start")
	timer := time.NewTimer(maxDuration)
	var err error
	select {
	case <-timer.C:
		err = ErrTimeout
	case err = <-in:
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			err = nil
		}
	}
	if err != nil {
		err2 := conn.Close()
		if err2 != nil {
			log.WithError(err2).Debug("Ignoring conn.Close failure")
		}
		return err
	}
	return conn.Close()
}
