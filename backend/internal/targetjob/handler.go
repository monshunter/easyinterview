package targetjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// IdempotencyKeyHeader is the canonical name of the header callers must
// supply for the four mutating TargetJob operations (D-6).
const IdempotencyKeyHeader = "Idempotency-Key"

// SessionResolver returns the authenticated user id for the request, or an
// empty string when the request is not authenticated. Wiring is deferred to
// `cmd/api/main.go`, which adapts auth.CurrentSessionFromContext into this
// shape so the handler does not import auth directly.
type SessionResolver func(ctx context.Context) (userID string, ok bool)

// Handler binds the four B2-defined TargetJob OpenAPI operations into the
// targetjob domain.
type Handler struct {
	service *Service
	session SessionResolver
}

// HandlerOptions wires the Handler. Service is required for production
// wiring; a missing Service is treated as server misconfiguration.
type HandlerOptions struct {
	Service *Service
	Session SessionResolver
}

// NewHandler returns a Handler.
func NewHandler(opts ...HandlerOptions) *Handler {
	h := &Handler{}
	if len(opts) > 0 {
		h.service = opts[0].Service
		h.session = opts[0].Session
	}
	return h
}

// ImportTargetJob is the POST /targets/import binding.
func (h *Handler) ImportTargetJob(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeTargetImportFailed, "targetjob service is not configured")
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required")
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get(IdempotencyKeyHeader))
	if idempotencyKey == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required")
		return
	}
	var body api.ImportTargetJobRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed")
		return
	}

	resp, err := h.service.ImportTargetJob(r.Context(), ImportRequest{
		UserID:          userID,
		IdempotencyKey:  idempotencyKey,
		TargetLanguage:  body.TargetLanguage,
		TitleHint:       derefString(body.TitleHint),
		CompanyNameHint: derefString(body.CompanyNameHint),
		Source:          body.Source,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := api.TargetJobWithJob{
		TargetJobId: resp.TargetJobID,
		Job:         resp.Job,
	}
	writeJSON(w, http.StatusAccepted, out)
}

// ListTargetJobs is the GET /targets binding.
func (h *Handler) ListTargetJobs(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeTargetImportFailed, "targetjob service is not configured")
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required")
		return
	}
	q := r.URL.Query()
	in := ListRequest{
		UserID:      userID,
		SearchQuery: q.Get("q"),
		Cursor:      q.Get("cursor"),
	}
	if statusStr := strings.TrimSpace(q.Get("status")); statusStr != "" {
		v := lifecycleStatusFrom(statusStr)
		in.Status = &v
	}
	if analysisStr := strings.TrimSpace(q.Get("analysisStatus")); analysisStr != "" {
		v := analysisStatusFrom(analysisStr)
		in.AnalysisStatus = &v
	}
	if pageSizeStr := strings.TrimSpace(q.Get("pageSize")); pageSizeStr != "" {
		var n int32
		if _, err := fmtSscan(pageSizeStr, &n); err == nil {
			in.PageSize = n
		}
	}
	res, err := h.service.ListTargetJobs(r.Context(), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// GetTargetJob is the GET /targets/{targetJobId} binding.
func (h *Handler) GetTargetJob(w http.ResponseWriter, r *http.Request, targetJobID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeTargetImportFailed, "targetjob service is not configured")
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required")
		return
	}
	res, err := h.service.GetTargetJob(r.Context(), userID, targetJobID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// UpdateTargetJob is the PATCH /targets/{targetJobId} binding.
func (h *Handler) UpdateTargetJob(w http.ResponseWriter, r *http.Request, targetJobID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeTargetImportFailed, "targetjob service is not configured")
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required")
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get(IdempotencyKeyHeader))
	if idempotencyKey == "" {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required")
		return
	}
	var body api.UpdateTargetJobRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed")
		return
	}
	res, err := h.service.UpdateTargetJob(r.Context(), UpdateRequest{
		UserID:          userID,
		TargetJobID:     targetJobID,
		IdempotencyKey:  idempotencyKey,
		Status:          body.Status,
		LocationText:    body.LocationText,
		Notes:           body.Notes,
		TitleHint:       body.TitleHint,
		CompanyNameHint: body.CompanyNameHint,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func lifecycleStatusFrom(s string) sharedtypes.TargetJobStatus {
	return sharedtypes.TargetJobStatus(s)
}

func analysisStatusFrom(s string) sharedtypes.TargetJobParseStatus {
	return sharedtypes.TargetJobParseStatus(s)
}

// fmtSscan is a tiny wrapper to keep the handler import list small.
func fmtSscan(in string, target *int32) (int, error) {
	var v int
	n, err := fmt.Sscanf(in, "%d", &v)
	if err == nil {
		*target = int32(v)
	}
	return n, err
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h.session == nil {
		return "", false
	}
	return h.session(r.Context())
}

// targetJobServerSurface mirrors the B2 generated ServerInterface for the
// four TargetJob operations exactly. handler_test.go pins this against the
// real ServerInterface via reflection so any B2 wire-shape change surfaces
// here as a compile error or test failure.
type targetJobServerSurface interface {
	ImportTargetJob(w http.ResponseWriter, r *http.Request)
	ListTargetJobs(w http.ResponseWriter, r *http.Request)
	GetTargetJob(w http.ResponseWriter, r *http.Request, targetJobID string)
	UpdateTargetJob(w http.ResponseWriter, r *http.Request, targetJobID string)
}

var _ targetJobServerSurface = (*Handler)(nil)

func writeJSON(w http.ResponseWriter, status int, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeTargetImportFailed, "response encoding failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	envelope := api.ApiErrorResponse{
		Error: api.ApiError{
			Code:      code,
			Message:   message,
			RequestID: "",
			Retryable: meta.Retryable,
		},
	}
	raw, _ := json.Marshal(envelope)
	_, _ = w.Write(raw)
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrIdempotencyKeyRequired) {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required")
		return
	}
	var svcErr *ServiceImportError
	if errors.As(err, &svcErr) {
		status := http.StatusBadRequest
		switch svcErr.Code {
		case sharederrors.CodeTargetJobNotFound:
			status = http.StatusNotFound
		case sharederrors.CodeTargetImportSourceUnavailable:
			status = http.StatusBadGateway
		}
		writeAPIError(w, status, svcErr.Code, svcErr.Message)
		return
	}
	writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeTargetImportFailed, "internal error")
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
