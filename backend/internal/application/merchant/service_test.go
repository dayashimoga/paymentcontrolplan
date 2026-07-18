package merchant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	appmch "github.com/paymentbridge/pcp/internal/application/merchant"
	"github.com/paymentbridge/pcp/internal/domain/common"
	domain "github.com/paymentbridge/pcp/internal/domain/merchant"
)

// mockRepository implements domain.Repository for unit testing.
type mockRepository struct {
	merchants map[uuid.UUID]*domain.Merchant
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		merchants: make(map[uuid.UUID]*domain.Merchant),
	}
}

func (r *mockRepository) Create(_ context.Context, m *domain.Merchant) error {
	for _, existing := range r.merchants {
		if existing.Name == m.Name {
			return domain.ErrDuplicateMerchant
		}
	}
	r.merchants[m.ID] = m
	return nil
}

func (r *mockRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.Merchant, error) {
	m, ok := r.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return m, nil
}

func (r *mockRepository) GetByAPIKey(_ context.Context, apiKey string) (*domain.Merchant, error) {
	for _, m := range r.merchants {
		if m.APIKey == apiKey {
			return m, nil
		}
	}
	return nil, domain.ErrMerchantNotFound
}

func (r *mockRepository) List(_ context.Context, offset, limit int) ([]*domain.Merchant, int, error) {
	all := make([]*domain.Merchant, 0, len(r.merchants))
	for _, m := range r.merchants {
		all = append(all, m)
	}
	total := len(all)
	if offset >= total {
		return []*domain.Merchant{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *mockRepository) Update(_ context.Context, m *domain.Merchant) error {
	if _, ok := r.merchants[m.ID]; !ok {
		return domain.ErrMerchantNotFound
	}
	r.merchants[m.ID] = m
	return nil
}

func (r *mockRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.merchants[id]; !ok {
		return domain.ErrMerchantNotFound
	}
	delete(r.merchants, id)
	return nil
}

func TestService_Create_Success(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	m, err := svc.Create(context.Background(), appmch.CreateInput{
		Name:       "Acme Corp",
		WebhookURL: "https://acme.com/webhook",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %s", m.Name)
	}
	if m.Status != domain.StatusActive {
		t.Errorf("expected active status, got %s", m.Status)
	}
	if m.APIKey == "" {
		t.Error("expected API key to be generated")
	}
	if len(m.APIKey) < 10 {
		t.Error("API key too short")
	}
}

func TestService_Create_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	_, err := svc.Create(context.Background(), appmch.CreateInput{
		Name: "",
	})

	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Create_DuplicateName(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	_, err := svc.Create(context.Background(), appmch.CreateInput{Name: "Acme"})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = svc.Create(context.Background(), appmch.CreateInput{Name: "Acme"})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
	if !errors.Is(err, common.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestService_GetByID_Success(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	created, _ := svc.Create(context.Background(), appmch.CreateInput{Name: "Acme"})

	m, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, m.ID)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	for i := 0; i < 5; i++ {
		_, err := svc.Create(context.Background(), appmch.CreateInput{
			Name: "Merchant " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}

	merchants, total, err := svc.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(merchants) != 5 {
		t.Errorf("expected 5 merchants, got %d", len(merchants))
	}
}

func TestService_List_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	for i := 0; i < 5; i++ {
		_, _ = svc.Create(context.Background(), appmch.CreateInput{
			Name: "Merchant " + string(rune('A'+i)),
		})
	}

	merchants, total, err := svc.List(context.Background(), 0, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(merchants) != 2 {
		t.Errorf("expected 2 merchants, got %d", len(merchants))
	}
}

func TestService_Update_Success(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	created, _ := svc.Create(context.Background(), appmch.CreateInput{Name: "Acme"})

	newName := "Acme Updated"
	updated, err := svc.Update(context.Background(), created.ID, appmch.UpdateInput{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Acme Updated" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	name := "test"
	_, err := svc.Update(context.Background(), uuid.New(), appmch.UpdateInput{Name: &name})
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Delete_Success(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	created, _ := svc.Create(context.Background(), appmch.CreateInput{Name: "Acme"})

	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetByID(context.Background(), created.ID)
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected merchant to be deleted")
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := appmch.NewService(repo)

	err := svc.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}
