package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a per-key sliding window rate limiter.
// Uses in-memory storage; swap to Redis for distributed deployments.
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rate     int           // max requests
	window   time.Duration // per time window
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter allowing 'rate' requests per 'window'.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(window * 2)
			rl.cleanup()
		}
	}()
	return rl
}

// RateLimit returns middleware that rate-limits by client IP or API key.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.RemoteAddr
		}

		if !rl.allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "rate_limit_exceeded", "message": "too many requests", "code": 429,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists || time.Since(v.lastReset) > rl.window {
		rl.visitors[key] = &visitor{tokens: rl.rate - 1, lastReset: time.Now()}
		return true
	}
	if v.tokens <= 0 {
		return false
	}
	v.tokens--
	return true
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for key, v := range rl.visitors {
		if time.Since(v.lastReset) > rl.window*3 {
			delete(rl.visitors, key)
		}
	}
}
