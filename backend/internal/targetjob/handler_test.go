package targetjob_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// targetJobOperationNames are the B2-defined operations the targetjob
// domain owns. Renaming any of these (e.g., due to B2 churn) must surface
// as a test failure here so the wiring in cmd/api/main.go stays in sync.
var targetJobOperationNames = []string{
	"ImportTargetJob",
	"ListTargetJobs",
	"GetTargetJob",
	"UpdateTargetJob",
	"ArchiveTargetJob",
}

func TestHandlerSignaturesMatchB2ServerInterface(t *testing.T) {
	apiType := reflect.TypeFor[api.ServerInterface]()
	handlerType := reflect.TypeFor[*targetjob.Handler]()
	for _, name := range targetJobOperationNames {
		apiMethod, ok := apiType.MethodByName(name)
		if !ok {
			t.Fatalf("B2 ServerInterface missing %q — Phase 0 owner contract drifted", name)
		}
		handlerMethod, ok := handlerType.MethodByName(name)
		if !ok {
			t.Fatalf("targetjob.Handler missing %q — Phase 1.3 surface incomplete", name)
		}
		// apiMethod.Type is iface method type; handlerMethod.Type includes the
		// pointer receiver as the first argument. Strip it before comparing.
		want := apiMethod.Type
		gotIn := handlerMethod.Type.NumIn() - 1
		if want.NumIn() != gotIn {
			t.Errorf("%s: in-arg count mismatch: B2=%d, handler=%d", name, want.NumIn(), gotIn)
			continue
		}
		for i := 0; i < want.NumIn(); i++ {
			if want.In(i) != handlerMethod.Type.In(i+1) {
				t.Errorf("%s: arg %d mismatch: B2=%v, handler=%v", name, i, want.In(i), handlerMethod.Type.In(i+1))
			}
		}
		if want.NumOut() != handlerMethod.Type.NumOut() {
			t.Errorf("%s: out-arg count mismatch: B2=%d, handler=%d", name, want.NumOut(), handlerMethod.Type.NumOut())
		}
	}
}

func newWiredHandler(t *testing.T, ids ...string) (*targetjob.Handler, *fakeStore) {
	t.Helper()
	store := &fakeStore{}
	idx := 0
	gen := func() string {
		if idx >= len(ids) {
			t.Fatalf("ran out of injected ids; have %d, requesting more", len(ids))
		}
		v := ids[idx]
		idx++
		return v
	}
	now := time.Date(2026, 5, 9, 20, 0, 0, 0, time.UTC)
	svc := targetjob.NewService(targetjob.ServiceOptions{
		Store:        store,
		NewID:        gen,
		Now:          func() time.Time { return now },
		DedupePepper: "test-pepper",
	})
	h := targetjob.NewHandler(targetjob.HandlerOptions{
		Service: svc,
		Session: func(ctx context.Context) (string, bool) {
			if v, ok := ctx.Value(testUserKey{}).(string); ok && v != "" {
				return v, true
			}
			return "", false
		},
	})
	return h, store
}

type testUserKey struct{}

func TestHandler_ImportTargetJob_Returns202WithTargetJobWithJob(t *testing.T) {
	h, _ := newWiredHandler(t,
		"018f2a40-0000-7000-9000-0000000000a1", // target id
		"018f2a40-0000-7000-9000-0000000000f1", // job id
		"018f2a40-0000-7000-9000-0000000000c1", // source id
		"018f2a40-0000-7000-9000-0000000000e1", // outbox id
	)

	body := api.ImportTargetJobRequest{
		Source: map[string]any{
			"type":    "manual_text",
			"rawText": "Backend Engineer needed.",
		},
		TargetLanguage: "en",
		ResumeId:       "018f2a40-0000-7000-9000-0000000000r1",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(raw))
	req.Header.Set("Idempotency-Key", "fresh-key")
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "018f2a40-0000-7000-9000-0000000000b1"))

	rec := httptest.NewRecorder()
	h.ImportTargetJob(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.TargetJobWithJob
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.TargetJobId != "018f2a40-0000-7000-9000-0000000000a1" {
		t.Fatalf("unexpected targetJobId: %s", out.TargetJobId)
	}
	if out.Job.Status != sharedtypes.JobStatusQueued || out.Job.JobType != api.JobTypeTargetImport {
		t.Fatalf("unexpected job: %+v", out.Job)
	}
}

