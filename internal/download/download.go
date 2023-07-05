// Package download contains ndt7 download code.
package download

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/internal/websocketx"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Run runs the download test. It runs until the ctx expires or the maximum
// download time expires. Uses the provided websocket connection. Emits zero
// or more measurements to the provided channel. Returns the error that caused
// the download loop to stop, which is mainly useful when testing, since the
// normal usage of this function is to be run in a separate goroutine. Note
// that this function would block if you don't read from the channel.
//
// Note that this function closes conn and ch when exiting.
func Run(ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement) error {
	defer close(ch)
	defer conn.Close()
	wholectx, cancel := context.WithTimeout(ctx, params.DownloadTimeout)
	defer cancel()
	conn.SetReadLimit(params.MaxMessageSize)
	start := time.Now()
	prev := start
	sent := false
	var total int64
	for wholectx.Err() == nil {
		err := conn.SetReadDeadline(time.Now().Add(params.IOTimeout))
		if err != nil {
			return err
		}
		mtype, r, err := conn.NextReader()
		if err != nil {
			// If the test finishes before any measurements have been sent through the channel,
			// and at least one message has been received, send the measurement data before exiting.
			if !sent && total > 0 {
				sendMeasurement(ch, time.Now().Sub(start), total)
			}
			return err
		}
		var msgSize int64
		var mdata []byte
		if mtype == websocket.BinaryMessage {
			// Binary messages are discarded, since we are only interested
			// in their size.
			msgSize, err = io.Copy(ioutil.Discard, r)
			if err != nil {
				return err
			}
		} else {
			mdata, err = ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			msgSize = int64(len(mdata))
		}
		total += msgSize
		now := time.Now()
		if now.Sub(prev) > params.UpdateInterval {
			prev = now
			elapsed := now.Sub(start)
			sendMeasurement(ch, elapsed, total)
			sent = true
			// FALLTHROUGH
		}
		if mtype != websocket.TextMessage {
			continue
		}
		var measurement spec.Measurement
		err = json.Unmarshal(mdata, &measurement)
		if err != nil {
			return err
		}
		measurement.Origin = spec.OriginServer
		measurement.Test = spec.TestDownload
		ch <- measurement
	}
	return nil // this is how success looks like
}

func sendMeasurement(ch chan<- spec.Measurement, elapsed time.Duration, total int64) {
	ch <- spec.Measurement{
		AppInfo: &spec.AppInfo{
			ElapsedTime: int64(elapsed) / int64(time.Microsecond),
			NumBytes:    total,
		},
		Origin: spec.OriginClient,
		Test:   spec.TestDownload,
	}
}
