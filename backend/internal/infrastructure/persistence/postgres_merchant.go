package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
)

// PostgresMerchantRepository implements merchant.Repository using PostgreSQL.
type PostgresMerchantRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresMerchantRepository creates a new PostgreSQL-backed merchant repository.
func NewPostgresMerchantRepository(pool *pgxpool.Pool) *PostgresMerchantRepository {
	return &PostgresMerchantRepository{pool: pool}
}

// Create inserts a new merchant into the database.
func (r *PostgresMerchantRepository) Create(ctx context.Context, m *merchant.Merchant) error {
	query := `
		INSERT INTO merchants (id, name, api_key, webhook_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		m.ID, m.Name, m.APIKey, m.WebhookURL, string(m.Status), m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return merchant.ErrDuplicateMerchant
		}
		return fmt.Errorf("inserting merchant: %w", err)
	}
	return nil
}

// GetByID retrieves a merchant by its UUID.
func (r *PostgresMerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
	query := `
		SELECT id, name, api_key, webhook_url, status, created_at, updated_at
		FROM merchants
		WHERE id = $1`

	return r.scanMerchant(r.pool.QueryRow(ctx, query, id))
}

// GetByAPIKey retrieves a merchant by its API key.
func (r *PostgresMerchantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*merchant.Merchant, error) {
	query := `
		SELECT id, name, api_key, webhook_url, status, created_at, updated_at
		FROM merchants
		WHERE api_key = $1`

	return r.scanMerchant(r.pool.QueryRow(ctx, query, apiKey))
}

// List retrieves a paginated list of merchants ordered by creation date descending.
func (r *PostgresMerchantRepository) List(ctx context.Context, offset, limit int) ([]*merchant.Merchant, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM merchants`
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting merchants: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, name, api_key, webhook_url, status, created_at, updated_at
		FROM merchants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing merchants: %w", err)
	}
	defer rows.Close()

	merchants := make([]*merchant.Merchant, 0)
	for rows.Next() {
		m, err := r.scanMerchantFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		merchants = append(merchants, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating merchant rows: %w", err)
	}

	return merchants, total, nil
}

// Update modifies an existing merchant in the database.
func (r *PostgresMerchantRepository) Update(ctx context.Context, m *merchant.Merchant) error {
	query := `
		UPDATE merchants
		SET name = $2, webhook_url = $3, status = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		m.ID, m.Name, m.WebhookURL, string(m.Status), m.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return merchant.ErrDuplicateMerchant
		}
		return fmt.Errorf("updating merchant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return merchant.ErrMerchantNotFound
	}
	return nil
}

// Delete removes a merchant from the database by ID.
func (r *PostgresMerchantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM merchants WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting merchant: %w", err)
	}
	if result.RowsAffected() == 0 {
		return merchant.ErrMerchantNotFound
	}
	return nil
}

// scanMerchant scans a single row into a Merchant entity.
func (r *PostgresMerchantRepository) scanMerchant(row pgx.Row) (*merchant.Merchant, error) {
	var m merchant.Merchant
	var status string
	err := row.Scan(&m.ID, &m.Name, &m.APIKey, &m.WebhookURL, &status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, merchant.ErrMerchantNotFound
		}
		return nil, fmt.Errorf("scanning merchant: %w", err)
	}
	m.Status = merchant.Status(status)
	return &m, nil
}

// scanMerchantFromRows scans a row from a Rows result set into a Merchant entity.
func (r *PostgresMerchantRepository) scanMerchantFromRows(rows pgx.Rows) (*merchant.Merchant, error) {
	var m merchant.Merchant
	var status string
	err := rows.Scan(&m.ID, &m.Name, &m.APIKey, &m.WebhookURL, &status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning merchant row: %w", err)
	}
	m.Status = merchant.Status(status)
	return &m, nil
}
