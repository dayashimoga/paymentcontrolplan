package handler

import (
	"net/http"
	"strconv"

	apprec "github.com/paymentbridge/pcp/internal/application/reconciliation"
	domrec "github.com/paymentbridge/pcp/internal/domain/reconciliation"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
	"go.uber.org/zap"
)

// ReconciliationHandler handles HTTP requests for payment reconciliation.
type ReconciliationHandler struct {
	service *domrec.Service
	worker  *apprec.Worker
	logger  *zap.Logger
}

// NewReconciliationHandler creates a new reconciliation handler.
func NewReconciliationHandler(service *domrec.Service, worker *apprec.Worker, logger *zap.Logger) *ReconciliationHandler {
	return &ReconciliationHandler{
		service: service,
		worker:  worker,
		logger:  logger,
	}
}

// ListUnmatched handles GET /api/v1/reconciliation/unmatched
func (h *ReconciliationHandler) ListUnmatched(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	records, total, err := h.service.ListUnmatched(r.Context(), offset, limit)
	if err != nil {
		h.logger.Error("failed to list unmatched reconciliation records", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "failed to fetch reconciliation records")
		return
	}

	respondJSON(w, http.StatusOK, dto.ToReconciliationListResponse(records, total, offset, limit))
}

// Run handles POST /api/v1/reconciliation/run (manually trigger a batch reconciliation run)
func (h *ReconciliationHandler) Run(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	processed, err := h.worker.RunBatch(r.Context(), limit)
	if err != nil {
		h.logger.Error("reconciliation batch run failed", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "reconciliation batch failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"processed": processed,
		"message":   "reconciliation run completed",
	})
}
