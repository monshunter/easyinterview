package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestHandlerImplementsResumeTailorSurfaces(t *testing.T) {
	var _ interface {
		RequestResumeTailor(http.ResponseWriter, *http.Request)
		GetResumeTailorRun(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestRequestResumeTailor(t *testing.T) {
	now := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)
	svc := &fakeTailorRunService{requestOut: tailorRunWithJobResponse("tailor-run-1", "job-1", now)}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.RequestResumeTailor(rec, newRequestTailorRequest(validRequestTailorBody("gap_review")))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.requestIn.UserID != "user-1" || svc.requestIn.ResumeAssetID != "asset-1" || svc.requestIn.ResumeVersionID != "version-1" || svc.requestIn.TargetJobID != "target-1" || svc.requestIn.Mode != "gap_review" || svc.requestIn.IdempotencyKey != "idem-tailor" {
		t.Fatalf("service input = %+v", svc.requestIn)
	}
	var got api.ResumeTailorRunWithJob
	decodeResponse(t, rec, &got)
	if got.TailorRunId != "tailor-run-1" || got.Job.JobType != api.JobTypeResumeTailor || got.Job.ResourceType != api.ResourceTypeResumeTailorRun {
		t.Fatalf("response = %+v", got)
	}
}

func TestRequestResumeTailorValidationAndErrors(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		err    error
		status int
		code   string
	}{
		{name: "invalid mode", body: validRequestTailorBody("unsupported"), status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "missing asset", body: `{"targetJobId":"target-1","mode":"gap_review"}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "not found", body: validRequestTailorBody("gap_review"), err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeTargetJobNotFound},
		{name: "validation", body: validRequestTailorBody("gap_review"), err: resume.ErrValidationFailed, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeTailorRunService{requestOut: tailorRunWithJobResponse("tailor-run-1", "job-1", time.Now()), requestErr: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.RequestResumeTailor(rec, newRequestTailorRequest(tc.body))

			assertAPIError(t, rec, tc.status, tc.code)
		})
	}
}

func TestRequestResumeTailorRequiresIdempotencyKey(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeTailorRunService{requestOut: tailorRunWithJobResponse("tailor-run-1", "job-1", time.Now())},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume/tailor", strings.NewReader(validRequestTailorBody("gap_review")))

	h.RequestResumeTailor(rec, req)

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestRequestResumeTailorIdempotencyReplay(t *testing.T) {
	store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
		State:          idempotency.StateReplay,
		RecordID:       "idem-rec-tailor",
		ResponseStatus: http.StatusAccepted,
		ResponseBody:   []byte(`{"tailorRunId":"tailor-run-replay","job":{"id":"job-replay","jobType":"resume_tailor","status":"queued","resourceType":"resume_tailor_run","resourceId":"tailor-run-replay","errorCode":null,"createdAt":"2026-05-18T09:00:00Z","updatedAt":"2026-05-18T09:00:00Z"}}`),
	}}
	svc := &fakeTailorRunService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{Store: store})
	wrapped := mw.Handler("resume", "requestResumeTailor", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.RequestResumeTailor))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newRequestTailorRequest(validRequestTailorBody("gap_review")))

	if rec.Code != http.StatusAccepted || rec.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
	}
	if svc.requestCalls != 0 {
		t.Fatalf("service calls on replay = %d, want 0", svc.requestCalls)
	}
}

func TestGetResumeTailorRun(t *testing.T) {
	now := time.Date(2026, 5, 18, 9, 30, 0, 0, time.UTC)
	for _, status := range []string{"queued", "generating", "ready", "failed"} {
		t.Run(status, func(t *testing.T) {
			svc := &fakeTailorRunService{getOut: tailorRunResponse("tailor-run-1", status, now)}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.GetResumeTailorRun(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resume/tailor-runs/tailor-run-1", nil), "tailor-run-1")

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if svc.getUserID != "user-1" || svc.getTailorRunID != "tailor-run-1" {
				t.Fatalf("service scope user=%q run=%q", svc.getUserID, svc.getTailorRunID)
			}
			var got api.ResumeTailorRun
			decodeResponse(t, rec, &got)
			if got.Status != status {
				t.Fatalf("run status = %q, want %q", got.Status, status)
			}
		})
	}
}

func TestGetResumeTailorRunNotFound(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeTailorRunService{getErr: resume.ErrNotFound},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.GetResumeTailorRun(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resume/tailor-runs/missing", nil), "missing")

	assertAPIError(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
}

func TestResumeTailorFixtureParity(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	requestFixture := loadResumeTailorRequestFixture(t)
	getFixture := loadResumeTailorGetFixture(t)

	for _, scenario := range []string{"default"} {
		t.Run("request "+scenario, func(t *testing.T) {
			entry := requestFixture.Scenarios[scenario]
			var out api.ResumeTailorRunWithJob
			if err := json.Unmarshal(entry.Response.Body, &out); err != nil {
				t.Fatalf("decode request fixture body: %v", err)
			}
			svc := &fakeTailorRunService{requestOut: out}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.RequestResumeTailor(rec, newRequestTailorFixtureRequest(entry))

			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
		})
	}

	t.Run("request idempotency replay", func(t *testing.T) {
		entry := requestFixture.Scenarios["idempotency-replay"]
		store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-request-tailor",
			ResponseStatus: entry.Response.Status,
			ResponseBody:   entry.Response.Body,
		}}
		svc := &fakeTailorRunService{}
		h := resumehandler.New(resumehandler.Options{
			Service: svc,
			Session: func(context.Context) (string, bool) { return "user-1", true },
		})
		mw := idempotency.New(idempotency.MiddlewareOptions{Store: store, Now: func() time.Time { return now }})
		wrapped := mw.Handler("resume", "requestResumeTailor", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.RequestResumeTailor))
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, newRequestTailorFixtureRequest(entry))

		if rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
		}
		if svc.requestCalls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", svc.requestCalls)
		}
		assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
	})

	for _, scenario := range []string{"default", "queued", "generating", "failed"} {
		t.Run("get "+scenario, func(t *testing.T) {
			entry := getFixture.Scenarios[scenario]
			var out api.ResumeTailorRun
			if err := json.Unmarshal(entry.Response.Body, &out); err != nil {
				t.Fatalf("decode get fixture body: %v", err)
			}
			svc := &fakeTailorRunService{getOut: out}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.GetResumeTailorRun(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resume/tailor-runs/"+out.Id, nil), out.Id)

			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
		})
	}
}

type fakeTailorRunService struct {
	requestCalls int
	requestIn    resume.RequestTailorRunInput
	requestOut   api.ResumeTailorRunWithJob
	requestErr   error

	getUserID      string
	getTailorRunID string
	getOut         api.ResumeTailorRun
	getErr         error
}

func (s *fakeTailorRunService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *fakeTailorRunService) RequestResumeTailor(_ context.Context, in resume.RequestTailorRunInput) (api.ResumeTailorRunWithJob, error) {
	s.requestCalls++
	s.requestIn = in
	return s.requestOut, s.requestErr
}

func (s *fakeTailorRunService) GetResumeTailorRun(_ context.Context, userID string, tailorRunID string) (api.ResumeTailorRun, error) {
	s.getUserID = userID
	s.getTailorRunID = tailorRunID
	return s.getOut, s.getErr
}

func newRequestTailorRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume/tailor", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-tailor")
	return req
}

func validRequestTailorBody(mode string) string {
	return `{"targetJobId":"target-1","resumeAssetId":"asset-1","resumeVersionId":"version-1","mode":"` + mode + `"}`
}

func tailorRunWithJobResponse(tailorRunID, jobID string, now time.Time) api.ResumeTailorRunWithJob {
	return api.ResumeTailorRunWithJob{
		TailorRunId: tailorRunID,
		Job: api.Job{
			Id:           jobID,
			JobType:      api.JobTypeResumeTailor,
			ResourceType: api.ResourceTypeResumeTailorRun,
			ResourceId:   tailorRunID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now.Format(time.RFC3339),
			UpdatedAt:    now.Format(time.RFC3339),
		},
	}
}

func tailorRunResponse(tailorRunID, status string, now time.Time) api.ResumeTailorRun {
	run := api.ResumeTailorRun{
		Id:            tailorRunID,
		Status:        status,
		TargetJobId:   "target-1",
		ResumeAssetId: "asset-1",
		Suggestions:   []api.ResumeTailorBulletSuggestion{},
		CreatedAt:     now.Format(time.RFC3339),
		UpdatedAt:     now.Format(time.RFC3339),
	}
	if status == "ready" {
		run.MatchSummary = &api.ResumeTailorMatchSummary{Strengths: []string{"Strong systems evidence"}, Gaps: []string{"Add edge runtime detail"}}
		run.Suggestions = []api.ResumeTailorBulletSuggestion{{
			OriginalBullet:  "Led design-system migration.",
			SuggestedBullet: "Led design-system migration across 12 teams.",
			Reason:          "Adds scope.",
		}}
		run.Provenance = &api.GenerationProvenance{
			PromptVersion:     "resume_tailor.v2",
			RubricVersion:     "not_applicable",
			ModelId:           "model-profile:contract.default",
			Language:          "zh-CN",
			FeatureFlag:       "none",
			DataSourceVersion: "target_job.v17",
		}
	}
	return run
}

type resumeTailorFixtureEntry struct {
	Request struct {
		Headers map[string]string `json:"headers"`
		Body    json.RawMessage   `json:"body"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

type resumeTailorFixture struct {
	Scenarios map[string]resumeTailorFixtureEntry `json:"scenarios"`
}

func loadResumeTailorRequestFixture(t *testing.T) resumeTailorFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "ResumeTailor", "requestResumeTailor.json"))
	if err != nil {
		t.Fatalf("read requestResumeTailor fixture: %v", err)
	}
	var fixture resumeTailorFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode requestResumeTailor fixture: %v", err)
	}
	for _, scenario := range []string{"default", "idempotency-replay"} {
		if _, ok := fixture.Scenarios[scenario]; !ok {
			t.Fatalf("request fixture missing scenario %q", scenario)
		}
	}
	return fixture
}

func loadResumeTailorGetFixture(t *testing.T) resumeTailorFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "ResumeTailor", "getResumeTailorRun.json"))
	if err != nil {
		t.Fatalf("read getResumeTailorRun fixture: %v", err)
	}
	var fixture resumeTailorFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode getResumeTailorRun fixture: %v", err)
	}
	for _, scenario := range []string{"default", "queued", "generating", "failed"} {
		if _, ok := fixture.Scenarios[scenario]; !ok {
			t.Fatalf("get fixture missing scenario %q", scenario)
		}
	}
	return fixture
}

func newRequestTailorFixtureRequest(entry resumeTailorFixtureEntry) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume/tailor", strings.NewReader(string(entry.Request.Body)))
	for key, value := range entry.Request.Headers {
		req.Header.Set(key, value)
	}
	return req
}
