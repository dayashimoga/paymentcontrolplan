package webhook

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	domainwebhook "github.com/paymentbridge/pcp/internal/domain/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockWebhookRepo is an in-memory webhook repository for testing.
type mockWebhookRepo struct {
	webhooks map[uuid.UUID]*domainwebhook.Webhook
}

func newMockRepo() *mockWebhookRepo {
	return &mockWebhookRepo{webhooks: make(map[uuid.UUID]*domainwebhook.Webhook)}
}

func (r *mockWebhookRepo) Create(_ context.Context, w *domainwebhook.Webhook) error {
	r.webhooks[w.ID] = w
	return nil
}

func (r *mockWebhookRepo) GetByID(_ context.Context, id uuid.UUID) (*domainwebhook.Webhook, error) {
	if w, ok := r.webhooks[id]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("webhook not found")
}

func (r *mockWebhookRepo) ListPending(_ context.Context, limit int) ([]*domainwebhook.Webhook, error) {
	var result []*domainwebhook.Webhook
	now := time.Now().UTC()
	for _, w := range r.webhooks {
		if w.Status == domainwebhook.StatusPending && !w.NextRetry.After(now) {
			result = append(result, w)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *mockWebhookRepo) ListByPayment(_ context.Context, paymentID uuid.UUID) ([]*domainwebhook.Webhook, error) {
	var result []*domainwebhook.Webhook
	for _, w := range r.webhooks {
		if w.PaymentID == paymentID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (r *mockWebhookRepo) Update(_ context.Context, w *domainwebhook.Webhook) error {
	r.webhooks[w.ID] = w
	return nil
}

func TestWebhookService_Enqueue(t *testing.T) {
	repo := newMockRepo()
	logger := zap.NewNop()
	svc := NewService(repo, "test-secret", logger)

	err := svc.Enqueue(context.Background(), uuid.New(), uuid.New(), "https://example.com/webhook", "payment.completed", map[string]string{"status": "completed"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.webhooks))

	for _, w := range repo.webhooks {
		assert.Equal(t, domainwebhook.StatusPending, w.Status)
		assert.Equal(t, "payment.completed", w.EventType)
		assert.Equal(t, 5, w.MaxRetries)
	}
}

func TestWebhookService_ProcessPending_Success(t *testing.T) {
	var callCount int32
	// Create a test HTTP server that accepts webhooks
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-PCP-Event"))
		assert.NotEmpty(t, r.Header.Get("X-PCP-Delivery"))
		assert.NotEmpty(t, r.Header.Get("X-PCP-Timestamp"))
		assert.NotEmpty(t, r.Header.Get("X-PCP-Signature"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	repo := newMockRepo()
	logger := zap.NewNop()
	svc := NewService(repo, "test-secret", logger)

	// Enqueue a webhook
	err := svc.Enqueue(context.Background(), uuid.New(), uuid.New(), ts.URL, "payment.completed", map[string]string{"status": "completed"})
	require.NoError(t, err)

	// Process pending
	processed, err := svc.ProcessPending(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Check that webhook status is now delivered
	for _, w := range repo.webhooks {
		assert.Equal(t, domainwebhook.StatusDelivered, w.Status)
	}
}

func TestWebhookService_ProcessPending_Failure(t *testing.T) {
	// Server that rejects webhooks
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	repo := newMockRepo()
	logger := zap.NewNop()
	svc := NewService(repo, "test-secret", logger)

	err := svc.Enqueue(context.Background(), uuid.New(), uuid.New(), ts.URL, "payment.failed", map[string]string{"status": "failed"})
	require.NoError(t, err)

	processed, err := svc.ProcessPending(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Webhook should still be pending with incremented attempt count
	for _, w := range repo.webhooks {
		assert.Equal(t, domainwebhook.StatusPending, w.Status)
		assert.Equal(t, 1, w.Attempts)
		assert.True(t, w.NextRetry.After(time.Now().UTC()))
	}
}

func TestVerifySignature(t *testing.T) {
	key := "my-signing-key"
	timestamp := "2026-01-01T00:00:00Z"
	payload := []byte(`{"status":"completed"}`)

	// Generate signature
	svc := NewService(nil, key, zap.NewNop())
	sig := svc.sign(payload, timestamp)

	// Verify should pass
	assert.True(t, VerifySignature(key, timestamp, sig, payload))

	// Tampered payload should fail
	assert.False(t, VerifySignature(key, timestamp, sig, []byte(`{"status":"failed"}`)))

	// Wrong key should fail
	assert.False(t, VerifySignature("wrong-key", timestamp, sig, payload))

	// Empty key or signature should fail
	assert.False(t, VerifySignature("", timestamp, sig, payload))
	assert.False(t, VerifySignature(key, timestamp, "", payload))
}

func TestPow(t *testing.T) {
	assert.Equal(t, 1, pow(3, 0))
	assert.Equal(t, 3, pow(3, 1))
	assert.Equal(t, 9, pow(3, 2))
	assert.Equal(t, 27, pow(3, 3))
}