func TestHandler_ImportTargetJob_RejectsMissingIdempotencyKey(t *testing.T) {
	h, _ := newWiredHandler(t, "a", "b", "c", "d")
	body, _ := json.Marshal(api.ImportTargetJobRequest{
		Source:         map[string]any{"type": "manual_text", "rawText": "x"},
		TargetLanguage: "en",
		ResumeId:       "018f2a40-0000-7000-9000-0000000000r1",
	})
	req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
	rec := httptest.NewRecorder()
	h.ImportTargetJob(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeValidationFailed, false)
}

func TestHandler_ImportTargetJob_RejectsMissingSession(t *testing.T) {
	h, _ := newWiredHandler(t, "a", "b", "c", "d")
	body, _ := json.Marshal(api.ImportTargetJobRequest{
		Source:         map[string]any{"type": "manual_text", "rawText": "x"},
		TargetLanguage: "en",
		ResumeId:       "018f2a40-0000-7000-9000-0000000000r1",
	})
	req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(body))
	req.Header.Set("Idempotency-Key", "k")
	rec := httptest.NewRecorder()
	h.ImportTargetJob(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeAuthUnauthorized, false)
}

func TestHandler_ErrorResponsesUseGeneratedEnvelope(t *testing.T) {
	errorBody := func() []byte {
		body, _ := json.Marshal(api.ImportTargetJobRequest{
			Source:         map[string]any{"type": "manual_text", "rawText": "x"},
			TargetLanguage: "en",
			ResumeId:       "018f2a40-0000-7000-9000-0000000000r1",
		})
		return body
	}

	t.Run("list missing session", func(t *testing.T) {
		h, _ := newWiredHandler(t)
		rec := httptest.NewRecorder()
		h.ListTargetJobs(rec, httptest.NewRequest(http.MethodGet, "/targets", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeAuthUnauthorized, false)
	})

	t.Run("get not found", func(t *testing.T) {
		h, store := newWiredHandler(t)
		store.getErr = targetjob.ErrTargetJobNotFound
		req := httptest.NewRequest(http.MethodGet, "/targets/target-1", nil)
		req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
		rec := httptest.NewRecorder()
		h.GetTargetJob(rec, req, "target-1")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeTargetJobNotFound, false)
	})

	t.Run("import source unavailable", func(t *testing.T) {
		h, store := newWiredHandler(t, "target-1", "job-1", "source-1", "outbox-1")
		store.err = &targetjob.ServiceImportError{Code: sharederrors.CodeTargetImportSourceUnavailable, Message: "source temporarily unavailable"}
		req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(errorBody()))
		req.Header.Set(targetjob.IdempotencyKeyHeader, "key-1")
		req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
		rec := httptest.NewRecorder()
		h.ImportTargetJob(rec, req)
		if rec.Code != http.StatusBadGateway {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeTargetImportSourceUnavailable, true)
	})

	t.Run("update invalid transition", func(t *testing.T) {
		h, store := newWiredHandler(t, "marker-1")
		store.updateErr = &targetjob.ServiceImportError{Code: sharederrors.CodeTargetInvalidStateTransition, Message: "transition draft -> offer is not allowed"}
		status := sharedtypes.TargetJobStatusOffer
		body, _ := json.Marshal(api.UpdateTargetJobRequest{Status: &status})
		req := httptest.NewRequest(http.MethodPatch, "/targets/target-1", bytes.NewReader(body))
		req.Header.Set(targetjob.IdempotencyKeyHeader, "key-1")
		req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
		rec := httptest.NewRecorder()
		h.UpdateTargetJob(rec, req, "target-1")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeTargetInvalidStateTransition, false)
	})

	t.Run("archive already archived", func(t *testing.T) {
		h, store := newWiredHandler(t, "marker-archive")
		store.archiveErr = targetjob.ErrTargetJobAlreadyArchived
		req := httptest.NewRequest(http.MethodPost, "/targets/target-1/archive", nil)
		req.Header.Set(targetjob.IdempotencyKeyHeader, "key-archive")
		req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
		rec := httptest.NewRecorder()
		h.ArchiveTargetJob(rec, req, "target-1")
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeTargetInvalidStateTransition, false)
	})
}

