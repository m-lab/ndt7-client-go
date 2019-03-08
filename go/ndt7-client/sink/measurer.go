package sink

import (
	"encoding/json"
	"time"

	"github.com/apex/log"
)

type appinfo struct {
	NumBytes int64 `json:"num_bytes"`
	// Speed is not part of the spec but it helps to see it on the
	// wire when we're not using BBR-enabled ndt7
	Speed float64 `json:"speed"`
}

type measurement struct {
	Elapsed float64 `json:"elapsed"`
	AppInfo appinfo `json:"app_info"`
}

type MeasureResult struct {
	Err error
	Measurement []byte
}

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
		const interval = 250 * time.Millisecond
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
