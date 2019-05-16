// Package upload contains ndt7 upload code
package upload

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
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

// ignoreIncoming ignores any incoming message. The error is typically ignored
// as this code runs in its own goroutine, yet it's useful for testing.
func ignoreIncoming(conn websocketx.Conn) error {
	conn.SetReadLimit(params.MaxMessageSize)
	for {
		// Implementation note: this guarantees that the websocket engine
		// is processing messages. Here we're using as timeout the timeout
		// for the whole upload, so that we know that this goroutine is
		// active for most of the time we care about, even in the case in
		// which the server is not sending us any messages.
		err := conn.SetReadDeadline(time.Now().Add(params.UploadTimeout))
		if err != nil {
			return err
		}
		mtype, _, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if mtype != websocket.TextMessage {
			return errNonTextMessage
		}
	}
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
// the amount of bytes written on the provided chan. The returned
// error is mainly useful for testing, as this code is meant to run
// in its own goroutine setup by the caller.
//
// Note that upload closes the out channel.
func upload(ctx context.Context, conn websocketx.Conn, out chan<- int64) error {
	defer close(out)
	wholectx, cancel := context.WithTimeout(ctx, params.UploadTimeout)
	defer cancel()
	preparedMessage, err := makePreparedMessage(params.BulkMessageSize)
	if err != nil {
		return err
	}
	var total int64
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
		out <- total
	}
	return nil
}

// uploadAsync runs the upload and returns a channel where progress is
// emitted. The channel will be close when done.
func uploadAsync(ctx context.Context, conn websocketx.Conn) <-chan int64 {
	out := make(chan int64)
	go upload(ctx, conn, out)
	return out
}

// Run runs the upload subtest. It runs until the ctx is expired or the
// upload timeout is expired. It uses the provided conn. It emits on the
// provided channel upload measurements. The returned error is mainly
// useful for making this function have the same API of download.Run, for
// which it makes more sense to return an error.
//
// Note that run closes both ch and conn.
func Run(ctx context.Context, conn websocketx.Conn, ch chan<- spec.Measurement) error {
	defer close(ch)
	defer conn.Close()
	go ignoreIncoming(conn)
	start := time.Now()
	prev := start
	for tot := range uploadAsync(ctx, conn) {
		now := time.Now()
		if now.Sub(prev) > params.UpdateInterval {
			emit(ch, now.Sub(start).Seconds(), tot)
			prev = now
		}
	}
	return nil
}
