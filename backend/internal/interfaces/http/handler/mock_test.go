package handler_test

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/paymentbridge/pcp/internal/domain/merchant"
)

// testMockRepo is a shared mock for handler tests.
type testMockRepo struct {
	merchants map[uuid.UUID]*domain.Merchant
}

func newTestMockRepo() *testMockRepo {
	return &testMockRepo{
		merchants: make(map[uuid.UUID]*domain.Merchant),
	}
}

func (r *testMockRepo) Create(_ context.Context, m *domain.Merchant) error {
	for _, existing := range r.merchants {
		if existing.Name == m.Name {
			return domain.ErrDuplicateMerchant
		}
	}
	r.merchants[m.ID] = m
	return nil
}

func (r *testMockRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Merchant, error) {
	m, ok := r.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return m, nil
}

func (r *testMockRepo) GetByAPIKey(_ context.Context, apiKey string) (*domain.Merchant, error) {
	for _, m := range r.merchants {
		if m.APIKey == apiKey {
			return m, nil
		}
	}
	return nil, domain.ErrMerchantNotFound
}

func (r *testMockRepo) List(_ context.Context, offset, limit int) ([]*domain.Merchant, int, error) {
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

func (r *testMockRepo) Update(_ context.Context, m *domain.Merchant) error {
	if _, ok := r.merchants[m.ID]; !ok {
		return domain.ErrMerchantNotFound
	}
	r.merchants[m.ID] = m
	return nil
}

func (r *testMockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.merchants[id]; !ok {
		return domain.ErrMerchantNotFound
	}
	delete(r.merchants, id)
	return nil
}
