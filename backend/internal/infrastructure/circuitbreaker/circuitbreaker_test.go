package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/paymentbridge/pcp/internal/infrastructure/circuitbreaker"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := circuitbreaker.New("test", 3, 2, time.Second)
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if cb.CurrentState() != circuitbreaker.StateClosed {
		t.Fatal("expected closed state")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := circuitbreaker.New("test", 3, 2, time.Second)
	testErr := errors.New("fail")
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error { return testErr })
	}
	if cb.CurrentState() != circuitbreaker.StateOpen {
		t.Fatal("expected open state after 3 failures")
	}
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Fatal("expected circuit open error")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := circuitbreaker.New("test", 2, 1, 50*time.Millisecond)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error { return errors.New("fail") })
	}
	if cb.CurrentState() != circuitbreaker.StateOpen {
		t.Fatal("expected open")
	}
	time.Sleep(60 * time.Millisecond)
	if cb.CurrentState() != circuitbreaker.StateHalfOpen {
		t.Fatal("expected half-open after timeout")
	}
}

func TestCircuitBreaker_ClosesAfterSuccess(t *testing.T) {
	cb := circuitbreaker.New("test", 2, 1, 50*time.Millisecond)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error { return errors.New("fail") })
	}
	time.Sleep(60 * time.Millisecond)
	_ = cb.Execute(func() error { return nil }) // success in half-open
	if cb.CurrentState() != circuitbreaker.StateClosed {
		t.Fatal("expected closed after success in half-open")
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	cb := circuitbreaker.New("test", 3, 2, time.Second)
	_ = cb.Execute(func() error { return errors.New("fail") })
	_ = cb.Execute(func() error { return errors.New("fail") })
	_ = cb.Execute(func() error { return nil }) // success resets count
	_ = cb.Execute(func() error { return errors.New("fail") })
	if cb.CurrentState() != circuitbreaker.StateClosed {
		t.Fatal("should still be closed, success reset the count")
	}
}
