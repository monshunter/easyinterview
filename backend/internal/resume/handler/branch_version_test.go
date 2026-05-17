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

func TestHandlerImplementsBranchResumeVersionSurface(t *testing.T) {
	var _ interface {
		BranchResumeVersion(http.ResponseWriter, *http.Request)
	} = (*resumehandler.Handler)(nil)
}

func TestBranchResumeVersion(t *testing.T) {
	now := time.Date(2026, 5, 17, 20, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		body   string
		result resume.BranchVersionResult
		status int
	}{
		{
			name: "copy master",
			body: validBranchBody("copy_master"),
			result: resume.BranchVersionResult{
				Status:  http.StatusCreated,
				Version: branchVersionResponse("version-copy", "copy_master", now),
			},
			status: http.StatusCreated,
		},
		{
			name: "blank",
			body: validBranchBody("blank"),
			result: resume.BranchVersionResult{
				Status:  http.StatusCreated,
				Version: branchVersionResponse("version-blank", "blank", now),
			},
			status: http.StatusCreated,
		},
		{
			name: "ai select",
			body: validBranchBody("ai_select"),
			result: resume.BranchVersionResult{
				Status: http.StatusAccepted,
				Accepted: &api.BranchResumeVersionAccepted{
					ResumeVersionId: "version-ai",
					Version:         branchVersionResponse("version-ai", "ai_select", now),
					Job: api.Job{
						Id:           "job-1",
						JobType:      api.JobTypeResumeTailor,
						ResourceType: api.ResourceTypeResumeTailorRun,
						ResourceId:   "tailor-run-1",
						Status:       sharedtypes.JobStatusQueued,
						CreatedAt:    now.Format(time.RFC3339),
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			status: http.StatusAccepted,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeBranchVersionService{result: tc.result}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.BranchResumeVersion(rec, newBranchRequest(tc.body))

			if rec.Code != tc.status {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if svc.in.UserID != "user-1" || svc.in.ParentVersionID != "parent-1" || svc.in.TargetJobID != "target-1" || svc.in.SeedStrategy == "" {
				t.Fatalf("service input = %+v", svc.in)
			}
			var got map[string]any
			decodeResponse(t, rec, &got)
			if tc.status == http.StatusAccepted {
				if got["resumeVersionId"] != tc.result.Accepted.ResumeVersionId {
					t.Fatalf("accepted response = %+v", got)
				}
			} else if got["id"] != tc.result.Version.Id {
				t.Fatalf("version response = %+v", got)
			}
		})
	}
}

func TestBranchResumeVersionValidationAndErrors(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		err    error
		status int
		code   string
	}{
		{name: "missing display name", body: `{"parentVersionId":"parent-1","targetJobId":"target-1","seedStrategy":"copy_master"}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "invalid seed strategy", body: validBranchBody("invalid"), status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "not found", body: validBranchBody("copy_master"), err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeTargetJobNotFound},
		{name: "validation", body: validBranchBody("copy_master"), err: resume.ErrValidationFailed, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeBranchVersionService{result: resume.BranchVersionResult{Status: http.StatusCreated, Version: branchVersionResponse("version-1", "copy_master", time.Now())}, err: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.BranchResumeVersion(rec, newBranchRequest(tc.body))

			assertAPIError(t, rec, tc.status, tc.code)
		})
	}
}

func TestBranchResumeVersionRequiresIdempotencyKey(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeBranchVersionService{result: resume.BranchVersionResult{Status: http.StatusCreated, Version: branchVersionResponse("version-1", "copy_master", time.Now())}},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions", strings.NewReader(validBranchBody("copy_master")))

	h.BranchResumeVersion(rec, req)

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestBranchResumeVersionIdempotencyReplay(t *testing.T) {
	store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
		State:          idempotency.StateReplay,
		RecordID:       "idem-rec-branch",
		ResponseStatus: http.StatusCreated,
		ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","parentVersionId":"parent-1","versionType":"targeted","targetJobId":"target-1","displayName":"Targeted","seedStrategy":"copy_master","structuredProfile":{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-17T20:00:00Z","updatedAt":"2026-05-17T20:00:00Z"}`),
	}}
	svc := &fakeBranchVersionService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 17, 20, 0, 0, 0, time.UTC) },
	})
	wrapped := mw.Handler("resume", "branchResumeVersion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.BranchResumeVersion))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newBranchRequest(validBranchBody("copy_master")))

	if rec.Code != http.StatusCreated || rec.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
	}
	if svc.calls != 0 {
		t.Fatalf("service calls on replay = %d, want 0", svc.calls)
	}
}

