package retry_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/paymentbridge/pcp/internal/application/retry"
)

func TestDefaultStrategy_NextDelay(t *testing.T) {
	s := retry.DefaultStrategy()
	d0 := s.NextDelay(0)
	d1 := s.NextDelay(1)
	if d1 < d0 {
		t.Fatal("delay should increase with attempts")
	}
}

func TestDefaultStrategy_ShouldRetry(t *testing.T) {
	s := retry.DefaultStrategy()
	if !s.ShouldRetry(0) {
		t.Fatal("should retry at attempt 0")
	}
	if !s.ShouldRetry(2) {
		t.Fatal("should retry at attempt 2")
	}
	if s.ShouldRetry(3) {
		t.Fatal("should not retry at attempt 3 (max=3)")
	}
}

func TestExecute_Success(t *testing.T) {
	s := retry.Strategy{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, BackoffFactor: 2}
	calls := 0
	err := retry.Execute(context.Background(), s, func(ctx context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestExecute_EventualSuccess(t *testing.T) {
	s := retry.Strategy{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, BackoffFactor: 2}
	calls := 0
	err := retry.Execute(context.Background(), s, func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return fmt.Errorf("temporary error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_AllFail(t *testing.T) {
	s := retry.Strategy{MaxRetries: 2, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, BackoffFactor: 2}
	err := retry.Execute(context.Background(), s, func(ctx context.Context) error {
		return fmt.Errorf("permanent error")
	})
	if err == nil {
		t.Fatal("expected error after all retries")
	}
}

func TestExecute_ContextCancelled(t *testing.T) {
	s := retry.Strategy{MaxRetries: 10, InitialDelay: time.Second, MaxDelay: time.Second, BackoffFactor: 1}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := retry.Execute(ctx, s, func(ctx context.Context) error {
		return fmt.Errorf("error")
	})
	if err == nil {
		t.Fatal("expected context error")
	}
}
