// Package handler implements HTTP request handlers for the PCP API.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paymentbridge/pcp/internal/infrastructure/persistence"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
)

// HealthHandler handles health and readiness probe requests.
type HealthHandler struct {
	pool *pgxpool.Pool
}

// NewHealthHandler creates a new health handler with DB pool for readiness checks.
func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Health returns 200 if the service is alive.
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "pcp-api",
	})
}

// Ready returns 200 if the service and all dependencies are ready.
// GET /ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := persistence.HealthCheck(r.Context(), h.pool); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "database connection failed",
			Code:    http.StatusServiceUnavailable,
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// respondError writes a standardized error response.
func respondError(w http.ResponseWriter, status int, errType, message string) {
	respondJSON(w, status, dto.ErrorResponse{
		Error:   errType,
		Message: message,
		Code:    status,
	})
}
