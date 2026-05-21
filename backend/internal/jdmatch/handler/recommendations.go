package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// RecommendationsReader is the read-side projection for the
// list/detail handlers.
type RecommendationsReader interface {
	ListRecommendationsByUser(ctx context.Context, userID string, filter store.ListRecommendationsFilter) (store.ListRecommendationsResult, error)
	GetRecommendationByIDForUser(ctx context.Context, userID, id string) (jdmatch.RecommendationRecord, error)
}

// RecommendationsMutator is the side-effect projection for the
// dismiss handler.
type RecommendationsMutator interface {
	MarkRecommendationDismissed(ctx context.Context, in store.MarkRecommendationDismissedInput) (jdmatch.RecommendationRecord, error)
}

// SetRecommendations wires the recommendation deps after Handler
// construction; cmd/api uses this once the runtime is fully
// composed.
func (h *Handler) SetRecommendations(reader RecommendationsReader, mutator RecommendationsMutator) {
	if h == nil {
		return
	}
	h.recReader = reader
	h.recMutator = mutator
}

// ListJobRecommendations projects the cursor-paginated list onto the
// generated JobMatchRecommendation DTO.
func (h *Handler) ListJobRecommendations(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.recReader == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch recommendations service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	q := r.URL.Query()
	pageSize := parsePageSize(q.Get("pageSize"), 20, 100)
	cursor := strings.TrimSpace(q.Get("cursor"))
	res, err := h.recReader.ListRecommendationsByUser(r.Context(), userID, store.ListRecommendationsFilter{
		PageSize: pageSize,
		Cursor:   cursor,
	})
	if err != nil {
		writeServiceError(w, err, "jdmatch recommendations list failed")
		return
	}
	items := make([]api.JobMatchRecommendation, 0, len(res.Items))
	for _, rec := range res.Items {
		items = append(items, recordToDTO(rec))
	}
	body := struct {
		Items    []api.JobMatchRecommendation `json:"items"`
		PageInfo struct {
			PageSize   int    `json:"pageSize"`
			HasMore    bool   `json:"hasMore"`
			NextCursor string `json:"nextCursor,omitempty"`
		} `json:"pageInfo"`
	}{Items: items}
	body.PageInfo.PageSize = res.PageSize
	body.PageInfo.HasMore = res.HasMore
	body.PageInfo.NextCursor = res.NextCursor
	writeJSON(w, http.StatusOK, body)
}

// GetJobRecommendation returns the detail projection of a single
// recommendation. Cross-user lookups map to 404 RESOURCE_NOT_FOUND.
func (h *Handler) GetJobRecommendation(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.recReader == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch recommendations service is not configured", nil)
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
	rec, err := h.recReader.GetRecommendationByIDForUser(r.Context(), userID, id)
	if err != nil {
		writeServiceError(w, err, "jdmatch recommendation read failed")
		return
	}
	writeJSON(w, http.StatusOK, recordToDTO(rec))
}

// MarkJobNotRelevant dismisses a recommendation. The free note lands
// in the row but never appears in log / audit / outbox per spec D-7.
func (h *Handler) MarkJobNotRelevant(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.recMutator == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch dismiss service is not configured", nil)
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
	var body struct {
		Reason   string `json:"reason"`
		FreeNote string `json:"freeNote"`
	}
	// Allow empty body — decode best-effort; errors leave fields zero.
	_ = json.NewDecoder(r.Body).Decode(&body)
	rec, err := h.recMutator.MarkRecommendationDismissed(r.Context(), store.MarkRecommendationDismissedInput{
		ID:       id,
		UserID:   userID,
		Reason:   body.Reason,
		FreeNote: body.FreeNote,
	})
	if err != nil {
		switch {
		case errors.Is(err, jdmatch.ErrAlreadyDismissed):
			writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "recommendation already dismissed", nil)
		case errors.Is(err, jdmatch.ErrNotFound):
			writeAPIError(w, http.StatusNotFound, sharederrors.CodeResourceNotFound, "recommendation not found", nil)
		default:
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "jdmatch dismiss failed", nil)
		}
		return
	}
	resp := struct {
		JobMatchID  string `json:"jobMatchId"`
		DismissedAt string `json:"dismissedAt"`
	}{
		JobMatchID: rec.ID,
	}
	if rec.DismissedAt != nil {
		resp.DismissedAt = rec.DismissedAt.Format("2006-01-02T15:04:05Z")
	}
	writeJSON(w, http.StatusOK, resp)
}

func recordToDTO(rec jdmatch.RecommendationRecord) api.JobMatchRecommendation {
	dto := api.JobMatchRecommendation{
		Id:       rec.ID,
		Title:    rec.Title,
		Company:  rec.Company,
		Location: rec.Location,
		Score:    int32(rec.Score),
		Fit: api.JobMatchFitFooter{
			Must:      int32(rec.FitMust),
			Total:     int32(rec.FitTotal),
			Plus:      int32(rec.FitPlus),
			TotalPlus: int32(rec.FitTotalPlus),
		},
		Reasons:    rec.Reasons,
		Risks:      rec.Risks,
		Highlights: rec.Highlights,
	}
	dto.CompanyTag = rec.CompanyTag
	dto.Level = rec.Level
	dto.Comp = rec.Comp
	if rec.PostedLabel != nil {
		dto.Posted = *rec.PostedLabel
	}
	dto.Seen = rec.Seen
	dto.Saved = false
	dto.SourceUrl = rec.SourceURL
	dto.SourceLabel = rec.SourceLabel
	dto.NetworkNote = rec.NetworkNote
	if rec.SimilarInterviewers != nil {
		v := int32(*rec.SimilarInterviewers)
		dto.SimilarInterviewers = &v
	}
	if rec.InterviewHypotheses != nil {
		dto.InterviewHypotheses = rec.InterviewHypotheses
	}
	provenance := api.GenerationProvenance{}
	if rec.PromptVersion != nil {
		provenance.PromptVersion = *rec.PromptVersion
	}
	if rec.RubricVersion != nil {
		provenance.RubricVersion = *rec.RubricVersion
	}
	if rec.ModelID != nil {
		provenance.ModelId = *rec.ModelID
	}
	provenance.Language = rec.Language
	provenance.FeatureFlag = rec.FeatureFlag
	provenance.DataSourceVersion = rec.DataSourceVersion
	dto.Provenance = provenance
	return dto
}

func parsePageSize(raw string, def, max int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	var n int
	for _, c := range raw {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

// extractPathParam is a minimal helper used until cmd/api wiring
// installs the path-param middleware in Phase 5.5. It reads the last
// non-empty segment of the URL path that follows the supplied name.
// Tests inject the path directly so this remains best-effort.
func extractPathParam(r *http.Request, _ string) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.TrimSpace(parts[i])
		if p != "" && !strings.Contains(p, ":") && p != "dismiss" {
			return p
		}
	}
	return ""
}
