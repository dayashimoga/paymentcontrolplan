package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apppay "github.com/paymentbridge/pcp/internal/application/payment"
	appwebhook "github.com/paymentbridge/pcp/internal/application/webhook"
	"github.com/paymentbridge/pcp/internal/domain/audit"
	"github.com/paymentbridge/pcp/internal/domain/common"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"go.uber.org/zap"
)

// PaymentHandler handles HTTP requests for payment processing.
type PaymentHandler struct {
	service        *apppay.Service
	auditService   *audit.Service
	webhookService *appwebhook.Service
	logger         *zap.Logger
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(service *apppay.Service, auditService *audit.Service, webhookService *appwebhook.Service, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		service:        service,
		auditService:   auditService,
		webhookService: webhookService,
		logger:         logger,
	}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	merchant := middleware.MerchantFromContext(r.Context())
	if merchant == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req dto.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	p, err := h.service.Create(r.Context(), apppay.CreateInput{
		MerchantID: merchant.ID, Amount: req.Amount, Currency: req.Currency,
		Description: req.Description, Metadata: req.Metadata, IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.logger.Info("payment created", zap.String("payment_id", p.ID.String()), zap.String("status", string(p.Status)))

	// Audit log
	h.logAudit(r, "payment", p.ID, merchant.ID, "create", map[string]interface{}{
		"amount": p.Amount, "currency": p.Currency, "status": string(p.Status),
	})

	// Enqueue webhook for payment status notification
	h.enqueueWebhook(r, merchant.ID, p.ID, merchant.WebhookURL, "payment."+string(p.Status), map[string]interface{}{
		"payment_id": p.ID, "status": string(p.Status), "amount": p.Amount, "currency": p.Currency,
	})

	respondJSON(w, http.StatusCreated, dto.ToPaymentResponse(p))
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "invalid UUID")
		return
	}
	p, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.ToPaymentResponse(p))
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	merchant := middleware.MerchantFromContext(r.Context())
	if merchant == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	payments, total, err := h.service.List(r.Context(), merchant.ID, offset, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}
	data := make([]dto.PaymentResponse, 0, len(payments))
	for _, p := range payments {
		data = append(data, dto.ToPaymentResponse(p))
	}
	if limit <= 0 { limit = 20 }
	respondJSON(w, http.StatusOK, dto.ListPaymentsResponse{Data: data, Total: total, Offset: offset, Limit: limit})
}

func (h *PaymentHandler) Refund(w http.ResponseWriter, r *http.Request) {
	merchant := middleware.MerchantFromContext(r.Context())

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "invalid UUID")
		return
	}
	p, err := h.service.Refund(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Audit log
	actorID := id // fallback if no merchant context
	if merchant != nil {
		actorID = merchant.ID
	}
	h.logAudit(r, "payment", id, actorID, "refund", map[string]interface{}{
		"amount": p.Amount, "currency": p.Currency, "status": string(p.Status),
	})

	// Enqueue webhook for refund notification
	if merchant != nil {
		h.enqueueWebhook(r, merchant.ID, p.ID, merchant.WebhookURL, "payment.refunded", map[string]interface{}{
			"payment_id": p.ID, "status": string(p.Status), "amount": p.Amount, "currency": p.Currency,
		})
	}

	respondJSON(w, http.StatusOK, dto.ToPaymentResponse(p))
}

func (h *PaymentHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		respondError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, common.ErrConflict):
		respondError(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, common.ErrInvalidInput):
		respondError(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		h.logger.Error("payment error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
	}
}

// logAudit records an audit log entry. Failures are logged but do not affect the response.
func (h *PaymentHandler) logAudit(r *http.Request, entityType string, entityID, actorID uuid.UUID, action string, changes map[string]interface{}) {
	if h.auditService == nil {
		return
	}
	if err := h.auditService.Log(r.Context(), entityType, entityID, actorID, action, changes, r.RemoteAddr, r.UserAgent()); err != nil {
		h.logger.Warn("audit log failed", zap.Error(err))
	}
}

// enqueueWebhook enqueues a webhook delivery if the merchant has a webhook URL.
func (h *PaymentHandler) enqueueWebhook(r *http.Request, merchantID, paymentID uuid.UUID, webhookURL, eventType string, payload interface{}) {
	if h.webhookService == nil || webhookURL == "" {
		return
	}
	if err := h.webhookService.Enqueue(r.Context(), merchantID, paymentID, webhookURL, eventType, payload); err != nil {
		h.logger.Warn("webhook enqueue failed", zap.Error(err))
	}
}

