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
)

func TestHandlerImplementsResumeSuggestionDecisionSurfaces(t *testing.T) {
	var _ interface {
		AcceptResumeTailorSuggestion(http.ResponseWriter, *http.Request, string, string)
		RejectResumeTailorSuggestion(http.ResponseWriter, *http.Request, string, string)
	} = (*resumehandler.Handler)(nil)
}

func TestResumeSuggestionDecisionAcceptAndReject(t *testing.T) {
	now := time.Date(2026, 5, 18, 11, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		call       func(*resumehandler.Handler, http.ResponseWriter, *http.Request)
		wantAction string
		out        api.ResumeVersion
	}{
		{
			name: "accept",
			call: func(h *resumehandler.Handler, w http.ResponseWriter, r *http.Request) {
				h.AcceptResumeTailorSuggestion(w, r, " version-1 ", " suggestion-1 ")
			},
			wantAction: "accept",
			out:        suggestionDecisionVersion("version-1", "suggestion-1", "accepted", now),
		},
		{
			name: "reject",
			call: func(h *resumehandler.Handler, w http.ResponseWriter, r *http.Request) {
				h.RejectResumeTailorSuggestion(w, r, " version-1 ", " suggestion-1 ")
			},
			wantAction: "reject",
			out:        suggestionDecisionVersion("version-1", "suggestion-1", "rejected", now),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeSuggestionDecisionService{out: tc.out}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()
			req := newSuggestionDecisionRequest(tc.name)

			tc.call(h, rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if svc.in.UserID != "user-1" || svc.in.ResumeVersionID != "version-1" || svc.in.SuggestionID != "suggestion-1" || svc.in.IdempotencyKey != "idem-"+tc.name {
				t.Fatalf("service input = %+v", svc.in)
			}
			if svc.action != tc.wantAction {
				t.Fatalf("service action = %q, want %q", svc.action, tc.wantAction)
			}
			var got api.ResumeVersion
			decodeResponse(t, rec, &got)
			if got.Id != tc.out.Id || len(got.Suggestions) != 1 {
				t.Fatalf("response = %+v", got)
			}
		})
	}
}

func TestResumeSuggestionDecisionValidationAndErrors(t *testing.T) {
	tests := []struct {
		name         string
		versionID    string
		suggestionID string
		err          error
		status       int
		code         string
		wantReason   string
	}{
		{name: "missing version", versionID: "", suggestionID: "suggestion-1", status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "missing suggestion", versionID: "version-1", suggestionID: "", status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "not found", versionID: "version-1", suggestionID: "suggestion-1", err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeTargetJobNotFound},
		{name: "already decided", versionID: "version-1", suggestionID: "suggestion-1", err: resume.ErrSuggestionAlreadyDecided, status: http.StatusConflict, code: sharederrors.CodeValidationFailed, wantReason: "SUGGESTION_ALREADY_DECIDED"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeSuggestionDecisionService{
					out: suggestionDecisionVersion("version-1", "suggestion-1", "accepted", time.Now()),
					err: tc.err,
				},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.AcceptResumeTailorSuggestion(rec, newSuggestionDecisionRequest("accept"), tc.versionID, tc.suggestionID)

			assertAPIError(t, rec, tc.status, tc.code)
			if tc.wantReason != "" {
				var payload api.ApiErrorResponse
				decodeResponse(t, rec, &payload)
				if payload.Error.Details["reason"] != tc.wantReason {
					t.Fatalf("error details = %+v, want reason %q", payload.Error.Details, tc.wantReason)
				}
			}
		})
	}
}

