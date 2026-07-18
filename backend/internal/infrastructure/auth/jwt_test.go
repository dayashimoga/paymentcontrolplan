package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	infraauth "github.com/paymentbridge/pcp/internal/infrastructure/auth"
	"github.com/stretchr/testify/assert"
)

func TestJWTService(t *testing.T) {
	ctx := context.Background()
	secret := "secret-key-123"
	issuer := "pcp-test"
	expiration := time.Hour

	svc := infraauth.NewJWTService(secret, issuer, expiration)

	merchantID := uuid.New()

	// Generate Token
	token, err := svc.GenerateToken(ctx, merchantID, "Acme Corp")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate Token
	claims, err := svc.ValidateToken(ctx, token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, merchantID, claims.MerchantID)
	assert.Equal(t, "Acme Corp", claims.MerchantName)
}

func TestJWTService_InvalidToken(t *testing.T) {
	ctx := context.Background()
	svc := infraauth.NewJWTService("secret", "issuer", time.Hour)

	_, err := svc.ValidateToken(ctx, "invalid.jwt.token")
	assert.Error(t, err)
}