func TestHandler_ArchiveTargetJob_Returns202AndRequiresIdempotencyKey(t *testing.T) {
	h, store := newWiredHandler(t, "marker-archive")
	now := time.Date(2026, 7, 9, 13, 45, 0, 0, time.UTC)
	store.archiveResult = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "user-1",
		Status:         sharedtypes.TargetJobStatusArchived,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          "Backend Engineer",
		CompanyName:    "Acme",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "en",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	req := httptest.NewRequest(http.MethodPost, "/targets/018f2a40-0000-7000-9000-0000000000a1/archive", nil)
	req.Header.Set(targetjob.IdempotencyKeyHeader, "key-archive")
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
	rec := httptest.NewRecorder()
	h.ArchiveTargetJob(rec, req, "018f2a40-0000-7000-9000-0000000000a1")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.TargetJob
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode archive response: %v", err)
	}
	if out.Status != sharedtypes.TargetJobStatusArchived {
		t.Fatalf("status = %s, want archived", out.Status)
	}
	if store.capturedArchiveInput.DedupeMarkerID != "marker-archive" {
		t.Fatalf("dedupe marker = %q", store.capturedArchiveInput.DedupeMarkerID)
	}

	req = httptest.NewRequest(http.MethodPost, "/targets/018f2a40-0000-7000-9000-0000000000a1/archive", nil)
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
	rec = httptest.NewRecorder()
	h.ArchiveTargetJob(rec, req, "018f2a40-0000-7000-9000-0000000000a1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing key status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeValidationFailed, false)
}

func assertGeneratedErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantCode string, wantRetryable bool) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode error envelope: %v; body=%s", err, rec.Body.String())
	}
	if _, ok := raw["errors"]; ok {
		t.Fatalf("out-of-scope errors envelope must not be present: %s", rec.Body.String())
	}
	errRaw, ok := raw["error"]
	if !ok {
		t.Fatalf("generated error envelope missing error object: %s", rec.Body.String())
	}
	var errObj map[string]any
	if err := json.Unmarshal(errRaw, &errObj); err != nil {
		t.Fatalf("decode error object: %v", err)
	}
	for _, key := range []string{"code", "message", "requestId", "retryable"} {
		if _, ok := errObj[key]; !ok {
			t.Fatalf("error object missing %q: %s", key, rec.Body.String())
		}
	}
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode generated ApiErrorResponse: %v", err)
	}
	if out.Error.Code != wantCode {
		t.Fatalf("error code = %q, want %q; body=%s", out.Error.Code, wantCode, rec.Body.String())
	}
	if out.Error.Message == "" {
		t.Fatalf("error message must be populated: %s", rec.Body.String())
	}
	if out.Error.Retryable != wantRetryable {
		t.Fatalf("retryable = %v, want %v; body=%s", out.Error.Retryable, wantRetryable, rec.Body.String())
	}
}

func TestHandlerMissingServiceReturnsInternalConfigurationError(t *testing.T) {
	h := targetjob.NewHandler()
	cases := []struct {
		name string
		exec func(w http.ResponseWriter, r *http.Request)
	}{
		{"ImportTargetJob", h.ImportTargetJob},
		{"ListTargetJobs", h.ListTargetJobs},
		{"GetTargetJob", func(w http.ResponseWriter, r *http.Request) {
			h.GetTargetJob(w, r, "018f2a40-0000-7000-9000-0000000000a1")
		}},
		{"UpdateTargetJob", func(w http.ResponseWriter, r *http.Request) {
			h.UpdateTargetJob(w, r, "018f2a40-0000-7000-9000-0000000000a1")
		}},
		{"ArchiveTargetJob", func(w http.ResponseWriter, r *http.Request) {
			h.ArchiveTargetJob(w, r, "018f2a40-0000-7000-9000-0000000000a1")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.exec(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("%s: status = %d, want %d", tc.name, rec.Code, http.StatusInternalServerError)
			}
			if strings.Contains(rec.Body.String(), "NOT_IMPLEMENTED") {
				t.Fatalf("%s: missing service must not expose stale NOT_IMPLEMENTED response: %s", tc.name, rec.Body.String())
			}
		})
	}
}