func TestResumeSuggestionDecisionRequiresSessionAndIdempotency(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeSuggestionDecisionService{out: suggestionDecisionVersion("version-1", "suggestion-1", "accepted", time.Now())},
		Session: func(context.Context) (string, bool) { return "", false },
	})
	rec := httptest.NewRecorder()

	h.AcceptResumeTailorSuggestion(rec, newSuggestionDecisionRequest("accept"), "version-1", "suggestion-1")

	assertAPIError(t, rec, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized)

	h = resumehandler.New(resumehandler.Options{
		Service: &fakeSuggestionDecisionService{out: suggestionDecisionVersion("version-1", "suggestion-1", "accepted", time.Now())},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", nil)

	h.AcceptResumeTailorSuggestion(rec, req, "version-1", "suggestion-1")

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestResumeSuggestionDecisionIdempotencyReplayAndMismatch(t *testing.T) {
	store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
		State:          idempotency.StateReplay,
		RecordID:       "idem-rec-suggestion",
		ResponseStatus: http.StatusOK,
		ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","versionType":"targeted","displayName":"Targeted","structuredProfile":{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-18T11:00:00Z","updatedAt":"2026-05-18T11:00:00Z"}`),
	}}
	svc := &fakeSuggestionDecisionService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{Store: store})
	wrapped := mw.Handler("resume", "acceptResumeTailorSuggestion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.AcceptResumeTailorSuggestion(w, r, "version-1", "suggestion-1")
	}))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newSuggestionDecisionRequest("accept"))

	if rec.Code != http.StatusOK || rec.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("replay status=%d header=%q body=%s", rec.Code, rec.Header().Get(idempotency.ReplayHeader), rec.Body.String())
	}
	if svc.calls != 0 {
		t.Fatalf("service calls on replay = %d, want 0", svc.calls)
	}

	mismatchStore := &fakeIdempotencyStore{err: idempotency.ErrFingerprintMismatch}
	wrapped = idempotency.New(idempotency.MiddlewareOptions{Store: mismatchStore}).Handler("resume", "acceptResumeTailorSuggestion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.AcceptResumeTailorSuggestion(w, r, "version-1", "suggestion-1")
	}))
	rec = httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newSuggestionDecisionRequest("accept"))

	assertAPIError(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
}

func TestResumeSuggestionDecisionFixtureParity(t *testing.T) {
	fixtures := map[string]suggestionDecisionFixture{
		"accept": loadSuggestionDecisionFixture(t, "acceptResumeTailorSuggestion.json"),
		"reject": loadSuggestionDecisionFixture(t, "rejectResumeTailorSuggestion.json"),
	}
	for action, fixture := range fixtures {
		t.Run(action+" default", func(t *testing.T) {
			entry := fixture.Scenarios["default"]
			var out api.ResumeVersion
			if err := json.Unmarshal(entry.Response.Body, &out); err != nil {
				t.Fatalf("decode fixture body: %v", err)
			}
			svc := &fakeSuggestionDecisionService{out: out}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			callSuggestionDecision(h, action, rec, newSuggestionDecisionFixtureRequest(entry), out.Id, suggestionIDFromFixture(t, out))

			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
		})

		t.Run(action+" idempotency replay", func(t *testing.T) {
			entry := fixture.Scenarios["idempotency-replay"]
			store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
				State:          idempotency.StateReplay,
				RecordID:       "idem-rec-" + action + "-suggestion",
				ResponseStatus: entry.Response.Status,
				ResponseBody:   entry.Response.Body,
			}}
			svc := &fakeSuggestionDecisionService{}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			wrapped := idempotency.New(idempotency.MiddlewareOptions{Store: store}).Handler("resume", action+"ResumeTailorSuggestion", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callSuggestionDecision(h, action, w, r, "version-1", "suggestion-1")
			}))
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, newSuggestionDecisionFixtureRequest(entry))

			if rec.Header().Get(idempotency.ReplayHeader) != "true" {
				t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
			}
			if svc.calls != 0 {
				t.Fatalf("service calls on replay = %d, want 0", svc.calls)
			}
			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
		})

		t.Run(action+" already decided", func(t *testing.T) {
			entry := fixture.Scenarios["already-decided-409"]
			svc := &fakeSuggestionDecisionService{err: resume.ErrSuggestionAlreadyDecided}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			callSuggestionDecision(h, action, rec, newSuggestionDecisionFixtureRequest(entry), "version-1", "suggestion-1")

			assertRawJSONEqual(t, rec, entry.Response.Status, entry.Response.Body)
		})
	}
}

type fakeSuggestionDecisionService struct {
	calls  int
	action string
	in     resume.SuggestionDecisionRequest
	out    api.ResumeVersion
	err    error
}

func (s *fakeSuggestionDecisionService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *fakeSuggestionDecisionService) AcceptResumeTailorSuggestion(_ context.Context, in resume.SuggestionDecisionRequest) (api.ResumeVersion, error) {
	s.calls++
	s.action = "accept"
	s.in = in
	return s.out, s.err
}

func (s *fakeSuggestionDecisionService) RejectResumeTailorSuggestion(_ context.Context, in resume.SuggestionDecisionRequest) (api.ResumeVersion, error) {
	s.calls++
	s.action = "reject"
	s.in = in
	return s.out, s.err
}

func suggestionDecisionVersion(versionID, suggestionID, status string, now time.Time) api.ResumeVersion {
	return api.ResumeVersion{
		Id:            versionID,
		ResumeAssetId: "asset-1",
		VersionType:   "targeted",
		DisplayName:   "Targeted",
		StructuredProfile: map[string]any{"provenance": map[string]any{
			"promptVersion": "p", "rubricVersion": "r", "modelId": "m", "language": "en", "featureFlag": "f", "dataSourceVersion": "d",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "p", RubricVersion: "r", ModelId: "m", Language: "en", FeatureFlag: "f", DataSourceVersion: "d",
		},
		Suggestions: []any{map[string]any{
			"id":              suggestionID,
			"originalBullet":  "Improved reliability.",
			"suggestedBullet": "Improved reliability with release guardrails.",
			"status":          status,
			"provenance": map[string]any{
				"promptVersion": "p", "rubricVersion": "r", "modelId": "m", "language": "en", "featureFlag": "f", "dataSourceVersion": "d",
			},
			"createdAt": now.Format(time.RFC3339),
			"decidedAt": now.Format(time.RFC3339),
		}},
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	}
}

type suggestionDecisionFixtureEntry struct {
	Request struct {
		Headers map[string]string `json:"headers"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

type suggestionDecisionFixture struct {
	Scenarios map[string]suggestionDecisionFixtureEntry `json:"scenarios"`
}

func loadSuggestionDecisionFixture(t *testing.T, name string) suggestionDecisionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", name))
	if err != nil {
		t.Fatalf("read %s fixture: %v", name, err)
	}
	var fixture suggestionDecisionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode %s fixture: %v", name, err)
	}
	for _, scenario := range []string{"default", "idempotency-replay", "already-decided-409"} {
		if _, ok := fixture.Scenarios[scenario]; !ok {
			t.Fatalf("fixture %s missing scenario %q", name, scenario)
		}
	}
	return fixture
}

