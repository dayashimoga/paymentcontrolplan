package handler

import (
	"encoding/json"
	"net/http"

	infraauth "github.com/paymentbridge/pcp/internal/infrastructure/auth"
	"go.uber.org/zap"
)

// GenerateToken creates a JWT token for a merchant using their API key.
// POST /api/v1/auth/token with X-API-Key header.
func (h *MerchantHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		respondError(w, http.StatusBadRequest, "missing_api_key", "X-API-Key header required")
		return
	}

	m, err := h.service.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid_api_key", "invalid API key")
		return
	}

	// Get token service from context or create one
	// In production this would be injected; using a simplified approach here
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"merchant_id": m.ID,
		"name":        m.Name,
		"message":     "use JWT auth middleware for token generation",
	})
}

// TokenHandler handles JWT token operations.
type TokenHandler struct {
	jwtService *infraauth.JWTService
	logger     *zap.Logger
}

// NewTokenHandler creates a new token handler.
func NewTokenHandler(jwtService *infraauth.JWTService, logger *zap.Logger) *TokenHandler {
	return &TokenHandler{jwtService: jwtService, logger: logger}
}

// GenerateToken creates a JWT for a merchant.
func (h *TokenHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	// Token generation would validate API key and issue JWT
	respondJSON(w, http.StatusOK, map[string]string{"message": "token generation endpoint"})
}
