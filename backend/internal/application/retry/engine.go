// Package retry implements retry strategies with exponential backoff and jitter.
package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Strategy defines retry behavior.
type Strategy struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	JitterFraction float64 // 0.0 to 1.0
}

// DefaultStrategy returns a sensible default retry strategy.
func DefaultStrategy() Strategy {
	return Strategy{
		MaxRetries:     3,
		InitialDelay:   500 * time.Millisecond,
		MaxDelay:       30 * time.Second,
		BackoffFactor:  2.0,
		JitterFraction: 0.2,
	}
}

// NextDelay calculates the delay for a given attempt number (0-based).
func (s Strategy) NextDelay(attempt int) time.Duration {
	delay := float64(s.InitialDelay) * math.Pow(s.BackoffFactor, float64(attempt))
	if delay > float64(s.MaxDelay) {
		delay = float64(s.MaxDelay)
	}
	// Add jitter
	jitter := delay * s.JitterFraction * (rand.Float64()*2 - 1)
	delay += jitter
	if delay < 0 {
		delay = float64(s.InitialDelay)
	}
	return time.Duration(delay)
}

// ShouldRetry returns true if another retry attempt should be made.
func (s Strategy) ShouldRetry(attempt int) bool {
	return attempt < s.MaxRetries
}

// Execute runs the given function with retry logic.
// Returns the result of the first successful call, or the last error.
func Execute(ctx context.Context, strategy Strategy, fn func(ctx context.Context) error) error {
	var lastErr error
	for attempt := 0; attempt <= strategy.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if attempt < strategy.MaxRetries {
			delay := strategy.NextDelay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return lastErr
}
