package handler

import (
	"net/http"
)

// RejectResumeTailorSuggestion binds
// POST /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/reject.
func (h *Handler) RejectResumeTailorSuggestion(w http.ResponseWriter, r *http.Request, resumeVersionID string, suggestionID string) {
	h.decideResumeTailorSuggestion(w, r, resumeVersionID, suggestionID, false)
}
