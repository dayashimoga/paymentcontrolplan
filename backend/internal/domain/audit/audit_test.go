package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, l *audit.AuditLog) error {
	return m.Called(ctx, l).Error(0)
}

func (m *mockRepo) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, offset, limit int) ([]*audit.AuditLog, int, error) {
	args := m.Called(ctx, entityType, entityID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*audit.AuditLog), args.Int(1), args.Error(2)
}

func (m *mockRepo) ListByActor(ctx context.Context, actorID uuid.UUID, offset, limit int) ([]*audit.AuditLog, int, error) {
	args := m.Called(ctx, actorID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*audit.AuditLog), args.Int(1), args.Error(2)
}

func TestAuditService(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	svc := audit.NewService(repo)

	entityID := uuid.New()
	actorID := uuid.New()
	changes := map[string]interface{}{"name": "Acme"}

	repo.On("Create", ctx, mock.Anything).Return(nil)

	err := svc.Log(ctx, "merchant", entityID, actorID, "CREATE", changes, "127.0.0.1", "Go-Test")
	assert.NoError(t, err)
}
