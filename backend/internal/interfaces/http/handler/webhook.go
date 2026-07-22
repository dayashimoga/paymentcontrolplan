package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	appwebhook "github.com/paymentbridge/pcp/internal/application/webhook"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"go.uber.org/zap"
)

// WebhookHandler handles HTTP requests for webhook management.
type WebhookHandler struct {
	service *appwebhook.Service
	logger  *zap.Logger
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(service *appwebhook.Service, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{service: service, logger: logger}
}

// ListByPayment handles GET /api/v1/payments/{id}/webhooks
func (h *WebhookHandler) ListByPayment(w http.ResponseWriter, r *http.Request) {
	merchant := middleware.MerchantFromContext(r.Context())
	if merchant == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	paymentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "invalid payment UUID")
		return
	}

	webhooks, err := h.service.ListByPayment(r.Context(), paymentID)
	if err != nil {
		h.logger.Error("failed to list webhooks", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "failed to fetch webhooks")
		return
	}

	respondJSON(w, http.StatusOK, dto.ToWebhookListResponse(webhooks))
}

// GetByID handles GET /api/v1/webhooks/{id}
func (h *WebhookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "invalid webhook UUID")
		return
	}

	wh, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "webhook not found")
		return
	}

	respondJSON(w, http.StatusOK, dto.ToWebhookResponse(wh))
}

// ProcessPending handles POST /api/v1/webhooks/process (admin endpoint)
func (h *WebhookHandler) ProcessPending(w http.ResponseWriter, r *http.Request) {
	batchSize, _ := strconv.Atoi(r.URL.Query().Get("batch_size"))
	if batchSize <= 0 {
		batchSize = 50
	}

	processed, err := h.service.ProcessPending(r.Context(), batchSize)
	if err != nil {
		h.logger.Error("webhook processing failed", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "processing failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"processed": processed,
		"message":   "webhook delivery batch processed",
	})
}
