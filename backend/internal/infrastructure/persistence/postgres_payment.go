package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paymentbridge/pcp/internal/domain/payment"
)

type PostgresPaymentRepository struct{ pool *pgxpool.Pool }

func NewPostgresPaymentRepository(pool *pgxpool.Pool) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{pool: pool}
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, p *payment.Payment) error {
	metaJSON, _ := json.Marshal(p.Metadata)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payments (id,merchant_id,provider_id,amount,currency,status,external_id,idempotency_key,description,metadata,error_message,attempt_count,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.MerchantID, p.ProviderID, p.Amount, p.Currency, string(p.Status), p.ExternalID,
		p.IdempotencyKey, p.Description, metaJSON, p.ErrorMessage, p.AttemptCount, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,merchant_id,provider_id,amount,currency,status,external_id,idempotency_key,description,metadata,error_message,attempt_count,created_at,updated_at FROM payments WHERE id=$1`, id)
	return r.scanPayment(row)
}

func (r *PostgresPaymentRepository) GetByIdempotencyKey(ctx context.Context, merchantID uuid.UUID, key string) (*payment.Payment, error) {
	if key == "" {
		return nil, payment.ErrPaymentNotFound
	}
	row := r.pool.QueryRow(ctx, `SELECT id,merchant_id,provider_id,amount,currency,status,external_id,idempotency_key,description,metadata,error_message,attempt_count,created_at,updated_at FROM payments WHERE merchant_id=$1 AND idempotency_key=$2`, merchantID, key)
	return r.scanPayment(row)
}

func (r *PostgresPaymentRepository) List(ctx context.Context, merchantID uuid.UUID, offset, limit int) ([]*payment.Payment, int, error) {
	var total int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE merchant_id=$1`, merchantID).Scan(&total)
	rows, err := r.pool.Query(ctx, `SELECT id,merchant_id,provider_id,amount,currency,status,external_id,idempotency_key,description,metadata,error_message,attempt_count,created_at,updated_at FROM payments WHERE merchant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, merchantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*payment.Payment
	for rows.Next() {
		p, err := r.scanPaymentFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, p)
	}
	return result, total, nil
}

func (r *PostgresPaymentRepository) Update(ctx context.Context, p *payment.Payment) error {
	metaJSON, _ := json.Marshal(p.Metadata)
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET provider_id=$2,status=$3,external_id=$4,error_message=$5,attempt_count=$6,metadata=$7,updated_at=$8 WHERE id=$1`,
		p.ID, p.ProviderID, string(p.Status), p.ExternalID, p.ErrorMessage, p.AttemptCount, metaJSON, p.UpdatedAt)
	return err
}

func (r *PostgresPaymentRepository) scanPayment(row pgx.Row) (*payment.Payment, error) {
	var p payment.Payment
	var status string
	var metaJSON []byte
	err := row.Scan(&p.ID, &p.MerchantID, &p.ProviderID, &p.Amount, &p.Currency, &status, &p.ExternalID, &p.IdempotencyKey, &p.Description, &metaJSON, &p.ErrorMessage, &p.AttemptCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, payment.ErrPaymentNotFound
		}
		return nil, err
	}
	p.Status = payment.Status(status)
	_ = json.Unmarshal(metaJSON, &p.Metadata)
	return &p, nil
}

func (r *PostgresPaymentRepository) scanPaymentFromRows(rows pgx.Rows) (*payment.Payment, error) {
	var p payment.Payment
	var status string
	var metaJSON []byte
	err := rows.Scan(&p.ID, &p.MerchantID, &p.ProviderID, &p.Amount, &p.Currency, &status, &p.ExternalID, &p.IdempotencyKey, &p.Description, &metaJSON, &p.ErrorMessage, &p.AttemptCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Status = payment.Status(status)
	_ = json.Unmarshal(metaJSON, &p.Metadata)
	return &p, nil
}
