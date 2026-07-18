package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/paymentbridge/pcp/internal/application/analytics"
	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"go.uber.org/zap"
)

// AnalyticsHandler handles analytics API requests.
type AnalyticsHandler struct {
	service *analytics.Service
	logger  *zap.Logger
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(service *analytics.Service, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, logger: logger}
}

func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	merchant := middleware.MerchantFromContext(r.Context())
	if merchant == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	from, to := parseTimeRange(r)
	summary, err := h.service.GetSummary(r.Context(), merchant.ID, from, to)
	if err != nil {
		h.logger.Error("analytics error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal_error", "failed to fetch analytics")
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (h *AnalyticsHandler) GetProviderStats(w http.ResponseWriter, r *http.Request) {
	merchant := middleware.MerchantFromContext(r.Context())
	if merchant == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	from, to := parseTimeRange(r)
	stats, err := h.service.GetProviderStats(r.Context(), merchant.ID, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "failed to fetch stats")
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func parseTimeRange(r *http.Request) (time.Time, time.Time) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -days)
	return from, to
}
