package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestHandlerImplementsConfirmStructuredMasterSurface(t *testing.T) {
	var _ interface {
		ConfirmResumeStructuredMaster(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestConfirmStructuredMaster(t *testing.T) {
	now := time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC)
	out := api.ResumeVersion{
		Id:            "version-1",
		ResumeAssetId: "asset-1",
		VersionType:   sharedtypes.ResumeVersionTypeStructuredMaster,
		DisplayName:   "Structured master",
		StructuredProfile: map[string]any{"headline": "Senior engineer", "provenance": map[string]any{
			"promptVersion": "resume_profile.v1", "rubricVersion": "not_applicable", "modelId": "model-1", "language": "en", "featureFlag": "none", "dataSourceVersion": "asset.v1",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "resume_profile.v1", RubricVersion: "not_applicable", ModelId: "model-1", Language: "en", FeatureFlag: "none", DataSourceVersion: "asset.v1",
		},
		Suggestions: []any{},
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
	svc := &fakeConfirmService{confirmOut: out}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.ConfirmResumeStructuredMaster(rec, newConfirmRequest(`{
		"displayName":" Structured master ",
		"language":"en",
		"structuredProfile":{
			"headline":"Senior engineer",
			"provenance":{
				"promptVersion":"resume_profile.v1",
				"rubricVersion":"not_applicable",
				"modelId":"model-1",
				"language":"en",
				"featureFlag":"none",
				"dataSourceVersion":"asset.v1"
			}
		}
	}`), "asset-1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.confirmIn.UserID != "user-1" || svc.confirmIn.ResumeAssetID != "asset-1" || svc.confirmIn.DisplayName != "Structured master" || svc.confirmIn.Language != "en" {
		t.Fatalf("confirm input = %+v", svc.confirmIn)
	}
	if svc.confirmIn.StructuredProfile["headline"] != "Senior engineer" {
		t.Fatalf("structured profile = %+v", svc.confirmIn.StructuredProfile)
	}
	var got api.ResumeVersion
	decodeResponse(t, rec, &got)
	if got.Id != out.Id || got.VersionType != sharedtypes.ResumeVersionTypeStructuredMaster {
		t.Fatalf("response = %+v", got)
	}
}

func TestConfirmStructuredMasterValidationAndErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		err        error
		status     int
		code       string
		wantReason string
	}{
		{name: "missing structured profile", body: `{"displayName":"Master"}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "blank display name", body: `{"displayName":" ","structuredProfile":{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "already exists", body: validConfirmBody(), err: resume.ErrStructuredMasterAlreadyExists, status: http.StatusConflict, code: sharederrors.CodeResumeStructuredMasterAlreadyExists},
		{name: "parse not ready", body: validConfirmBody(), err: resume.ErrAssetParseNotReady, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed, wantReason: "PARSE_NOT_READY"},
		{name: "not found", body: validConfirmBody(), err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeTargetJobNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeConfirmService{confirmErr: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.ConfirmResumeStructuredMaster(rec, newConfirmRequest(tc.body), "asset-1")

			assertAPIError(t, rec, tc.status, tc.code)
			if tc.wantReason != "" {
				var payload api.ApiErrorResponse
				decodeResponse(t, rec, &payload)
				if payload.Error.Details["reason"] != tc.wantReason {
					t.Fatalf("details.reason = %#v, want %q", payload.Error.Details["reason"], tc.wantReason)
				}
			}
		})
	}
}

func TestConfirmStructuredMasterIdempotencyReplay(t *testing.T) {
	store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
		State:          idempotency.StateReplay,
		RecordID:       "idem-rec-1",
		ResponseStatus: http.StatusCreated,
		ResponseBody:   []byte(`{"id":"version-replay","resumeAssetId":"asset-1","versionType":"structured_master","displayName":"Structured master","structuredProfile":{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}},"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"},"suggestions":[],"createdAt":"2026-05-17T16:00:00Z","updatedAt":"2026-05-17T16:00:00Z"}`),
	}}
	svc := &fakeConfirmService{}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC) },
	})
	wrapped := mw.Handler("resume", "confirmResumeStructuredMaster", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ConfirmResumeStructuredMaster(w, r, "asset-1")
	}))
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, newConfirmRequest(validConfirmBody()))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
	}
	if svc.confirmCalls != 0 {
		t.Fatalf("service calls on replay = %d, want 0", svc.confirmCalls)
	}
}

type fakeConfirmService struct {
	confirmCalls int
	confirmIn    resume.ConfirmStructuredMasterInput
	confirmOut   api.ResumeVersion
	confirmErr   error
}

func (s *fakeConfirmService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *fakeConfirmService) ConfirmStructuredMaster(_ context.Context, in resume.ConfirmStructuredMasterInput) (api.ResumeVersion, error) {
	s.confirmCalls++
	s.confirmIn = in
	return s.confirmOut, s.confirmErr
}

func newConfirmRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/asset-1/structured-master", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-confirm")
	return req
}

func validConfirmBody() string {
	return `{"displayName":"Structured master","language":"en","structuredProfile":{"headline":"Senior engineer","provenance":{"promptVersion":"resume_profile.v1","rubricVersion":"not_applicable","modelId":"model-1","language":"en","featureFlag":"none","dataSourceVersion":"asset.v1"}}}`
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
}
