// Package provider implements provider management use cases.
package provider

import (
	"context"
	"time"

	"github.com/google/uuid"
	domain "github.com/paymentbridge/pcp/internal/domain/provider"
)

// Service orchestrates provider management operations.
type Service struct {
	repo     domain.Repository
	gateways map[domain.Type]domain.Gateway
}

// NewService creates a new provider service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo, gateways: make(map[domain.Type]domain.Gateway)}
}

// RegisterGateway registers a payment gateway implementation for a provider type.
func (s *Service) RegisterGateway(providerType domain.Type, gw domain.Gateway) {
	s.gateways[providerType] = gw
}

// GetGateway returns the gateway for a given provider type.
func (s *Service) GetGateway(providerType domain.Type) (domain.Gateway, error) {
	gw, ok := s.gateways[providerType]
	if !ok {
		return nil, domain.ErrProviderUnavailable
	}
	return gw, nil
}

// CreateInput holds input for creating a provider.
type CreateInput struct {
	Name     string
	Type     domain.Type
	Config   map[string]string
	Priority int
}

// Create registers a new provider.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Provider, error) {
	now := time.Now().UTC()
	p := &domain.Provider{
		ID:        uuid.New(),
		Name:      input.Name,
		Type:      input.Type,
		Config:    input.Config,
		Status:    domain.StatusActive,
		Priority:  input.Priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetByID retrieves a provider by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Provider, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves providers with pagination.
func (s *Service) List(ctx context.Context, offset, limit int) ([]*domain.Provider, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, offset, limit)
}

// ListActive retrieves all active providers.
func (s *Service) ListActive(ctx context.Context) ([]*domain.Provider, error) {
	return s.repo.ListActive(ctx)
}

// Update modifies a provider.
func (s *Service) Update(ctx context.Context, id uuid.UUID, name *string, config map[string]string, status *domain.Status, priority *int) (*domain.Provider, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		p.Name = *name
	}
	if config != nil {
		p.Config = config
	}
	if status != nil {
		p.Status = *status
	}
	if priority != nil {
		p.Priority = *priority
	}
	p.UpdatedAt = time.Now().UTC()
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete removes a provider.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
