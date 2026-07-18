package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

type PostgresProviderRepository struct{ pool *pgxpool.Pool }

func NewPostgresProviderRepository(pool *pgxpool.Pool) *PostgresProviderRepository {
	return &PostgresProviderRepository{pool: pool}
}

func (r *PostgresProviderRepository) Create(ctx context.Context, p *provider.Provider) error {
	configJSON, _ := json.Marshal(p.Config)
	_, err := r.pool.Exec(ctx, `INSERT INTO providers (id,name,type,config,status,priority,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		p.ID, p.Name, string(p.Type), configJSON, string(p.Status), p.Priority, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return provider.ErrDuplicateProvider
		}
		return fmt.Errorf("inserting provider: %w", err)
	}
	return nil
}

func (r *PostgresProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*provider.Provider, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,type,config,status,priority,created_at,updated_at FROM providers WHERE id=$1`, id)
	return r.scanProvider(row)
}

func (r *PostgresProviderRepository) List(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM providers`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id,name,type,config,status,priority,created_at,updated_at FROM providers ORDER BY priority LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*provider.Provider
	for rows.Next() {
		p, err := r.scanProviderFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, p)
	}
	return result, total, nil
}

func (r *PostgresProviderRepository) ListActive(ctx context.Context) ([]*provider.Provider, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,type,config,status,priority,created_at,updated_at FROM providers WHERE status='active' ORDER BY priority`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*provider.Provider
	for rows.Next() {
		p, err := r.scanProviderFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *PostgresProviderRepository) Update(ctx context.Context, p *provider.Provider) error {
	configJSON, _ := json.Marshal(p.Config)
	res, err := r.pool.Exec(ctx, `UPDATE providers SET name=$2,config=$3,status=$4,priority=$5,updated_at=$6 WHERE id=$1`,
		p.ID, p.Name, configJSON, string(p.Status), p.Priority, p.UpdatedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return provider.ErrProviderNotFound
	}
	return nil
}

func (r *PostgresProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM providers WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return provider.ErrProviderNotFound
	}
	return nil
}

func (r *PostgresProviderRepository) scanProvider(row pgx.Row) (*provider.Provider, error) {
	var p provider.Provider
	var typ, status string
	var configJSON []byte
	err := row.Scan(&p.ID, &p.Name, &typ, &configJSON, &status, &p.Priority, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, provider.ErrProviderNotFound
		}
		return nil, err
	}
	p.Type = provider.Type(typ)
	p.Status = provider.Status(status)
	_ = json.Unmarshal(configJSON, &p.Config)
	return &p, nil
}

func (r *PostgresProviderRepository) scanProviderFromRows(rows pgx.Rows) (*provider.Provider, error) {
	var p provider.Provider
	var typ, status string
	var configJSON []byte
	err := rows.Scan(&p.ID, &p.Name, &typ, &configJSON, &status, &p.Priority, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Type = provider.Type(typ)
	p.Status = provider.Status(status)
	_ = json.Unmarshal(configJSON, &p.Config)
	return &p, nil
}