func TestBranchResumeVersionIdempotencyMismatch(t *testing.T) {
	store := &fakeIdempotencyStore{err: idempotency.ErrFingerprintMismatch}
	svc := &fakeBranchVersionService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{Store: store})
	wrapped := mw.Handler("resume", "branchResumeVersion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.BranchResumeVersion))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newBranchRequest(validBranchBody("copy_master")))

	assertAPIError(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
	if svc.calls != 0 {
		t.Fatalf("service calls on mismatch = %d, want 0", svc.calls)
	}
}

func TestBranchResumeVersionFixtureParity(t *testing.T) {
	fixture := loadBranchVersionFixture(t)
	for _, scenario := range []string{"default", "copy-master-sync", "blank-sync", "ai-select-202-with-job", "validation-error-422"} {
		t.Run(scenario, func(t *testing.T) {
			entry := fixture.Scenarios[scenario]
			result := resume.BranchVersionResult{Status: entry.Response.Status}
			switch entry.Response.Status {
			case http.StatusCreated:
				if err := json.Unmarshal(entry.Response.Body, &result.Version); err != nil {
					t.Fatalf("decode version body: %v", err)
				}
			case http.StatusAccepted:
				var accepted api.BranchResumeVersionAccepted
				if err := json.Unmarshal(entry.Response.Body, &accepted); err != nil {
					t.Fatalf("decode accepted body: %v", err)
				}
				result.Accepted = &accepted
			}
			svc := &fakeBranchVersionService{result: result}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.BranchResumeVersion(rec, newBranchFixtureRequest(entry))

			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
			if entry.Response.Status != http.StatusCreated && entry.Response.Status != http.StatusAccepted && svc.calls != 0 {
				t.Fatalf("service calls = %d, want 0", svc.calls)
			}
		})
	}

	t.Run("idempotent-replay", func(t *testing.T) {
		entry := fixture.Scenarios["idempotent-replay"]
		store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-branch-replay",
			ResponseStatus: entry.Response.Status,
			ResponseBody:   entry.Response.Body,
		}}
		svc := &fakeBranchVersionService{}
		h := resumehandler.New(resumehandler.Options{
			Service: svc,
			Session: func(context.Context) (string, bool) { return "user-1", true },
		})
		mw := idempotency.New(idempotency.MiddlewareOptions{
			Store: store,
			Now:   func() time.Time { return time.Date(2026, 5, 12, 8, 35, 0, 0, time.UTC) },
		})
		wrapped := mw.Handler("resume", "branchResumeVersion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.BranchResumeVersion))
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, newBranchFixtureRequest(entry))

		if rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
		}
		if svc.calls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", svc.calls)
		}
		assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
	})
}

type fakeBranchVersionService struct {
	calls  int
	in     resume.BranchVersionRequest
	result resume.BranchVersionResult
	err    error
}

func (s *fakeBranchVersionService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *fakeBranchVersionService) BranchResumeVersion(_ context.Context, in resume.BranchVersionRequest) (resume.BranchVersionResult, error) {
	s.calls++
	s.in = in
	return s.result, s.err
}

func newBranchRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-branch")
	return req
}

func validBranchBody(seedStrategy string) string {
	return `{"parentVersionId":"parent-1","targetJobId":"target-1","seedStrategy":"` + seedStrategy + `","displayName":" Targeted ","focusAngle":" Platform evidence "}`
}

func branchVersionResponse(id string, seedStrategy string, now time.Time) api.ResumeVersion {
	strategy := sharedtypes.ResumeSeedStrategy(seedStrategy)
	focusAngle := "Platform evidence"
	parentID := "parent-1"
	targetID := "target-1"
	return api.ResumeVersion{
		Id:              id,
		ResumeAssetId:   "asset-1",
		ParentVersionId: &parentID,
		VersionType:     sharedtypes.ResumeVersionTypeTargeted,
		TargetJobId:     &targetID,
		DisplayName:     "Targeted",
		SeedStrategy:    &strategy,
		FocusAngle:      &focusAngle,
		StructuredProfile: map[string]any{"provenance": map[string]any{
			"promptVersion": "p", "rubricVersion": "r", "modelId": "m", "language": "en", "featureFlag": "f", "dataSourceVersion": "d",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "p", RubricVersion: "r", ModelId: "m", Language: "en", FeatureFlag: "f", DataSourceVersion: "d",
		},
		Suggestions: []any{},
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
}

type branchVersionFixtureEntry struct {
	Request struct {
		Headers map[string]string `json:"headers"`
		Body    json.RawMessage   `json:"body"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

type branchVersionFixture struct {
	Scenarios map[string]branchVersionFixtureEntry `json:"scenarios"`
}

func loadBranchVersionFixture(t *testing.T) branchVersionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "branchResumeVersion.json"))
	if err != nil {
		t.Fatalf("read branchResumeVersion fixture: %v", err)
	}
	var fixture branchVersionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode branchResumeVersion fixture: %v", err)
	}
	for _, scenario := range []string{"default", "copy-master-sync", "blank-sync", "ai-select-202-with-job", "idempotent-replay", "validation-error-422"} {
		if _, ok := fixture.Scenarios[scenario]; !ok {
			t.Fatalf("fixture missing scenario %q", scenario)
		}
	}
	return fixture
}

func newBranchFixtureRequest(entry branchVersionFixtureEntry) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions", strings.NewReader(string(entry.Request.Body)))
	for key, value := range entry.Request.Headers {
		req.Header.Set(key, value)
	}
	return req
}
