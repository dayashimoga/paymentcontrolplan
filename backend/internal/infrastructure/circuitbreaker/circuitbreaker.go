// Package circuitbreaker implements the circuit breaker resilience pattern.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // Normal operation, requests pass through
	StateOpen                  // Failure threshold exceeded, requests blocked
	StateHalfOpen              // Testing if service recovered
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	}
	return "unknown"
}

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker prevents cascading failures by tracking error rates.
type CircuitBreaker struct {
	mu               sync.RWMutex
	name             string
	state            State
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int // consecutive successes needed in half-open to close
	timeout          time.Duration
	lastFailureTime  time.Time
}

// New creates a new circuit breaker.
// failureThreshold: number of failures before opening the circuit.
// timeout: how long the circuit stays open before transitioning to half-open.
// successThreshold: number of consecutive successes in half-open needed to close.
func New(name string, failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Execute runs the given function through the circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		return StateHalfOpen
	}
	return cb.state
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
		}
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	if cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// Name returns the circuit breaker name.
func (cb *CircuitBreaker) Name() string { return cb.name }

// Counts returns failure and success counts for monitoring.
func (cb *CircuitBreaker) Counts() (failures, successes int) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount, cb.successCount
}
