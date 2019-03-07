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

// Measurer is a filter that runs in a goroutine. It fully drains the input
// channel that returns the read bytes. It returns a channel where it will
// periodically posts application level measurements ready to be sent to the
// other party by the consumer goroutine reading the channel.
func Measurer(in <-chan int64) <-chan []byte {
	out := make(chan []byte)
	go func() {
		log.Debug("sink.Measurer: start")
		defer log.Debug("sink.Measurer: stop")
		defer func() {
			log.Debug("sink.Measurer: draining reader's channel")
			for range in {
				// Just drain the channel
			}
		}()
		defer close(out)
		const interval = 250 * time.Millisecond
		var total int64
		prev := time.Now()
		begin := prev
		for n := range in {
			t := time.Now()
			total += n // int64 overflow is unlikely here
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
				return
			}
			out <- data
		}
	}()
	return out
}
