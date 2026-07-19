// Package webhook implements the webhook delivery engine with HTTP dispatch,
// HMAC-SHA256 signature verification, and retry with exponential backoff.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	domainwebhook "github.com/paymentbridge/pcp/internal/domain/webhook"
	"go.uber.org/zap"
)

// Service handles webhook creation and delivery.
type Service struct {
	repo       domainwebhook.Repository
	httpClient *http.Client
	logger     *zap.Logger
	signingKey string
}

// NewService creates a new webhook delivery service.
func NewService(repo domainwebhook.Repository, signingKey string, logger *zap.Logger) *Service {
	return &Service{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:     logger,
		signingKey: signingKey,
	}
}

// DeliveryResult captures the outcome of a webhook delivery attempt.
type DeliveryResult struct {
	WebhookID  uuid.UUID `json:"webhook_id"`
	StatusCode int       `json:"status_code"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

// Enqueue creates a new webhook entry for async delivery.
func (s *Service) Enqueue(ctx context.Context, merchantID, paymentID uuid.UUID, url, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	now := time.Now().UTC()
	w := &domainwebhook.Webhook{
		ID:         uuid.New(),
		MerchantID: merchantID,
		PaymentID:  paymentID,
		URL:        url,
		EventType:  eventType,
		Payload:    string(data),
		Status:     domainwebhook.StatusPending,
		Attempts:   0,
		MaxRetries: 5,
		NextRetry:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return s.repo.Create(ctx, w)
}

// ProcessPending fetches pending webhooks and attempts delivery.
// Returns the number of webhooks processed.
func (s *Service) ProcessPending(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 50
	}

	webhooks, err := s.repo.ListPending(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("listing pending webhooks: %w", err)
	}

	processed := 0
	for _, w := range webhooks {
		if err := ctx.Err(); err != nil {
			return processed, err
		}

		result := s.deliver(ctx, w)
		if result.Success {
			w.Status = domainwebhook.StatusDelivered
		} else {
			w.Attempts++
			if w.Attempts >= w.MaxRetries {
				w.Status = domainwebhook.StatusFailed
				s.logger.Warn("webhook delivery permanently failed",
					zap.String("webhook_id", w.ID.String()),
					zap.String("url", w.URL),
					zap.Int("attempts", w.Attempts),
				)
			} else {
				// Exponential backoff: 10s, 30s, 90s, 270s, 810s
				delay := time.Duration(10*pow(3, w.Attempts)) * time.Second
				w.NextRetry = time.Now().UTC().Add(delay)
			}
		}
		w.UpdatedAt = time.Now().UTC()

		if err := s.repo.Update(ctx, w); err != nil {
			s.logger.Error("failed to update webhook status",
				zap.String("webhook_id", w.ID.String()),
				zap.Error(err),
			)
		}
		processed++
	}

	return processed, nil
}

// deliver sends the webhook payload to the target URL with signature.
func (s *Service) deliver(ctx context.Context, w *domainwebhook.Webhook) DeliveryResult {
	result := DeliveryResult{WebhookID: w.ID}

	body := []byte(w.Payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		result.Error = fmt.Sprintf("creating request: %v", err)
		return result
	}

	// Set headers
	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := s.sign(body, timestamp)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PCP-Event", w.EventType)
	req.Header.Set("X-PCP-Delivery", w.ID.String())
	req.Header.Set("X-PCP-Timestamp", timestamp)
	req.Header.Set("X-PCP-Signature", signature)
	req.Header.Set("User-Agent", "PCP-Webhook/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("delivering webhook: %v", err)
		s.logger.Warn("webhook delivery failed",
			zap.String("webhook_id", w.ID.String()),
			zap.String("url", w.URL),
			zap.Error(err),
		)
		return result
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	result.StatusCode = resp.StatusCode
	// 2xx status codes are considered successful
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !result.Success {
		result.Error = fmt.Sprintf("non-2xx status: %d", resp.StatusCode)
		s.logger.Warn("webhook delivery rejected",
			zap.String("webhook_id", w.ID.String()),
			zap.String("url", w.URL),
			zap.Int("status", resp.StatusCode),
		)
	}

	return result
}

// sign creates an HMAC-SHA256 signature for webhook verification.
// Merchants verify by: HMAC-SHA256(signing_key, timestamp + "." + payload)
func (s *Service) sign(payload []byte, timestamp string) string {
	if s.signingKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(s.signingKey))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies a webhook signature. Exported for merchant SDK use.
func VerifySignature(signingKey, timestamp, signature string, payload []byte) bool {
	if signingKey == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ListByPayment returns webhook deliveries for a specific payment.
func (s *Service) ListByPayment(ctx context.Context, paymentID uuid.UUID) ([]*domainwebhook.Webhook, error) {
	return s.repo.ListByPayment(ctx, paymentID)
}

// GetByID retrieves a specific webhook by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domainwebhook.Webhook, error) {
	return s.repo.GetByID(ctx, id)
}

// pow returns base^exp for integers (simple helper, no math.Pow float conversion).
func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
