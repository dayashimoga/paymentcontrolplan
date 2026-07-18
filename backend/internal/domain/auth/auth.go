// Package auth defines authentication domain types and the token service port.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Claims represents the validated claims from an authentication token.
type Claims struct {
	MerchantID   uuid.UUID
	MerchantName string
	ExpiresAt    time.Time
	IssuedAt     time.Time
	Issuer       string
}

// TokenService is the port for token generation and validation.
type TokenService interface {
	GenerateToken(ctx context.Context, merchantID uuid.UUID, merchantName string) (string, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
}
