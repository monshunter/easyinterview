package handler

import (
	"context"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// MarketSignalsBuilder is the orchestrator dependency for
// GetMarketSignals.
type MarketSignalsBuilder func(ctx context.Context, userID string, window service.MarketSignalsWindow) (api.MarketSignalsResponse, error)

// SetMarketSignals wires the orchestrator function used by
// GetMarketSignals. cmd/api Phase 5.5 binds service.BuildMarketSignals
// with the concrete deps.
func (h *Handler) SetMarketSignals(builder MarketSignalsBuilder) {
	if h == nil {
		return
	}
	h.marketSignals = builder
}

// GetMarketSignals projects the 4-signal aggregate per spec C-11.
// Window query parameter (7d / 14d / 30d) defaults to 7d when empty.
// Invalid window strings return 422.
func (h *Handler) GetMarketSignals(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.marketSignals == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch market signals service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "7d"
	}
	if !service.IsValidMarketSignalsWindow(window) {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "window must be one of 7d / 14d / 30d", nil)
		return
	}
	resp, err := h.marketSignals(r.Context(), userID, service.MarketSignalsWindow(window))
	if err != nil {
		writeServiceError(w, err, "jdmatch market signals failed")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