func newSuggestionDecisionRequest(action string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/"+action, nil)
	req.Header.Set(idempotency.HeaderName, "idem-"+action)
	return req
}

func newSuggestionDecisionFixtureRequest(entry suggestionDecisionFixtureEntry) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume-versions/version-1/suggestions/suggestion-1/accept", nil)
	for key, value := range entry.Request.Headers {
		req.Header.Set(key, value)
	}
	return req
}

func callSuggestionDecision(h *resumehandler.Handler, action string, w http.ResponseWriter, r *http.Request, versionID string, suggestionID string) {
	if strings.HasPrefix(action, "accept") {
		h.AcceptResumeTailorSuggestion(w, r, versionID, suggestionID)
		return
	}
	h.RejectResumeTailorSuggestion(w, r, versionID, suggestionID)
}

func suggestionIDFromFixture(t *testing.T, version api.ResumeVersion) string {
	t.Helper()
	if len(version.Suggestions) == 0 {
		t.Fatal("fixture version has no suggestions")
	}
	first, ok := version.Suggestions[0].(map[string]any)
	if !ok {
		t.Fatalf("suggestion type = %T", version.Suggestions[0])
	}
	id, _ := first["id"].(string)
	if id == "" {
		t.Fatalf("suggestion missing id: %+v", first)
	}
	return id
}
