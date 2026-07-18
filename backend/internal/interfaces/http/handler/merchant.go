package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	appmerchant "github.com/paymentbridge/pcp/internal/application/merchant"
	"github.com/paymentbridge/pcp/internal/domain/common"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
	"go.uber.org/zap"
)

// MerchantHandler handles HTTP requests for the merchant resource.
type MerchantHandler struct {
	service *appmerchant.Service
	logger  *zap.Logger
}

// NewMerchantHandler creates a new merchant handler.
func NewMerchantHandler(service *appmerchant.Service, logger *zap.Logger) *MerchantHandler {
	return &MerchantHandler{
		service: service,
		logger:  logger,
	}
}

// Create handles POST /api/v1/merchants
func (h *MerchantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON payload")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "name is required")
		return
	}

	m, err := h.service.Create(r.Context(), appmerchant.CreateInput{
		Name:       req.Name,
		WebhookURL: req.WebhookURL,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.logger.Info("merchant created",
		zap.String("merchant_id", m.ID.String()),
		zap.String("merchant_name", m.Name),
	)

	respondJSON(w, http.StatusCreated, dto.ToMerchantResponse(m))
}

// Get handles GET /api/v1/merchants/{id}
func (h *MerchantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return
	}

	m, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dto.ToMerchantResponse(m))
}

// List handles GET /api/v1/merchants
func (h *MerchantHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	merchants, total, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	respondJSON(w, http.StatusOK, dto.ToMerchantListResponse(merchants, total, offset, limit))
}

// Update handles PUT /api/v1/merchants/{id}
func (h *MerchantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return
	}

	var req dto.UpdateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON payload")
		return
	}

	input := appmerchant.UpdateInput{
		Name:       req.Name,
		WebhookURL: req.WebhookURL,
	}
	if req.Status != nil {
		s := merchant.Status(*req.Status)
		input.Status = &s
	}

	m, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.logger.Info("merchant updated",
		zap.String("merchant_id", m.ID.String()),
	)

	respondJSON(w, http.StatusOK, dto.ToMerchantResponse(m))
}

// Delete handles DELETE /api/v1/merchants/{id}
func (h *MerchantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	h.logger.Info("merchant deleted",
		zap.String("merchant_id", id.String()),
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleError maps domain errors to HTTP responses.
func (h *MerchantHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		respondError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, common.ErrConflict):
		respondError(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, common.ErrInvalidInput):
		respondError(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		h.logger.Error("unhandled error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}
