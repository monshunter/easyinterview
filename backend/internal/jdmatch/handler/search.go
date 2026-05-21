package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// SavedSearchStore is the slice of store the saved-search handlers
// consume.
type SavedSearchStore interface {
	ListSavedSearchesByUser(ctx context.Context, userID string) ([]store.SavedSearchRecord, error)
	CreateSavedSearch(ctx context.Context, in store.CreateSavedSearchInput) (store.SavedSearchRecord, error)
}

// SearchRunStore is the slice the searchJobs handler consumes.
type SearchRunStore interface {
	CreateSearchRun(ctx context.Context, in store.CreateSearchRunInput) (store.SearchRunRecord, error)
}

// SearchCompletedEvent is the privacy-safe B3 event emitted after a
// search run is persisted. It intentionally excludes query, filters,
// labels, and recommendation content.
type SearchCompletedEvent struct {
	UserID      string
	SearchRunID string
	ResultCount int
}

// SearchCompletedEmitter writes jd_match.search.completed to the
// outbox. cmd/api wires the SQL outbox writer; unit tests inject a
// recorder.
type SearchCompletedEmitter interface {
	EmitSearchCompleted(ctx context.Context, event SearchCompletedEvent) error
}

// SearchAI is the A3 routing slice used by searchJobs. cmd/api wires
// the real adapter that calls feature_key jd_match.search.
type SearchAI interface {
	Search(ctx context.Context, userID, query string, filters json.RawMessage) (SearchAIResult, error)
}

// SearchAIResult bundles the matched recommendation IDs (joined back
// to jd_match_recommendations by the handler) and the AI provenance.
type SearchAIResult struct {
	MatchedJobMatchIDs []string
	PromptVersion      string
	RubricVersion      string
	ModelProfileName   string
	Language           string
	FeatureFlag        string
	DataSourceVersion  string
}

// SearchTimeoutErr is returned by the AI adapter when the call exceeds
// the 30s budget; the handler maps it to 502 AI_PROVIDER_TIMEOUT.
var SearchTimeoutErr = errors.New("jdmatch search: AI provider timeout")

// SearchInvalidOutputErr is returned when the AI adapter receives a
// syntactically valid provider response that fails the jd_match.search
// output schema.
var SearchInvalidOutputErr = errors.New("jdmatch search: invalid AI output")

// SetSearch wires the saved-search + search-run + AI deps.
func (h *Handler) SetSearch(saved SavedSearchStore, runs SearchRunStore, ai SearchAI) {
	if h == nil {
		return
	}
	h.savedSearches = saved
	h.searchRuns = runs
	h.searchAI = ai
}

// SetSearchCompleted wires the B3 outbox emitter for successful
// search runs.
func (h *Handler) SetSearchCompleted(emitter SearchCompletedEmitter) {
	if h == nil {
		return
	}
	h.searchCompleted = emitter
}

// ListSavedSearches returns the user's saved_searches rows.
func (h *Handler) ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.savedSearches == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch saved-search service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	rows, err := h.savedSearches.ListSavedSearchesByUser(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err, "jdmatch saved-search list failed")
		return
	}
	items := make([]savedSearchResponse, 0, len(rows))
	for _, rec := range rows {
		items = append(items, savedSearchToDTO(rec))
	}
	writeJSON(w, http.StatusOK, struct {
		Items []savedSearchResponse `json:"items"`
	}{Items: items})
}

