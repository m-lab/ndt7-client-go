// Package sink implements a ndt7 sink. This is the role of the client
// during the download and of the server during the upload.
package sink

import (
	"encoding/json"
	"time"

	"github.com/apex/log"
)

// appinfo contains an application level measurement. It is part of
// the ndt7 specification.
type appinfo struct {
	// NumBytes indicates the number of bytes received so far.
	NumBytes int64 `json:"num_bytes"`

	// Speed contains the speed in Mbit/s. This is not part of the spec
	// but is useful to understand how fast we're going here.
	Speed float64 `json:"speed"`
}

// measurement is a measurement message compliant with the ndt7 spec.
type measurement struct {
	// Elapsed is the number of seconds since the beginning.
	Elapsed float64 `json:"elapsed"`

	// AppInfo contains application level measurements.
	AppInfo appinfo `json:"app_info"`
}

// MeasureResult is one of the results emitted by Measurer
// on the channel that it returns.
type MeasureResult struct {
	// Err is the error that may have occurred.
	Err error

	// Measurement is a serialized measurement to be sent to
	// the peer using a counterflow message.
	Measurement []byte
}

// Measurer aggregates read results coming from the input channel and
// periodically emits serialized measurements to be uploaded on the
// output channel. If an error comes in the input channel, we will pass
// the error on the output channel and leave.
func Measurer(input <-chan ReadResult) <-chan MeasureResult {
	output := make(chan MeasureResult)
	go func() {
		defer log.Debug("sink.Measurer: stop")
		defer func() {
			for range input {
				// Just drain the channel
			}
		}()
		defer close(output)
		log.Debug("sink.Measurer: start")
		const interval = 250 * time.Millisecond // min counter-flow interval
		var total int64
		prev := time.Now()
		begin := prev
		for rr := range input {
			if rr.Err != nil {
				output <- MeasureResult{Err: rr.Err}
				return
			}
			t := time.Now()
			total += rr.Count // int64 overflow is unlikely here
			if t.Sub(prev) < interval {
				continue
			}
			prev = t
			var m measurement
			m.Elapsed = float64(t.Sub(begin)) / float64(time.Second)
			m.AppInfo.NumBytes = total
			m.AppInfo.Speed = 8.0 * float64(total) / m.Elapsed / 1000.0 / 1000.0
			data, err := json.Marshal(m)
			if err != nil {
				output <- MeasureResult{Err: err}
				return
			}
			output <- MeasureResult{Measurement: data}
		}
	}()
	return output
}
