package dto

import (
	"time"
	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/payment"
)

type CreatePaymentRequest struct {
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Description    string            `json:"description"`
	Metadata       map[string]string `json:"metadata"`
	IdempotencyKey string            `json:"idempotency_key"`
}

type RefundRequest struct {
	PaymentID string `json:"payment_id"`
}

type PaymentResponse struct {
	ID             uuid.UUID         `json:"id"`
	MerchantID     uuid.UUID         `json:"merchant_id"`
	ProviderID     uuid.UUID         `json:"provider_id"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Status         string            `json:"status"`
	ExternalID     string            `json:"external_id"`
	Description    string            `json:"description"`
	Metadata       map[string]string `json:"metadata"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ListPaymentsResponse struct {
	Data   []PaymentResponse `json:"data"`
	Total  int               `json:"total"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
}

func ToPaymentResponse(p *payment.Payment) PaymentResponse {
	return PaymentResponse{
		ID: p.ID, MerchantID: p.MerchantID, ProviderID: p.ProviderID,
		Amount: p.Amount, Currency: p.Currency, Status: string(p.Status),
		ExternalID: p.ExternalID, Description: p.Description, Metadata: p.Metadata,
		ErrorMessage: p.ErrorMessage, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
