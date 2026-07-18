package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// IdempotencyStore provides in-memory idempotency key storage.
// Swap to Redis for distributed deployments.
type IdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]*idempotencyEntry
	ttl     time.Duration
}

type idempotencyEntry struct {
	statusCode int
	body       []byte
	headers    http.Header
	createdAt  time.Time
}

// NewIdempotencyStore creates an idempotency store with the given TTL.
func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	s := &IdempotencyStore{entries: make(map[string]*idempotencyEntry), ttl: ttl}
	go func() {
		for {
			time.Sleep(ttl)
			s.cleanup()
		}
	}()
	return s
}

// Idempotency middleware returns cached responses for duplicate idempotency keys.
// Only applies to POST/PUT/PATCH methods. Uses the Idempotency-Key header.
func (s *IdempotencyStore) Idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Hash the key for consistent storage
		hash := sha256.Sum256([]byte(key))
		storeKey := hex.EncodeToString(hash[:])

		// Check for cached response
		s.mu.RLock()
		entry, exists := s.entries[storeKey]
		s.mu.RUnlock()

		if exists && time.Since(entry.createdAt) < s.ttl {
			for k, v := range entry.headers {
				w.Header()[k] = v
			}
			w.Header().Set("X-Idempotent-Replayed", "true")
			w.WriteHeader(entry.statusCode)
			w.Write(entry.body)
			return
		}

		// Capture the response
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Cache successful responses
		if rec.statusCode >= 200 && rec.statusCode < 300 {
			s.mu.Lock()
			s.entries[storeKey] = &idempotencyEntry{
				statusCode: rec.statusCode,
				body:       rec.body,
				headers:    w.Header().Clone(),
				createdAt:  time.Now(),
			}
			s.mu.Unlock()
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
	written    bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func (s *IdempotencyStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.entries {
		if time.Since(v.createdAt) > s.ttl {
			delete(s.entries, k)
		}
	}
}

// writeJSON is a helper for middleware JSON responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
