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

func TestHandlerImplementsUpdateResumeVersionSurface(t *testing.T) {
	var _ interface {
		UpdateResumeVersion(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestUpdateResumeVersion(t *testing.T) {
	now := time.Date(2026, 5, 17, 19, 0, 0, 0, time.UTC)
	out := updateVersionResponse("version-1", now)
	svc := &fakeUpdateVersionService{out: out}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.UpdateResumeVersion(rec, newUpdateVersionRequest(`{
		"displayName":" Updated version ",
		"focusAngle":" Reliability ",
		"matchScore":0.82,
		"structuredProfile":{"summary":"new summary"}
	}`), "version-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.in.UserID != "user-1" || svc.in.VersionID != "version-1" {
		t.Fatalf("scope = %+v", svc.in)
	}
	if svc.in.DisplayName == nil || *svc.in.DisplayName != "Updated version" {
		t.Fatalf("displayName = %#v", svc.in.DisplayName)
	}
	if svc.in.FocusAngle == nil || *svc.in.FocusAngle != "Reliability" {
		t.Fatalf("focusAngle = %#v", svc.in.FocusAngle)
	}
	if svc.in.MatchScore == nil || *svc.in.MatchScore != 0.82 {
		t.Fatalf("matchScore = %#v", svc.in.MatchScore)
	}
	if svc.in.StructuredProfile["summary"] != "new summary" {
		t.Fatalf("structured profile = %#v", svc.in.StructuredProfile)
	}
	var got api.ResumeVersion
	decodeResponse(t, rec, &got)
	if got.Id != "version-1" || got.DisplayName != out.DisplayName {
		t.Fatalf("response = %+v", got)
	}
}

func TestUpdateResumeVersionRejectsServerOwnedFields(t *testing.T) {
	serverOwnedBodies := []string{
		`{"versionType":"targeted"}`,
		`{"resumeAssetId":"asset-1"}`,
		`{"parentVersionId":"version-0"}`,
		`{"targetJobId":"target-1"}`,
		`{"seedStrategy":"copy_master"}`,
		`{"structuredProfile":{"provenance":{"promptVersion":"client-controlled"}}}`,
	}
	for _, body := range serverOwnedBodies {
		t.Run(body, func(t *testing.T) {
			svc := &fakeUpdateVersionService{out: updateVersionResponse("version-1", time.Now())}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.UpdateResumeVersion(rec, newUpdateVersionRequest(body), "version-1")

			assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
			if svc.calls != 0 {
				t.Fatalf("service calls = %d, want 0", svc.calls)
			}
		})
	}
}

func TestUpdateResumeVersionErrors(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		err    error
		status int
		code   string
	}{
		{name: "blank display name", body: `{"displayName":" "}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "not found", body: `{"displayName":"Updated"}`, err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeTargetJobNotFound},
		{name: "validation", body: `{"displayName":"Updated"}`, err: resume.ErrValidationFailed, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeUpdateVersionService{out: updateVersionResponse("version-1", time.Now()), err: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.UpdateResumeVersion(rec, newUpdateVersionRequest(tc.body), "version-1")

			assertAPIError(t, rec, tc.status, tc.code)
		})
	}
}

func TestUpdateResumeVersionRequiresIdempotencyKey(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeUpdateVersionService{out: updateVersionResponse("version-1", time.Now())},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", strings.NewReader(`{"displayName":"Updated"}`))

	h.UpdateResumeVersion(rec, req, "version-1")

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestUpdateResumeVersionIdempotencyReplay(t *testing.T) {
	store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
		State:          idempotency.StateReplay,
		RecordID:       "idem-rec-1",
		ResponseStatus: http.StatusOK,
		ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","versionType":"structured_master","displayName":"Updated","structuredProfile":{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-17T19:00:00Z","updatedAt":"2026-05-17T19:00:00Z"}`),
	}}
	svc := &fakeUpdateVersionService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 17, 19, 0, 0, 0, time.UTC) },
	})
	wrapped := mw.Handler("resume", "updateResumeVersion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.UpdateResumeVersion(w, r, "version-1")
	}))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newUpdateVersionRequest(`{"displayName":"Updated"}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
	}
	if svc.calls != 0 {
		t.Fatalf("service calls on replay = %d, want 0", svc.calls)
	}
}

func TestUpdateResumeVersionIdempotencyMismatch(t *testing.T) {
	store := &fakeIdempotencyStore{err: idempotency.ErrFingerprintMismatch}
	svc := &fakeUpdateVersionService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{Store: store})
	wrapped := mw.Handler("resume", "updateResumeVersion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.UpdateResumeVersion(w, r, "version-1")
	}))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newUpdateVersionRequest(`{"displayName":"Updated"}`))

	assertAPIError(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
	if svc.calls != 0 {
		t.Fatalf("service calls on mismatch = %d, want 0", svc.calls)
	}
}

func TestUpdateResumeVersionFixtureParity(t *testing.T) {
	fixture := loadUpdateVersionFixture(t)
	for _, scenario := range []string{"default", "validation-error-422"} {
		t.Run(scenario, func(t *testing.T) {
			entry := fixture.Scenarios[scenario]
			var version api.ResumeVersion
			if entry.Response.Status == http.StatusOK {
				if err := json.Unmarshal(entry.Response.Body, &version); err != nil {
					t.Fatalf("decode version body: %v", err)
				}
			}
			svc := &fakeUpdateVersionService{out: version}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.UpdateResumeVersion(rec, newUpdateVersionFixtureRequest(entry), "0195f2d0-0001-7000-8000-000000000202")

			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
			if entry.Response.Status != http.StatusOK && svc.calls != 0 {
				t.Fatalf("service calls = %d, want 0", svc.calls)
			}
		})
	}

	t.Run("idempotency-replay", func(t *testing.T) {
		entry := fixture.Scenarios["idempotency-replay"]
		store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-replay",
			ResponseStatus: entry.Response.Status,
			ResponseBody:   entry.Response.Body,
		}}
		svc := &fakeUpdateVersionService{}
		h := resumehandler.New(resumehandler.Options{
			Service: svc,
			Session: func(context.Context) (string, bool) { return "user-1", true },
		})
		mw := idempotency.New(idempotency.MiddlewareOptions{
			Store: store,
			Now:   func() time.Time { return time.Date(2026, 5, 12, 8, 40, 0, 0, time.UTC) },
		})
		wrapped := mw.Handler("resume", "updateResumeVersion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.UpdateResumeVersion(w, r, "0195f2d0-0001-7000-8000-000000000202")
		}))
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, newUpdateVersionFixtureRequest(entry))

		if rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
		}
		if svc.calls != 0 {
			t.Fatalf("service calls on replay = %d, want 0", svc.calls)
		}
		assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
	})
}

type fakeUpdateVersionService struct {
	calls int
	in    resume.UpdateVersionRequest
	out   api.ResumeVersion
	err   error
}

func (s *fakeUpdateVersionService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *fakeUpdateVersionService) UpdateResumeVersion(_ context.Context, in resume.UpdateVersionRequest) (api.ResumeVersion, error) {
	s.calls++
	s.in = in
	return s.out, s.err
}

func newUpdateVersionRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/resume-versions/version-1", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-update")
	return req
}

func updateVersionResponse(id string, now time.Time) api.ResumeVersion {
	return api.ResumeVersion{
		Id:            id,
		ResumeAssetId: "asset-1",
		VersionType:   sharedtypes.ResumeVersionTypeStructuredMaster,
		DisplayName:   "Updated version",
		StructuredProfile: map[string]any{"summary": "new summary", "provenance": map[string]any{
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

type updateVersionFixtureEntry struct {
	Request struct {
		Headers map[string]string `json:"headers"`
		Body    json.RawMessage   `json:"body"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

type updateVersionFixture struct {
	Scenarios map[string]updateVersionFixtureEntry `json:"scenarios"`
}

func loadUpdateVersionFixture(t *testing.T) updateVersionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "updateResumeVersion.json"))
	if err != nil {
		t.Fatalf("read updateResumeVersion fixture: %v", err)
	}
	var fixture updateVersionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode updateResumeVersion fixture: %v", err)
	}
	for _, scenario := range []string{"default", "idempotency-replay", "validation-error-422"} {
		if _, ok := fixture.Scenarios[scenario]; !ok {
			t.Fatalf("fixture missing scenario %q", scenario)
		}
	}
	return fixture
}

func newUpdateVersionFixtureRequest(entry updateVersionFixtureEntry) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/resume-versions/0195f2d0-0001-7000-8000-000000000202", strings.NewReader(string(entry.Request.Body)))
	for key, value := range entry.Request.Headers {
		req.Header.Set(key, value)
	}
	return req
}
