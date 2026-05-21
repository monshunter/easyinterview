package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// WatchlistStore is the slice of the store layer the watchlist
// handlers consume.
type WatchlistStore interface {
	ListWatchlistByUser(ctx context.Context, userID string) ([]store.WatchlistRecord, error)
	AddWatchlistItem(ctx context.Context, in store.AddWatchlistItemInput) (store.WatchlistRecord, error)
	RemoveWatchlistItem(ctx context.Context, userID, linkedJobMatchID string) (int64, error)
}

// NewID returns a fresh UUIDv7 / equivalent identifier; cmd/api wires
// the project's id factory in Phase 5.5.
type NewID func() string

// SetWatchlist wires the watchlist deps after Handler construction.
func (h *Handler) SetWatchlist(store WatchlistStore, newID NewID) {
	if h == nil {
		return
	}
	h.watchlist = store
	h.newID = newID
}

// ListWatchlist projects the joined watchlist rows onto the
// generated DTO and derives tone per spec Q-4.
func (h *Handler) ListWatchlist(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.watchlist == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch watchlist service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	rows, err := h.watchlist.ListWatchlistByUser(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err, "jdmatch watchlist list failed")
		return
	}
	items := make([]watchlistItemResponse, 0, len(rows))
	for _, rec := range rows {
		items = append(items, watchlistRecordToDTO(rec))
	}
	writeJSON(w, http.StatusOK, struct {
		Items []watchlistItemResponse `json:"items"`
	}{Items: items})
}

// AddToWatchlist appends a recommendation to the watchlist. On UNIQUE
// conflict it returns the existing row per spec C-6.
func (h *Handler) AddToWatchlist(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.watchlist == nil || h.newID == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch watchlist service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body struct {
		JobMatchID string  `json:"jobMatchId"`
		Label      *string `json:"label,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.JobMatchID == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "jobMatchId is required", nil)
		return
	}
	rec, err := h.watchlist.AddWatchlistItem(r.Context(), store.AddWatchlistItemInput{
		ID:               h.newID(),
		UserID:           userID,
		LinkedJobMatchID: body.JobMatchID,
		Label:            body.Label,
	})
	if err != nil {
		if errors.Is(err, jdmatch.ErrNotFound) {
			writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "linked recommendation not found", nil)
			return
		}
		writeServiceError(w, err, "jdmatch watchlist add failed")
		return
	}
	writeJSON(w, http.StatusOK, watchlistRecordToDTO(rec))
}

// RemoveFromWatchlist deletes a watchlist row by linked job match id.
// Returns 204 on success, 404 when no row matches (cross-user
// included).
func (h *Handler) RemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.watchlist == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch watchlist service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	id := extractPathParam(r, "jobMatchId")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "jobMatchId is required", nil)
		return
	}
	count, err := h.watchlist.RemoveWatchlistItem(r.Context(), userID, id)
	if err != nil {
		writeServiceError(w, err, "jdmatch watchlist remove failed")
		return
	}
	if count == 0 {
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "watchlist entry not found", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type watchlistItemResponse struct {
	ID               string                `json:"id"`
	LinkedJobMatchID string                `json:"linkedJobMatchId"`
	Label            *string               `json:"label"`
	Title            string                `json:"title"`
	Company          string                `json:"company"`
	Tone             api.WatchlistItemTone `json:"tone"`
	AddedAt          string                `json:"addedAt"`
	Change           *string               `json:"change"`
}

func watchlistRecordToDTO(rec store.WatchlistRecord) watchlistItemResponse {
	addedAt := rec.AddedAt.Format("2006-01-02T15:04:05Z")
	return watchlistItemResponse{
		ID:               rec.ID,
		LinkedJobMatchID: rec.LinkedJobMatchID,
		Label:            rec.Label,
		Title:            rec.LinkedTitle,
		Company:          rec.LinkedCompany,
		Tone:             api.WatchlistItemTone(deriveTone(rec.LinkedScore)),
		AddedAt:          addedAt,
		Change:           rec.ChangeNote,
	}
}

// deriveTone implements spec Q-4: score >= 80 -> ok / 50-79 -> warn
// / <50 -> muted. Falls back to muted on negative or unknown.
func deriveTone(score int) string {
	switch {
	case score >= 80:
		return "ok"
	case score >= 50:
		return "warn"
	default:
		return "muted"
	}
}
