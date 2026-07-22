package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	appprov "github.com/paymentbridge/pcp/internal/application/provider"
	"github.com/paymentbridge/pcp/internal/domain/audit"
	"github.com/paymentbridge/pcp/internal/domain/common"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
	"go.uber.org/zap"
)

// ProviderHandler handles HTTP requests for the provider resource.
type ProviderHandler struct {
	service      *appprov.Service
	auditService *audit.Service
	logger       *zap.Logger
}

// NewProviderHandler creates a new provider handler.
func NewProviderHandler(service *appprov.Service, auditService *audit.Service, logger *zap.Logger) *ProviderHandler {
	return &ProviderHandler{service: service, auditService: auditService, logger: logger}
}

func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	p, err := h.service.Create(r.Context(), appprov.CreateInput{
		Name: req.Name, Type: provider.Type(req.Type), Config: req.Config, Priority: req.Priority,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Audit log
	h.logAudit(r, "provider", p.ID, p.ID, "create", map[string]interface{}{"name": p.Name, "type": string(p.Type)})

	respondJSON(w, http.StatusCreated, dto.ToProviderResponse(p))
}

func (h *ProviderHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	respondJSON(w, http.StatusOK, dto.ToProviderResponse(p))
}

func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	providers, total, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}
	data := make([]dto.ProviderResponse, 0, len(providers))
	for _, p := range providers {
		data = append(data, dto.ToProviderResponse(p))
	}
	respondJSON(w, http.StatusOK, dto.ListProvidersResponse{Data: data, Total: total, Offset: offset, Limit: limit})
}

func (h *ProviderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "invalid UUID")
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	// Audit log
	h.logAudit(r, "provider", id, id, "delete", nil)

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProviderHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		respondError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, common.ErrConflict):
		respondError(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, common.ErrInvalidInput):
		respondError(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		h.logger.Error("provider error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
	}
}

// logAudit records an audit log entry. Failures are logged but do not affect the response.
func (h *ProviderHandler) logAudit(r *http.Request, entityType string, entityID, actorID uuid.UUID, action string, changes map[string]interface{}) {
	if h.auditService == nil {
		return
	}
	if err := h.auditService.Log(r.Context(), entityType, entityID, actorID, action, changes, r.RemoteAddr, r.UserAgent()); err != nil {
		h.logger.Warn("audit log failed", zap.Error(err))
	}
}

