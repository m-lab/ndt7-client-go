// Package merger contains code for merging channels.
package merger

import (
	"sync"

	"github.com/m-lab/ndt7-client-go/spec"
)

// Merge merges the content of two channels into the returned channel.
func Merge(leftch, rightch <-chan spec.Measurement) <-chan spec.Measurement {
	// Implementation note: the following is the well known golang
	// pattern for joining channels together
	outch := make(chan spec.Measurement)
	var wg sync.WaitGroup
	wg.Add(2)
	// leftch; note that it MUST provide a liveness guarantee
	go func(out chan<- spec.Measurement) {
		for m := range leftch {
			out <- m
		}
		wg.Done()
	}(outch)
	// rightch; note that it MUST provide a liveness guarantee
	go func(out chan<- spec.Measurement) {
		for m := range rightch {
			out <- m
		}
		wg.Done()
	}(outch)
	// closer; will always terminate because of above liveness guarantees
	go func() {
		wg.Wait()
		close(outch)
	}()
	return outch
}
