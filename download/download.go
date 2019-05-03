// Package download contains ndt7 download code.
package download

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Run runs the download subtest.
func Run(ctx context.Context, conn *websocket.Conn, ch chan<- spec.Measurement) {
	defer close(ch)
	defer conn.Close()
	wholectx, cancel := context.WithTimeout(ctx, spec.DownloadTimeout)
	defer cancel()
	conn.SetReadLimit(spec.MaxMessageSize)
	for {
		select {
		case <-wholectx.Done():
			return // don't fail the test if we're running for too much time
		default:
			// nothing
		}
		conn.SetReadDeadline(time.Now().Add(spec.IOTimeout))
		mtype, mdata, err := conn.ReadMessage()
		if err != nil {
			return // don't fail the test because of an I/O error
		}
		if mtype != websocket.TextMessage {
			continue
		}
		var measurement spec.Measurement
		err = json.Unmarshal(mdata, &measurement)
		if err != nil {
			return // fail the test if we got an invalid JSON
		}
		measurement.Direction = spec.DirectionDownload
		measurement.Origin = spec.OriginServer
		ch <- measurement
	}
}
