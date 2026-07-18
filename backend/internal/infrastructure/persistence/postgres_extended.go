package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paymentbridge/pcp/internal/application/analytics"
	"github.com/paymentbridge/pcp/internal/domain/audit"
	"github.com/paymentbridge/pcp/internal/domain/reconciliation"
	"github.com/paymentbridge/pcp/internal/domain/webhook"
)

// --- Audit Repository ---
type PostgresAuditRepository struct{ pool *pgxpool.Pool }

func NewPostgresAuditRepository(pool *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{pool: pool}
}

func (r *PostgresAuditRepository) Create(ctx context.Context, l *audit.AuditLog) error {
	changesJSON, _ := json.Marshal(l.Changes)
	_, err := r.pool.Exec(ctx, `INSERT INTO audit_logs (id,entity_type,entity_id,action,actor_id,changes,ip_address,user_agent,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		l.ID, l.EntityType, l.EntityID, l.Action, l.ActorID, changesJSON, l.IPAddress, l.UserAgent, l.CreatedAt)
	return err
}

func (r *PostgresAuditRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, offset, limit int) ([]*audit.AuditLog, int, error) {
	var total int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE entity_type=$1 AND entity_id=$2`, entityType, entityID).Scan(&total)
	rows, err := r.pool.Query(ctx, `SELECT id,entity_type,entity_id,action,actor_id,changes,ip_address,user_agent,created_at FROM audit_logs WHERE entity_type=$1 AND entity_id=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, entityType, entityID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*audit.AuditLog
	for rows.Next() {
		var l audit.AuditLog
		var changesJSON []byte
		rows.Scan(&l.ID, &l.EntityType, &l.EntityID, &l.Action, &l.ActorID, &changesJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt)
		json.Unmarshal(changesJSON, &l.Changes)
		result = append(result, &l)
	}
	return result, total, nil
}

func (r *PostgresAuditRepository) ListByActor(ctx context.Context, actorID uuid.UUID, offset, limit int) ([]*audit.AuditLog, int, error) {
	var total int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE actor_id=$1`, actorID).Scan(&total)
	rows, err := r.pool.Query(ctx, `SELECT id,entity_type,entity_id,action,actor_id,changes,ip_address,user_agent,created_at FROM audit_logs WHERE actor_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, actorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*audit.AuditLog
	for rows.Next() {
		var l audit.AuditLog
		var changesJSON []byte
		rows.Scan(&l.ID, &l.EntityType, &l.EntityID, &l.Action, &l.ActorID, &changesJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt)
		json.Unmarshal(changesJSON, &l.Changes)
		result = append(result, &l)
	}
	return result, total, nil
}

// --- Webhook Repository ---
type PostgresWebhookRepository struct{ pool *pgxpool.Pool }

func NewPostgresWebhookRepository(pool *pgxpool.Pool) *PostgresWebhookRepository {
	return &PostgresWebhookRepository{pool: pool}
}

func (r *PostgresWebhookRepository) Create(ctx context.Context, w *webhook.Webhook) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO webhooks (id,merchant_id,payment_id,url,event_type,payload,status,attempts,max_retries,next_retry,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		w.ID, w.MerchantID, w.PaymentID, w.URL, w.EventType, w.Payload, string(w.Status), w.Attempts, w.MaxRetries, w.NextRetry, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *PostgresWebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*webhook.Webhook, error) {
	var w webhook.Webhook
	var status string
	err := r.pool.QueryRow(ctx, `SELECT id,merchant_id,payment_id,url,event_type,payload,status,attempts,max_retries,next_retry,created_at,updated_at FROM webhooks WHERE id=$1`, id).
		Scan(&w.ID, &w.MerchantID, &w.PaymentID, &w.URL, &w.EventType, &w.Payload, &status, &w.Attempts, &w.MaxRetries, &w.NextRetry, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	w.Status = webhook.Status(status)
	return &w, nil
}

func (r *PostgresWebhookRepository) ListPending(ctx context.Context, limit int) ([]*webhook.Webhook, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,merchant_id,payment_id,url,event_type,payload,status,attempts,max_retries,next_retry,created_at,updated_at FROM webhooks WHERE status='pending' AND next_retry <= NOW() ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*webhook.Webhook
	for rows.Next() {
		var w webhook.Webhook
		var status string
		rows.Scan(&w.ID, &w.MerchantID, &w.PaymentID, &w.URL, &w.EventType, &w.Payload, &status, &w.Attempts, &w.MaxRetries, &w.NextRetry, &w.CreatedAt, &w.UpdatedAt)
		w.Status = webhook.Status(status)
		result = append(result, &w)
	}
	return result, nil
}

func (r *PostgresWebhookRepository) ListByPayment(ctx context.Context, paymentID uuid.UUID) ([]*webhook.Webhook, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,merchant_id,payment_id,url,event_type,payload,status,attempts,max_retries,next_retry,created_at,updated_at FROM webhooks WHERE payment_id=$1 ORDER BY created_at`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*webhook.Webhook
	for rows.Next() {
		var w webhook.Webhook
		var status string
		rows.Scan(&w.ID, &w.MerchantID, &w.PaymentID, &w.URL, &w.EventType, &w.Payload, &status, &w.Attempts, &w.MaxRetries, &w.NextRetry, &w.CreatedAt, &w.UpdatedAt)
		w.Status = webhook.Status(status)
		result = append(result, &w)
	}
	return result, nil
}

func (r *PostgresWebhookRepository) Update(ctx context.Context, w *webhook.Webhook) error {
	_, err := r.pool.Exec(ctx, `UPDATE webhooks SET status=$2,attempts=$3,next_retry=$4,updated_at=$5 WHERE id=$1`,
		w.ID, string(w.Status), w.Attempts, w.NextRetry, w.UpdatedAt)
	return err
}

// --- Reconciliation Repository ---
type PostgresReconciliationRepository struct{ pool *pgxpool.Pool }

func NewPostgresReconciliationRepository(pool *pgxpool.Pool) *PostgresReconciliationRepository {
	return &PostgresReconciliationRepository{pool: pool}
}

func (r *PostgresReconciliationRepository) Create(ctx context.Context, rec *reconciliation.Record) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO reconciliation_records (id,payment_id,provider_id,internal_amount,external_amount,internal_status,external_status,is_matched,discrepancy,reconciled_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		rec.ID, rec.PaymentID, rec.ProviderID, rec.InternalAmount, rec.ExternalAmount, string(rec.InternalStatus), rec.ExternalStatus, rec.IsMatched, rec.Discrepancy, rec.ReconciledAt, rec.CreatedAt)
	return err
}

