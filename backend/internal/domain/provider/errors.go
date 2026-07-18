package provider

import (
	"fmt"
	"github.com/paymentbridge/pcp/internal/domain/common"
)

var (
	ErrProviderNotFound    = fmt.Errorf("provider: %w", common.ErrNotFound)
	ErrDuplicateProvider   = fmt.Errorf("provider: %w", common.ErrConflict)
	ErrInvalidProviderName = fmt.Errorf("provider name is invalid: %w", common.ErrInvalidInput)
	ErrInvalidProviderType = fmt.Errorf("provider type is invalid: %w", common.ErrInvalidInput)
	ErrProviderUnavailable = fmt.Errorf("provider is unavailable: %w", common.ErrInternal)
	ErrChargeDeclined      = fmt.Errorf("charge was declined: %w", common.ErrInternal)
	ErrRefundFailed        = fmt.Errorf("refund failed: %w", common.ErrInternal)
)