// CreateSavedSearch persists a new saved search; label / query / filters
// never enter log / audit / outbox per D-7.
func (h *Handler) CreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.savedSearches == nil || h.newID == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch saved-search service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body struct {
		Label   string          `json:"label"`
		Query   string          `json:"query"`
		Filters json.RawMessage `json:"filters,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Label == "" || body.Query == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "label and query are required", nil)
		return
	}
	rec, err := h.savedSearches.CreateSavedSearch(r.Context(), store.CreateSavedSearchInput{
		ID: h.newID(), UserID: userID, Label: body.Label, Query: body.Query, Filters: body.Filters,
	})
	if err != nil {
		if errors.Is(err, jdmatch.ErrValidationFailed) {
			writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "label and query are required", nil)
			return
		}
		writeServiceError(w, err, "jdmatch saved-search create failed")
		return
	}
	writeJSON(w, http.StatusOK, savedSearchToDTO(rec))
}

// SearchJobs runs a synchronous natural-language search via the A3
// AIClient (feature_key jd_match.search), persists the search-run
// audit row, and projects the matched recommendation rows back to the
// caller.
func (h *Handler) SearchJobs(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.searchAI == nil || h.searchRuns == nil || h.recReader == nil || h.newID == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch search service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body struct {
		Query   string          `json:"query"`
		Filters json.RawMessage `json:"filters,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Query == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "query is required", nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	result, err := h.searchAI.Search(ctx, userID, body.Query, body.Filters)
	if err != nil {
		if errors.Is(err, SearchTimeoutErr) || errors.Is(err, context.DeadlineExceeded) {
			writeAPIError(w, http.StatusBadGateway, sharederrors.CodeAiProviderTimeout, "search backend timed out", nil)
			return
		}
		if errors.Is(err, SearchInvalidOutputErr) {
			writeAPIError(w, http.StatusBadGateway, sharederrors.CodeAiOutputInvalid, "search backend output invalid", nil)
			return
		}
		writeAPIError(w, http.StatusBadGateway, sharederrors.CodeAiProviderTimeout, "search backend failed", nil)
		return
	}
	provenance := generationProvenanceResponse{
		PromptVersion:     result.PromptVersion,
		RubricVersion:     result.RubricVersion,
		ModelID:           result.ModelProfileName,
		Language:          result.Language,
		FeatureFlag:       result.FeatureFlag,
		DataSourceVersion: result.DataSourceVersion,
	}
	matched := make([]jobMatchRecommendationResponse, 0, len(result.MatchedJobMatchIDs))
	for _, id := range result.MatchedJobMatchIDs {
		rec, getErr := h.recReader.GetRecommendationByIDForUser(r.Context(), userID, id)
		if getErr != nil {
			if errors.Is(getErr, jdmatch.ErrNotFound) {
				continue
			}
			writeServiceError(w, getErr, "jdmatch search projection failed")
			return
		}
		matched = append(matched, recordToDTOWithProvenance(rec, provenance))
	}
	searchRunID := h.newID()
	if _, err := h.searchRuns.CreateSearchRun(r.Context(), store.CreateSearchRunInput{
		ID:                h.newID(),
		UserID:            userID,
		SearchRunID:       searchRunID,
		Query:             body.Query,
		Filters:           body.Filters,
		ResultCount:       len(matched),
		PromptVersion:     result.PromptVersion,
		RubricVersion:     result.RubricVersion,
		ModelID:           result.ModelProfileName,
		DataSourceVersion: result.DataSourceVersion,
	}); err != nil {
		writeServiceError(w, err, "jdmatch search audit failed")
		return
	}
	if h.searchCompleted != nil {
		if err := h.searchCompleted.EmitSearchCompleted(r.Context(), SearchCompletedEvent{
			UserID:      userID,
			SearchRunID: searchRunID,
			ResultCount: len(matched),
		}); err != nil {
			writeServiceError(w, err, "jdmatch search completed event failed")
			return
		}
	}
	writeJSON(w, http.StatusOK, struct {
		SearchRunID string                           `json:"searchRunId"`
		Items       []jobMatchRecommendationResponse `json:"items"`
	}{
		SearchRunID: searchRunID,
		Items:       matched,
	})
}

type savedSearchResponse struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	Query        string                 `json:"query"`
	Filters      *api.SearchJobsFilters `json:"filters,omitempty"`
	NewJobsCount *int32                 `json:"newJobsCount"`
	LastRunAt    *string                `json:"lastRunAt"`
	CreatedAt    string                 `json:"createdAt"`
}

func savedSearchToDTO(rec store.SavedSearchRecord) savedSearchResponse {
	dto := savedSearchResponse{
		ID:        rec.ID,
		Label:     rec.Label,
		Query:     rec.Query,
		CreatedAt: rec.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if len(rec.Filters) > 0 && string(rec.Filters) != "{}" {
		var filters api.SearchJobsFilters
		if err := json.Unmarshal(rec.Filters, &filters); err == nil {
			dto.Filters = &filters
		}
	}
	if rec.NewJobsCount != nil {
		v := int32(*rec.NewJobsCount)
		dto.NewJobsCount = &v
	}
	if rec.LastRunAt != nil {
		ts := rec.LastRunAt.Format("2006-01-02T15:04:05Z")
		dto.LastRunAt = &ts
	}
	return dto
}
