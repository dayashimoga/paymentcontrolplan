package merchant

import (
	"fmt"

	"github.com/paymentbridge/pcp/internal/domain/common"
)

// Domain errors for the Merchant bounded context.
// These wrap common sentinel errors so callers can use errors.Is()
// for classification while retaining context-specific messages.
var (
	// ErrMerchantNotFound indicates the requested merchant does not exist.
	ErrMerchantNotFound = fmt.Errorf("merchant: %w", common.ErrNotFound)

	// ErrDuplicateMerchant indicates a merchant with the same name or API key already exists.
	ErrDuplicateMerchant = fmt.Errorf("merchant: %w", common.ErrConflict)

	// ErrInvalidName indicates the merchant name is empty or exceeds length constraints.
	ErrInvalidName = fmt.Errorf("merchant name is invalid: %w", common.ErrInvalidInput)

	// ErrInvalidStatus indicates the provided status is not a recognized value.
	ErrInvalidStatus = fmt.Errorf("merchant status is invalid: %w", common.ErrInvalidInput)
)
