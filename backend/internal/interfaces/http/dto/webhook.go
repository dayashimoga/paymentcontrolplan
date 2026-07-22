package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/webhook"
)

// WebhookResponse represents a webhook delivery in API responses.
type WebhookResponse struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	PaymentID  uuid.UUID `json:"payment_id"`
	URL        string    `json:"url"`
	EventType  string    `json:"event_type"`
	Status     string    `json:"status"`
	Attempts   int       `json:"attempts"`
	MaxRetries int       `json:"max_retries"`
	NextRetry  time.Time `json:"next_retry"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListWebhooksResponse represents a list of webhook deliveries.
type ListWebhooksResponse struct {
	Data []WebhookResponse `json:"data"`
}

// ToWebhookResponse converts a domain Webhook to an API response DTO.
func ToWebhookResponse(w *webhook.Webhook) WebhookResponse {
	return WebhookResponse{
		ID:         w.ID,
		MerchantID: w.MerchantID,
		PaymentID:  w.PaymentID,
		URL:        w.URL,
		EventType:  w.EventType,
		Status:     string(w.Status),
		Attempts:   w.Attempts,
		MaxRetries: w.MaxRetries,
		NextRetry:  w.NextRetry,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

// ToWebhookListResponse converts a slice of domain webhooks to a list response.
func ToWebhookListResponse(webhooks []*webhook.Webhook) ListWebhooksResponse {
	data := make([]WebhookResponse, 0, len(webhooks))
	for _, w := range webhooks {
		data = append(data, ToWebhookResponse(w))
	}
	return ListWebhooksResponse{Data: data}
}
