package protocol

import (
	"encoding/json"
	"time"

	"github.com/apex/log"
)

type appinfo struct {
	NumBytes int64 `json:"num_bytes"`
}

type measurement struct {
	Elapsed float64 `json:"elapsed"`
	AppInfo appinfo `json:"app_info"`
}

// Measurer is a filter that runs in a goroutine. It fully drains the input
// channel that returns the read bytes. It returns a channel where it will
// periodically posts application level measurements ready to be sent by some
// other goroutine to the other party.
func Measurer(in <-chan int64) <-chan []byte {
	out := make(chan []byte)
	go func() {
		log.Debug("Measurer: start")
		defer log.Debug("Measurer: stop")
		defer func() {
			log.Debug("Measurer: reader detached; draining channel")
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
			data, err := json.Marshal(m)
			if err != nil {
				return
			}
			out <- data
		}
	}()
	return out
}
