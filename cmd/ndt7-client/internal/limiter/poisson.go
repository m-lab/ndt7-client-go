package limiter

import (
	"math/rand"
	"sync"
	"time"
)

var once sync.Once

// This limiter can be used to implement an approximate Poisson
// process by generating inter-test wait times from an exponential
// distribution with lambda=1/Mean but clipped to the interval [Min,
// Max].
type PoissonLimiter struct {
	mean, min, max float64
}

func NewPoissonLimiter(mean, min, max float64) Limiter {
	return PoissonLimiter{mean, min, max}
}

func (l PoissonLimiter) Wait() {
	time.Sleep(l.sleepTime())
}

// Source:
// https://github.com/m-lab/ndt-server/blob/e30c1c9022417170e448ad5d1625abfc361a4301/spec/ndt7-protocol.md#requirements-for-non-interactive-clients
func (l PoissonLimiter) sleepTime() time.Duration{
	once.Do(func() {
		rand.Seed(time.Now().UTC().UnixNano())
	})
	t := rand.ExpFloat64() * l.mean
	if t < l.min {
		t = l.min
	} else if t > l.max {
		t = l.max
	}
	return time.Duration(t * float64(time.Second))
}
