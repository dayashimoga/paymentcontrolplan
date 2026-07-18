package dto

import (
	"time"
	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

type CreateProviderRequest struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Config   map[string]string `json:"config"`
	Priority int               `json:"priority"`
}

type ProviderResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListProvidersResponse struct {
	Data   []ProviderResponse `json:"data"`
	Total  int                `json:"total"`
	Offset int                `json:"offset"`
	Limit  int                `json:"limit"`
}

func ToProviderResponse(p *provider.Provider) ProviderResponse {
	return ProviderResponse{ID: p.ID, Name: p.Name, Type: string(p.Type), Status: string(p.Status), Priority: p.Priority, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
}
