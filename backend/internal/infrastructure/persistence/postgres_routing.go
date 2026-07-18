package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paymentbridge/pcp/internal/domain/routing"
)

type PostgresRoutingRuleRepository struct{ pool *pgxpool.Pool }

func NewPostgresRoutingRuleRepository(pool *pgxpool.Pool) *PostgresRoutingRuleRepository {
	return &PostgresRoutingRuleRepository{pool: pool}
}

func (r *PostgresRoutingRuleRepository) GetByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*routing.Rule, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,merchant_id,provider_id,priority,weight,currency,min_amount,max_amount,is_active FROM routing_rules WHERE merchant_id=$1 AND is_active=true ORDER BY priority`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []*routing.Rule
	for rows.Next() {
		var rule routing.Rule
		if err := rows.Scan(&rule.ID, &rule.MerchantID, &rule.ProviderID, &rule.Priority, &rule.Weight, &rule.Currency, &rule.MinAmount, &rule.MaxAmount, &rule.IsActive); err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func (r *PostgresRoutingRuleRepository) Create(ctx context.Context, rule *routing.Rule) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO routing_rules (id,merchant_id,provider_id,priority,weight,currency,min_amount,max_amount,is_active) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rule.ID, rule.MerchantID, rule.ProviderID, rule.Priority, rule.Weight, rule.Currency, rule.MinAmount, rule.MaxAmount, rule.IsActive)
	return err
}

func (r *PostgresRoutingRuleRepository) Update(ctx context.Context, rule *routing.Rule) error {
	_, err := r.pool.Exec(ctx, `UPDATE routing_rules SET priority=$2,weight=$3,currency=$4,min_amount=$5,max_amount=$6,is_active=$7 WHERE id=$1`,
		rule.ID, rule.Priority, rule.Weight, rule.Currency, rule.MinAmount, rule.MaxAmount, rule.IsActive)
	return err
}

func (r *PostgresRoutingRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM routing_rules WHERE id=$1`, id)
	return err
}
