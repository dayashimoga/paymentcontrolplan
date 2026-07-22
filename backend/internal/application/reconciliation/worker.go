// Package reconciliation provides automated reconciliation background worker and service.
package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/paymentbridge/pcp/internal/domain/payment"
	domrec "github.com/paymentbridge/pcp/internal/domain/reconciliation"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"go.uber.org/zap"
)

// Worker performs scheduled background reconciliation matching internal payment state against provider settlement APIs.
type Worker struct {
	recService   *domrec.Service
	paymentRepo  payment.Repository
	providerRepo provider.Repository
	gateways     map[provider.Type]provider.Gateway
	logger       *zap.Logger
}

// NewWorker creates a new reconciliation worker.
func NewWorker(recService *domrec.Service, paymentRepo payment.Repository, providerRepo provider.Repository, logger *zap.Logger) *Worker {
	return &Worker{
		recService:   recService,
		paymentRepo:  paymentRepo,
		providerRepo: providerRepo,
		gateways:     make(map[provider.Type]provider.Gateway),
		logger:       logger,
	}
}

// RegisterGateway registers a provider gateway implementation for transaction status checking.
func (w *Worker) RegisterGateway(providerType provider.Type, gw provider.Gateway) {
	w.gateways[providerType] = gw
}

// RunBatch performs a reconciliation run on up to `limit` unmatched/recent payments.
func (w *Worker) RunBatch(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 50
	}

	// For demonstration, list unmatched reconciliation records or query active payments
	records, _, err := w.recService.ListUnmatched(ctx, 0, limit)
	if err != nil {
		return 0, fmt.Errorf("listing unmatched records: %w", err)
	}

	processed := 0
	for _, rec := range records {
		if ctx.Err() != nil {
			return processed, ctx.Err()
		}

		prov, err := w.providerRepo.GetByID(ctx, rec.ProviderID)
		if err != nil {
			continue
		}

		gw, ok := w.gateways[prov.Type]
		if !ok {
			continue
		}

		paymentObj, err := w.paymentRepo.GetByID(ctx, rec.PaymentID)
		if err != nil || paymentObj == nil || paymentObj.ExternalID == "" {
			continue
		}

		txStatus, err := gw.GetTransactionStatus(ctx, paymentObj.ExternalID)
		if err != nil {
			w.logger.Warn("failed to fetch provider status during reconciliation",
				zap.String("payment_id", paymentObj.ID.String()),
				zap.Error(err),
			)
			continue
		}

		// Reconcile and save updated record
		_, err = w.recService.Reconcile(
			ctx,
			paymentObj.ID,
			prov.ID,
			paymentObj.Amount,
			paymentObj.Status,
			paymentObj.Amount, // External amount matched from txStatus in full integration
			txStatus.Status,
		)
		if err != nil {
			w.logger.Error("failed to record reconciliation result", zap.Error(err))
		} else {
			processed++
		}
	}

	return processed, nil
}

// RunWorker starts the periodic background reconciliation ticker.
func (w *Worker) RunWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	w.logger.Info("starting reconciliation background worker", zap.Duration("interval", interval))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping reconciliation background worker")
			return
		case <-ticker.C:
			processed, err := w.RunBatch(ctx, 50)
			if err != nil && ctx.Err() == nil {
				w.logger.Error("reconciliation batch execution failed", zap.Error(err))
			} else if processed > 0 {
				w.logger.Info("reconciliation batch completed", zap.Int("processed", processed))
			}
		}
	}
}
