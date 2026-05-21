package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// ListExperienceCards binds GET /api/v1/profiles/me/experience-cards.
// Cursor pagination uses `updated_at DESC, id DESC` stable order (spec §2.1).
func (h *Handler) ListExperienceCards(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.store == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "profile service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	rawCursor := strings.TrimSpace(r.URL.Query().Get("cursor"))
	rawPageSize := strings.TrimSpace(r.URL.Query().Get("pageSize"))
	pageSize := h.defaultPage
	if rawPageSize != "" {
		parsed, err := strconv.Atoi(rawPageSize)
		if err != nil || parsed < 1 || int32(parsed) > h.maxPageSize {
			writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "pageSize is invalid", nil)
			return
		}
		pageSize = int32(parsed)
	}
	var cursor *profile.ListCardsCursor
	if rawCursor != "" {
		parsed, err := decodeCursor(rawCursor)
		if err != nil {
			writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "cursor is invalid", nil)
			return
		}
		cursor = &parsed
	}
	result, err := h.store.ListExperienceCardsByUser(r.Context(), userID, cursor, pageSize)
	if err != nil {
		writeServiceError(w, err, "list experience cards failed")
		return
	}
	page := api.PaginatedExperienceCard{
		Items: make([]api.ExperienceCard, 0, len(result.Items)),
		PageInfo: api.PageInfo{
			NextCursor: nil,
			HasMore:    result.HasMore,
			PageSize:   int(result.PageSize),
		},
	}
	for i := range result.Items {
		card := result.Items[i]
		page.Items = append(page.Items, mapExperienceCard(&card))
	}
	if result.HasMore && result.NextCursor != "" {
		nc := result.NextCursor
		page.PageInfo.NextCursor = &nc
	}
	writeJSON(w, http.StatusOK, page)
}

type cursorPayload struct {
	UpdatedAt string `json:"u"`
	ID        string `json:"i"`
}

// encodeCursor renders an opaque base64(json) cursor. Format is private to
// this package — clients must treat it as opaque per spec §2.1.
func encodeCursor(updatedAt time.Time, id string) string {
	raw, _ := json.Marshal(cursorPayload{
		UpdatedAt: updatedAt.UTC().Format(time.RFC3339Nano),
		ID:        id,
	})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(raw string) (profile.ListCardsCursor, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return profile.ListCardsCursor{}, profile.ErrInvalidCursor
	}
	var payload cursorPayload
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return profile.ListCardsCursor{}, profile.ErrInvalidCursor
	}
	t, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
	if err != nil || strings.TrimSpace(payload.ID) == "" {
		return profile.ListCardsCursor{}, profile.ErrInvalidCursor
	}
	return profile.ListCardsCursor{UpdatedAt: t.UTC(), ID: payload.ID}, nil
}
