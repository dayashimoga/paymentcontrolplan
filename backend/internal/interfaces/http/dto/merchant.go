// Package dto defines the Data Transfer Objects for HTTP request/response payloads.
// DTOs decouple the HTTP interface contract from domain entities, allowing each to evolve independently.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
)

// CreateMerchantRequest represents the payload for creating a new merchant.
type CreateMerchantRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=255"`
	WebhookURL string `json:"webhook_url" validate:"omitempty,url"`
}

// UpdateMerchantRequest represents the payload for updating an existing merchant.
type UpdateMerchantRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	WebhookURL *string `json:"webhook_url,omitempty" validate:"omitempty,url"`
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=active suspended inactive"`
}

// MerchantResponse represents a merchant in API responses.
type MerchantResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	APIKey     string    `json:"api_key"`
	WebhookURL string    `json:"webhook_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListMerchantsResponse represents a paginated list of merchants.
type ListMerchantsResponse struct {
	Data       []MerchantResponse `json:"data"`
	Total      int                `json:"total"`
	Offset     int                `json:"offset"`
	Limit      int                `json:"limit"`
}

// ErrorResponse represents a standardized API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ToMerchantResponse converts a domain Merchant entity to an API response DTO.
func ToMerchantResponse(m *merchant.Merchant) MerchantResponse {
	return MerchantResponse{
		ID:         m.ID,
		Name:       m.Name,
		APIKey:     m.APIKey,
		WebhookURL: m.WebhookURL,
		Status:     string(m.Status),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// ToMerchantListResponse converts a slice of domain merchants to a paginated response.
func ToMerchantListResponse(merchants []*merchant.Merchant, total, offset, limit int) ListMerchantsResponse {
	data := make([]MerchantResponse, 0, len(merchants))
	for _, m := range merchants {
		data = append(data, ToMerchantResponse(m))
	}
	return ListMerchantsResponse{
		Data:   data,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}
}