func (r *PostgresReconciliationRepository) GetByPayment(ctx context.Context, paymentID uuid.UUID) (*reconciliation.Record, error) {
	var rec reconciliation.Record
	var intStatus string
	err := r.pool.QueryRow(ctx, `SELECT id,payment_id,provider_id,internal_amount,external_amount,internal_status,external_status,is_matched,discrepancy,reconciled_at,created_at FROM reconciliation_records WHERE payment_id=$1 ORDER BY created_at DESC LIMIT 1`, paymentID).
		Scan(&rec.ID, &rec.PaymentID, &rec.ProviderID, &rec.InternalAmount, &rec.ExternalAmount, &intStatus, &rec.ExternalStatus, &rec.IsMatched, &rec.Discrepancy, &rec.ReconciledAt, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	rec.InternalStatus = reconciliation.Status(intStatus)
	return &rec, nil
}

func (r *PostgresReconciliationRepository) ListUnmatched(ctx context.Context, offset, limit int) ([]*reconciliation.Record, int, error) {
	var total int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reconciliation_records WHERE is_matched=false`).Scan(&total)
	rows, err := r.pool.Query(ctx, `SELECT id,payment_id,provider_id,internal_amount,external_amount,internal_status,external_status,is_matched,discrepancy,reconciled_at,created_at FROM reconciliation_records WHERE is_matched=false ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*reconciliation.Record
	for rows.Next() {
		var rec reconciliation.Record
		var intStatus string
		rows.Scan(&rec.ID, &rec.PaymentID, &rec.ProviderID, &rec.InternalAmount, &rec.ExternalAmount, &intStatus, &rec.ExternalStatus, &rec.IsMatched, &rec.Discrepancy, &rec.ReconciledAt, &rec.CreatedAt)
		rec.InternalStatus = reconciliation.Status(intStatus)
		result = append(result, &rec)
	}
	return result, total, nil
}

// --- Analytics Repository ---
type PostgresAnalyticsRepository struct{ pool *pgxpool.Pool }

func NewPostgresAnalyticsRepository(pool *pgxpool.Pool) *PostgresAnalyticsRepository {
	return &PostgresAnalyticsRepository{pool: pool}
}

func (r *PostgresAnalyticsRepository) GetSummary(ctx context.Context, merchantID uuid.UUID, from, to time.Time) (*analytics.Summary, error) {
	var s analytics.Summary
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN status='failed' THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN status='completed' THEN amount ELSE 0 END),0)
		FROM payments WHERE merchant_id=$1 AND created_at BETWEEN $2 AND $3`, merchantID, from, to).
		Scan(&s.TotalPayments, &s.CompletedPayments, &s.FailedPayments, &s.TotalAmount)
	if err != nil {
		return nil, err
	}
	if s.TotalPayments > 0 {
		s.SuccessRate = float64(s.CompletedPayments) / float64(s.TotalPayments) * 100
	}
	return &s, nil
}

func (r *PostgresAnalyticsRepository) GetProviderStats(ctx context.Context, merchantID uuid.UUID, from, to time.Time) ([]*analytics.ProviderStats, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT p.provider_id, pr.name, COUNT(*), SUM(CASE WHEN p.status='completed' THEN 1 ELSE 0 END), SUM(CASE WHEN p.status='failed' THEN 1 ELSE 0 END), COALESCE(SUM(p.amount),0)
		FROM payments p LEFT JOIN providers pr ON p.provider_id=pr.id WHERE p.merchant_id=$1 AND p.created_at BETWEEN $2 AND $3 AND p.provider_id != '00000000-0000-0000-0000-000000000000' GROUP BY p.provider_id, pr.name`, merchantID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*analytics.ProviderStats
	for rows.Next() {
		var s analytics.ProviderStats
		var name *string
		rows.Scan(&s.ProviderID, &name, &s.TotalCharges, &s.SuccessCount, &s.FailureCount, &s.TotalAmount)
		if name != nil {
			s.ProviderName = *name
		}
		if s.TotalCharges > 0 {
			s.SuccessRate = float64(s.SuccessCount) / float64(s.TotalCharges) * 100
		}
		result = append(result, &s)
	}
	return result, nil
}
