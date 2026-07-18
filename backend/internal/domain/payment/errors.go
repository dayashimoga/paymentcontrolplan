package payment

import (
	"fmt"
	"github.com/paymentbridge/pcp/internal/domain/common"
)

var (
	ErrPaymentNotFound   = fmt.Errorf("payment: %w", common.ErrNotFound)
	ErrInvalidAmount     = fmt.Errorf("payment amount must be positive: %w", common.ErrInvalidInput)
	ErrInvalidCurrency   = fmt.Errorf("payment currency must be 3-letter ISO code: %w", common.ErrInvalidInput)
	ErrInvalidMerchantID = fmt.Errorf("payment merchant ID is required: %w", common.ErrInvalidInput)
	ErrDuplicatePayment  = fmt.Errorf("payment with this idempotency key exists: %w", common.ErrConflict)
)
