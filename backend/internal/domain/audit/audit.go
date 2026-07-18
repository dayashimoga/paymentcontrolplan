// Package audit defines the Audit Log bounded context.
package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an immutable audit trail entry.
type AuditLog struct {
	ID         uuid.UUID              `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   uuid.UUID              `json:"entity_id"`
	Action     string                 `json:"action"`
	ActorID    uuid.UUID              `json:"actor_id"`
	Changes    map[string]interface{} `json:"changes"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Repository defines the port for audit log persistence.
type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, offset, limit int) ([]*AuditLog, int, error)
	ListByActor(ctx context.Context, actorID uuid.UUID, offset, limit int) ([]*AuditLog, int, error)
}

// Service provides audit logging capabilities.
type Service struct {
	repo Repository
}

// NewService creates a new audit service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Log records an audit entry.
func (s *Service) Log(ctx context.Context, entityType string, entityID, actorID uuid.UUID, action string, changes map[string]interface{}, ip, ua string) error {
	entry := &AuditLog{
		ID: uuid.New(), EntityType: entityType, EntityID: entityID,
		Action: action, ActorID: actorID, Changes: changes,
		IPAddress: ip, UserAgent: ua, CreatedAt: time.Now().UTC(),
	}
	return s.repo.Create(ctx, entry)
}

// ListByEntity retrieves audit logs for a specific entity.
func (s *Service) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, offset, limit int) ([]*AuditLog, int, error) {
	return s.repo.ListByEntity(ctx, entityType, entityID, offset, limit)
}
