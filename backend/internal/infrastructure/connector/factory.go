package connector

import (
	"fmt"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

// NewGateway is the factory that creates the appropriate Gateway implementation
// based on provider type and configuration. This is the single point where
// new provider integrations are registered.
func NewGateway(providerType provider.Type, config map[string]string) (provider.Gateway, error) {
	switch providerType {
	case provider.TypeStripe:
		return NewStripeGateway(config), nil
	case provider.TypePayPal:
		return NewPayPalGateway(config), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
