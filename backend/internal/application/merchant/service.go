// Package merchant implements application-level use cases for the Merchant bounded context.
// It orchestrates domain logic and infrastructure concerns without leaking implementation details.
package merchant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/paymentbridge/pcp/internal/domain/merchant"
)

// Service implements merchant use cases by coordinating the domain model
// and repository port. It is the primary driving adapter target.
type Service struct {
	repo domain.Repository
}

// NewService creates a new merchant application service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput holds the validated input for creating a merchant.
type CreateInput struct {
	Name       string
	WebhookURL string
}

// UpdateInput holds the optional fields for updating a merchant.
// Nil fields are not changed.
type UpdateInput struct {
	Name       *string
	WebhookURL *string
	Status     *domain.Status
}

// Create registers a new merchant, generating a unique API key.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Merchant, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generating API key: %w", err)
	}

	now := time.Now().UTC()
	m := &domain.Merchant{
		ID:         uuid.New(),
		Name:       input.Name,
		APIKey:     apiKey,
		WebhookURL: input.WebhookURL,
		Status:     domain.StatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// GetByID retrieves a single merchant by its unique identifier.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves a paginated list of merchants.
func (s *Service) List(ctx context.Context, offset, limit int) ([]*domain.Merchant, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, offset, limit)
}

// Update modifies an existing merchant with the provided fields.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Merchant, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		m.Name = *input.Name
	}
	if input.WebhookURL != nil {
		m.WebhookURL = *input.WebhookURL
	}
	if input.Status != nil {
		m.Status = *input.Status
	}
	m.UpdatedAt = time.Now().UTC()

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// Delete removes a merchant by its unique identifier.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// GetByAPIKey retrieves a merchant by API key for authentication.
func (s *Service) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Merchant, error) {
	return s.repo.GetByAPIKey(ctx, apiKey)
}

// RotateAPIKey generates a new API key for a merchant and updates persistence.
func (s *Service) RotateAPIKey(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generating API key: %w", err)
	}

	m.APIKey = newKey
	m.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// generateAPIKey creates a cryptographically random API key with the "pcp_" prefix.
func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pcp_" + hex.EncodeToString(b), nil
}
