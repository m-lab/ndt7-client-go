// Package limiter contains the ndt7-client rate limiter.
package limiter

// Limiter is used to rate limit tests from non-interactive clients.
//
// Clients are expected to call Wait() between tests. See
// PoissonLimiter for a specific implementation.
type Limiter interface {
	// Block for a certain amount of time.
	Wait()
}
