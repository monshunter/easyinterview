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
		"018f2a40-0000-7000-9000-0000000000e1", // outbox id
	)

	body := api.ImportTargetJobRequest{
		RawText:        "Backend Engineer needed.",
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
		RawText:        "x",
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
		RawText:        "x",
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

	t.Run("import blank rawText", func(t *testing.T) {
		h, _ := newWiredHandler(t)
		body, _ := json.Marshal(api.ImportTargetJobRequest{
			RawText:        " \n\t ",
			TargetLanguage: "en",
			ResumeId:       "018f2a40-0000-7000-9000-0000000000r1",
		})
		req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(body))
		req.Header.Set(targetjob.IdempotencyKeyHeader, "key-1")
		req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
		rec := httptest.NewRecorder()
		h.ImportTargetJob(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeValidationFailed, false)
	})

	t.Run("import rejects removed source wrapper", func(t *testing.T) {
		h, store := newWiredHandler(t)
		body := []byte(`{"rawText":"Backend Engineer JD","targetLanguage":"en","resumeId":"resume-1","source":{"type":"url","url":"https://example.test/job"}}`)
		req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(body))
		req.Header.Set(targetjob.IdempotencyKeyHeader, "key-source-wrapper")
		req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
		rec := httptest.NewRecorder()
		h.ImportTargetJob(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertGeneratedErrorEnvelope(t, rec, sharederrors.CodeValidationFailed, false)
		if store.callCount != 0 {
			t.Fatalf("removed source wrapper must not reach store: calls=%d", store.callCount)
		}
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
		if rec.Code != http.StatusConflict {
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

func TestHandler_GetAndListTargetJobs_ReturnPracticeProgressWithWireParity(t *testing.T) {
	h, store := newWiredHandler(t)
	now := time.Date(2026, 7, 12, 13, 0, 0, 0, time.UTC)
	rec := targetjob.TargetJobRecord{
		ID:                  "018f2a40-0000-7000-9000-0000000000a1",
		UserID:              "user-1",
		Status:              sharedtypes.TargetJobStatusInterviewing,
		AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
		Title:               "Backend Engineer",
		CompanyName:         "Acme",
		TargetLanguage:      "en",
		Summary:             threeRoundSummaryJSON(),
		PracticeFactsLoaded: true,
		CompletedRoundFacts: []targetjob.PracticeRoundFact{
			{RoundID: "round-1-hr", RoundSequence: 1},
		},
		ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{
			{PlanID: "018f2a40-0000-7000-9000-0000000000p2", RoundID: "round-2-technical", RoundSequence: 2, CreatedAt: now},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.getRecord = rec
	store.listResult = targetjob.ListResult{Items: []targetjob.TargetJobRecord{rec}}

	withUser := func(req *http.Request) *http.Request {
		return req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
	}
	getRecorder := httptest.NewRecorder()
	h.GetTargetJob(getRecorder, withUser(httptest.NewRequest(http.MethodGet, "/targets/"+rec.ID, nil)), rec.ID)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status = %d, body=%s", getRecorder.Code, getRecorder.Body.String())
	}
	listRecorder := httptest.NewRecorder()
	h.ListTargetJobs(listRecorder, withUser(httptest.NewRequest(http.MethodGet, "/targets", nil)))
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status = %d, body=%s", listRecorder.Code, listRecorder.Body.String())
	}

	var detail api.TargetJob
	if err := json.Unmarshal(getRecorder.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	var list api.PaginatedTargetJob
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("list items = %d, want 1", len(list.Items))
	}
	for name, got := range map[string]api.TargetJob{"get": detail, "list": list.Items[0]} {
		if got.PracticeProgress == nil || got.PracticeProgress.Status != "in_progress" || got.PracticeProgress.CurrentRound == nil || got.PracticeProgress.CurrentRound.RoundId != "round-2-technical" {
			t.Fatalf("%s progress = %+v", name, got.PracticeProgress)
		}
		if got.CurrentPracticePlanId == nil || *got.CurrentPracticePlanId != "018f2a40-0000-7000-9000-0000000000p2" {
			t.Fatalf("%s current plan = %v", name, got.CurrentPracticePlanId)
		}
	}
}

func TestHandler_GetTargetJob_OmitsPracticeProgressWhenSummaryIsInvalid(t *testing.T) {
	h, store := newWiredHandler(t)
	now := time.Date(2026, 7, 12, 13, 30, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             "target-invalid-summary",
		UserID:         "user-1",
		Status:         sharedtypes.TargetJobStatusInterviewing,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          "Backend Engineer",
		TargetLanguage: "en",
		Summary:        json.RawMessage(`{"interviewRounds":[]}`),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	req := httptest.NewRequest(http.MethodGet, "/targets/target-invalid-summary", nil)
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
	recorder := httptest.NewRecorder()
	h.GetTargetJob(recorder, req, "target-invalid-summary")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, exists := raw["practiceProgress"]; exists {
		t.Fatalf("invalid summary must omit practiceProgress: %s", recorder.Body.String())
	}
	if _, exists := raw["currentPracticePlanId"]; exists {
		t.Fatalf("invalid summary must omit currentPracticePlanId: %s", recorder.Body.String())
	}
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
