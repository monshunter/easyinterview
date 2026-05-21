package handler

import (
	"errors"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// GetAgentScanStatus returns the latest agent_scans row for the
// authenticated user. First-time callers (no row) get an idle baseline
// per spec D-3 with last/next scan timestamps left null and message
// null.
func (h *Handler) GetAgentScanStatus(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.agentScans == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch agent-scan service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	rec, err := h.agentScans.GetLatestAgentScanForUser(r.Context(), userID)
	if err != nil && !errors.Is(err, jdmatch.ErrNotFound) {
		writeServiceError(w, err, "jdmatch agent-scan read failed")
		return
	}
	resp := api.AgentScanStatus{
		Status: api.JobMatchAgentStatusIdle,
	}
	if !errors.Is(err, jdmatch.ErrNotFound) {
		resp.Status = api.JobMatchAgentStatus(rec.Status)
		if rec.LastScanAt != nil {
			ts := rec.LastScanAt.Format("2006-01-02T15:04:05Z")
			resp.LastScanAt = &ts
		}
		if rec.NextScanAt != nil {
			ts := rec.NextScanAt.Format("2006-01-02T15:04:05Z")
			resp.NextScanAt = &ts
		}
		if rec.ErrorMessage != nil && *rec.ErrorMessage != "" {
			msg := *rec.ErrorMessage
			resp.Message = &msg
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
