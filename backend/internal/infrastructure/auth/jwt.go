// Package auth provides JWT token service implementation.
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	domainauth "github.com/paymentbridge/pcp/internal/domain/auth"
)

// JWTService implements the TokenService port using JWT.
type JWTService struct {
	secret     []byte
	issuer     string
	expiration time.Duration
}

// NewJWTService creates a new JWT token service.
func NewJWTService(secret, issuer string, expiration time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), issuer: issuer, expiration: expiration}
}

// GenerateToken creates a signed JWT for the given merchant.
func (s *JWTService) GenerateToken(_ context.Context, merchantID uuid.UUID, merchantName string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  merchantID.String(),
		"name": merchantName,
		"iss":  s.issuer,
		"iat":  now.Unix(),
		"exp":  now.Add(s.expiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken parses and validates a JWT, returning the extracted claims.
func (s *JWTService) ValidateToken(_ context.Context, tokenString string) (*domainauth.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	sub, _ := claims["sub"].(string)
	merchantID, err := uuid.Parse(sub)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant ID in token: %w", err)
	}

	name, _ := claims["name"].(string)

	return &domainauth.Claims{
		MerchantID:   merchantID,
		MerchantName: name,
		Issuer:       s.issuer,
	}, nil
}
