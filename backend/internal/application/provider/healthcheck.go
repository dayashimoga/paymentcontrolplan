// Package provider implements provider health monitoring worker.
package provider

import (
	"context"
	"time"

	domain "github.com/paymentbridge/pcp/internal/domain/provider"
	"go.uber.org/zap"
)

// HealthMonitor checks the health of registered provider gateways.
type HealthMonitor struct {
	service *Service
	logger  *zap.Logger
}

// NewHealthMonitor creates a new health monitor service.
func NewHealthMonitor(service *Service, logger *zap.Logger) *HealthMonitor {
	return &HealthMonitor{
		service: service,
		logger:  logger,
	}
}

// HealthStatus represents the result of a health check for a provider.
type HealthStatus struct {
	ProviderType domain.Type `json:"provider_type"`
	Healthy      bool        `json:"healthy"`
	Error        string      `json:"error,omitempty"`
}

// CheckAll performs a health check across all registered gateways.
func (hm *HealthMonitor) CheckAll(ctx context.Context) map[domain.Type]HealthStatus {
	results := make(map[domain.Type]HealthStatus)

	for pType, gw := range hm.service.gateways {
		err := gw.HealthCheck(ctx)
		status := HealthStatus{
			ProviderType: pType,
			Healthy:      err == nil,
		}
		if err != nil {
			status.Error = err.Error()
			hm.logger.Warn("provider gateway health check failed",
				zap.String("provider_type", string(pType)),
				zap.Error(err),
			)
		} else {
			hm.logger.Debug("provider gateway health check passed",
				zap.String("provider_type", string(pType)),
			)
		}
		results[pType] = status
	}

	return results
}

// RunWorker starts a background ticker that periodically runs provider health checks.
func (hm *HealthMonitor) RunWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}

	hm.logger.Info("starting provider health check worker", zap.Duration("interval", interval))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hm.logger.Info("stopping provider health check worker")
			return
		case <-ticker.C:
			_ = hm.CheckAll(ctx)
		}
	}
}
