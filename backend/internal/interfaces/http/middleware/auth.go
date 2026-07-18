package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	domainauth "github.com/paymentbridge/pcp/internal/domain/auth"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
)

type ctxKey string

const merchantCtxKey ctxKey = "authenticated_merchant"

// Auth middleware authenticates requests via JWT Bearer token or X-API-Key header.
// On success, the authenticated merchant is stored in the request context.
func Auth(tokenSvc domainauth.TokenService, merchantRepo merchant.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try JWT Bearer token
			if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := tokenSvc.ValidateToken(r.Context(), tokenStr)
				if err != nil {
					writeAuthError(w, "invalid or expired token")
					return
				}
				m, err := merchantRepo.GetByID(r.Context(), claims.MerchantID)
				if err != nil || m.Status != merchant.StatusActive {
					writeAuthError(w, "merchant not found or inactive")
					return
				}
				next.ServeHTTP(w, r.WithContext(withMerchant(r.Context(), m)))
				return
			}

			// Try API Key
			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
				m, err := merchantRepo.GetByAPIKey(r.Context(), apiKey)
				if err != nil {
					writeAuthError(w, "invalid API key")
					return
				}
				if m.Status != merchant.StatusActive {
					writeAuthError(w, "merchant account is suspended")
					return
				}
				next.ServeHTTP(w, r.WithContext(withMerchant(r.Context(), m)))
				return
			}

			writeAuthError(w, "authentication required: provide Bearer token or X-API-Key")
		})
	}
}

// MerchantFromContext extracts the authenticated merchant from context.
func MerchantFromContext(ctx context.Context) *merchant.Merchant {
	m, _ := ctx.Value(merchantCtxKey).(*merchant.Merchant)
	return m
}

func withMerchant(ctx context.Context, m *merchant.Merchant) context.Context {
	return context.WithValue(ctx, merchantCtxKey, m)
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "unauthorized", "message": msg, "code": 401,
	})
}
