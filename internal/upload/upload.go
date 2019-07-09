// Package upload contains ndt7 upload code
package upload

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt7-client-go/internal/merger"
	"github.com/m-lab/ndt7-client-go/internal/params"
	"github.com/m-lab/ndt7-client-go/internal/websocketx"
	"github.com/m-lab/ndt7-client-go/spec"
)

// makePreparedMessage generates a prepared message that should be sent
// over the network for generating network load.
var makePreparedMessage = func(size int) (*websocket.PreparedMessage, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	data := make([]byte, size)
	// This is not the fastest algorithm to generate a random string, yet it
	// is most likely good enough for our purposes. See [1] for a comprehensive
	// discussion regarding how to generate a random string in Golang.
	//
	// .. [1] https://stackoverflow.com/a/31832326/4354461
	//
	// Also, the ndt7 specification does not require us to use this algorithm
	// and we could send purely random data as well. We're sending textual data
	// here just because in some debugging cases it's easier to read.
	for i := range data {
		data[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return websocket.NewPreparedMessage(websocket.BinaryMessage, data)
}

// errNonTextMessage indicates we've got a non textual message
var errNonTextMessage = errors.New("Received non textual message")

func counterflowReader(
	ctx context.Context, conn websocketx.Conn, out chan<- spec.Measurement,
) error {
	defer close(out)
	wholectx, cancel := context.WithTimeout(ctx, params.UploadTimeout)
	defer cancel()
	conn.SetReadLimit(params.MaxMessageSize)
	for wholectx.Err() == nil {
		// Implementation note: this guarantees that the websocket engine
		// is processing messages. Here we're using as timeout the timeout
		// for the whole upload, so that we know that this goroutine is
		// active for most of the time we care about, even in the case in
		// which the server is not sending us any messages.
		err := conn.SetReadDeadline(time.Now().Add(params.UploadTimeout))
		if err != nil {
			return err
		}
		mtype, mdata, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if mtype != websocket.TextMessage {
			return errNonTextMessage
		}
		var m spec.Measurement
		err = json.Unmarshal(mdata, &m)
		if err != nil {
			return err
		}
		m.Origin = spec.OriginServer
		m.Direction = spec.DirectionUpload
		out <- m
	}
	return nil
}

// startCounterflowReader starts the goroutine that is reading
// the incoming counterflow messages.
func startCounterflowReader(
	ctx context.Context, conn websocketx.Conn,
) <-chan spec.Measurement {
	out := make(chan spec.Measurement)
	go counterflowReader(ctx, conn, out)
	return out
}

// emit emits an event during the upload.
func emit(ch chan<- spec.Measurement, elapsed float64, numBytes int64) {
	ch <- spec.Measurement{
		AppInfo: spec.AppInfo{
			NumBytes: numBytes,
		},
		Direction: spec.DirectionUpload,
		Elapsed:   elapsed,
		Origin:    spec.OriginClient,
	}
}

// upload runs the upload until the context is done or the upload
// timeout expires. It uses the provided websocket conn. It wil emit
// the measurements on the provided chan. The returned error is
// mainly useful for testing, as this code is meant to run in its
// own goroutine setup by the caller.
//
// Note that upload closes the out channel.
func upload(
	ctx context.Context, conn websocketx.Conn,
	out chan<- spec.Measurement,
) error {
	defer close(out)
	wholectx, cancel := context.WithTimeout(ctx, params.UploadTimeout)
	defer cancel()
	preparedMessage, err := makePreparedMessage(params.BulkMessageSize)
	if err != nil {
		return err
	}
	var total int64
	start := time.Now()
	prev := start
	for wholectx.Err() == nil {
		err := conn.SetWriteDeadline(time.Now().Add(params.IOTimeout))
		if err != nil {
			return err
		}
		if err := conn.WritePreparedMessage(preparedMessage); err != nil {
			return err
		}
		// Note that the following is slightly inaccurate because we
		// are ignoring the WebSocket overhead et al.
		total += params.BulkMessageSize
		now := time.Now()
		if now.Sub(prev) > params.UpdateInterval {
			emit(out, now.Sub(start).Seconds(), total)
			prev = now
		}
	}
	return nil
}

// startUpload starts the upload goroutine and returns a channel where progress
// is emitted. The channel will be close when done.
func startUpload(
	ctx context.Context, conn websocketx.Conn,
) <-chan spec.Measurement {
	out := make(chan spec.Measurement)
	go upload(ctx, conn, out)
	return out
}

// closeconn closes the WebSocket connection and returns the error.
func closeconn(conn websocketx.Conn) error {
	return conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Done sending"),
		time.Now().Add(params.IOTimeout),
	)
}

// Run runs the upload subtest. It runs until the ctx is expired or the
// upload timeout is expired. It uses the provided conn. It emits on the
// provided channel upload measurements. The returned error is mainly
// useful for making this function have the same API of download.Run, for
// which it makes more sense to return an error.
//
// Note that run closes both ch and conn.
func Run(
	ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement,
) error {
	defer close(ch)
	defer conn.Close()
	serverch := startCounterflowReader(ctx, conn)
	clientch := startUpload(ctx, conn)
	for m := range merger.Merge(serverch, clientch) {
		ch <- m
	}
	closeconn(conn) // ignoring return value on purpose
	return nil
}
