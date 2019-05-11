// Package download contains ndt7 download code.
package download

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/internal/websocketx"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Run runs the download subtest. It runs until the ctx expires or the maximum
// download time expires. Uses the provided websocket connection. Emits zero
// or more measurements to the provided channel. Returns the error that caused
// the download loop to stop, which is mainly useful when testing, since the
// normal usage of this function is to be run in a separate goroutine.
//
// Note that this function closes conn and ch when exiting.
func Run(ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement) error {
	defer close(ch)
	defer conn.Close()
	wholectx, cancel := context.WithTimeout(ctx, params.DownloadTimeout)
	defer cancel()
	conn.SetReadLimit(params.MaxMessageSize)
	for wholectx.Err() == nil {
		err := conn.SetReadDeadline(time.Now().Add(params.IOTimeout))
		if err != nil {
			return err
		}
		mtype, mdata, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if mtype != websocket.TextMessage {
			continue
		}
		var measurement spec.Measurement
		err = json.Unmarshal(mdata, &measurement)
		if err != nil {
			return err
		}
		measurement.Direction = spec.DirectionDownload
		measurement.Origin = spec.OriginServer
		ch <- measurement
	}
	return nil // this is how success looks like
}
