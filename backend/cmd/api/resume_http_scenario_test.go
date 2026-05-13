package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestResumeRegisterListHTTPScenario(t *testing.T) {
	dir := t.TempDir()
	writeAPIFile(t, filepath.Join(dir, "config.yaml"), `
runtime:
  appVersion: "1.2.3"
  defaultUiLanguage: zh-CN
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	userID := "01918fa0-0000-7000-8000-000000009001"
	authStore := &apiAuthStore{
		session: auth.SessionRecord{
			ID:        "session-1",
			UserID:    userID,
			Status:    auth.SessionStatusActive,
			ExpiresAt: time.Now().Add(auth.SessionTTL),
		},
		user: auth.UserContext{ID: userID, Email: "resume@example.com"},
	}
	authService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "session-secret",
	})
	resumeSvc := newResumeScenarioService()
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		authService,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{Handler: resumehandler.New(resumehandler.Options{
			Service: resumeSvc,
			Session: currentUserFromContext,
		})},
	)

	unauth := httptest.NewRecorder()
	handler.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, "/api/v1/resumes", nil))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized list status = %d body=%s", unauth.Code, unauth.Body.String())
	}

	registerRaw := doResumeJSON(t, handler, true, http.MethodPost, "/api/v1/resumes", "resume-scenario-register", api.RegisterResumeRequest{
		Title:      "Scenario Resume",
		Language:   "en",
		SourceType: strPtr("paste"),
		RawText:    strPtr("Private resume body that must not leak."),
	}, http.StatusAccepted)
	var registered api.ResumeAssetWithJob
	decodeJSON(t, registerRaw, &registered)
	if registered.ResumeAssetId == "" || registered.Job.Status != sharedtypes.JobStatusQueued || registered.Job.JobType != api.JobTypeResumeParse {
		t.Fatalf("unexpected register response: %+v", registered)
	}

	replayRaw := doResumeJSON(t, handler, true, http.MethodPost, "/api/v1/resumes", "resume-scenario-register", api.RegisterResumeRequest{
		Title:      "Scenario Resume",
		Language:   "en",
		SourceType: strPtr("paste"),
		RawText:    strPtr("Private resume body that must not leak."),
	}, http.StatusAccepted)
	var replay api.ResumeAssetWithJob
	decodeJSON(t, replayRaw, &replay)
	if replay.ResumeAssetId != registered.ResumeAssetId || resumeSvc.createCount != 1 {
		t.Fatalf("idempotent replay created duplicate: replay=%+v creates=%d", replay, resumeSvc.createCount)
	}

	detailRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes/"+registered.ResumeAssetId, "", nil, http.StatusOK)
	var detail api.ResumeAsset
	decodeJSON(t, detailRaw, &detail)
	if detail.Id != registered.ResumeAssetId || detail.ParseStatus != sharedtypes.TargetJobParseStatusQueued {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	listRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes?pageSize=20", "", nil, http.StatusOK)
	var list api.PaginatedResumeAsset
	decodeJSON(t, listRaw, &list)
	if len(list.Items) != 1 || list.Items[0].Id != registered.ResumeAssetId || list.PageInfo.PageSize != 20 {
		t.Fatalf("unexpected list: %+v", list)
	}

	missingRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes/01918fa0-0000-7000-8000-000000009999", "", nil, http.StatusNotFound)
	var missing api.ApiErrorResponse
	decodeJSON(t, missingRaw, &missing)
	if missing.Error.Code != sharederrors.CodeTargetJobNotFound {
		t.Fatalf("missing error = %+v", missing.Error)
	}
}

func doResumeJSON(t *testing.T, handler http.Handler, authenticated bool, method, path string, idempotencyKey string, body any, wantStatus int) []byte {
	t.Helper()
	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reqBody = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if authenticated {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec.Body.Bytes()
}

type resumeScenarioService struct {
	assets      map[string]api.ResumeAsset
	byKey       map[string]api.ResumeAssetWithJob
	createCount int
}

func newResumeScenarioService() *resumeScenarioService {
	return &resumeScenarioService{
		assets: map[string]api.ResumeAsset{},
		byKey:  map[string]api.ResumeAssetWithJob{},
	}
}

func (s *resumeScenarioService) RegisterResume(_ context.Context, in resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	key := in.UserID + ":" + in.IdempotencyKey
	if existing, ok := s.byKey[key]; ok {
		return existing, nil
	}
	s.createCount++
	assetID := "01918fa0-0000-7000-8000-000000009101"
	jobID := "01918fa0-0000-7000-8000-000000009201"
	now := "2026-05-13T09:00:00Z"
	sourceType := in.SourceType
	rawText := in.RawText
	s.assets[assetID] = api.ResumeAsset{
		Id:           assetID,
		Title:        in.Title,
		Language:     in.Language,
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceType:   &sourceType,
		OriginalText: &rawText,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	out := api.ResumeAssetWithJob{
		ResumeAssetId: assetID,
		Job: api.Job{
			Id:           jobID,
			JobType:      api.JobTypeResumeParse,
			ResourceType: api.ResourceTypeResumeAsset,
			ResourceId:   assetID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	s.byKey[key] = out
	return out, nil
}

func (s *resumeScenarioService) GetResume(_ context.Context, _ string, resumeAssetID string) (api.ResumeAsset, error) {
	asset, ok := s.assets[resumeAssetID]
	if !ok {
		return api.ResumeAsset{}, resume.ErrNotFound
	}
	return asset, nil
}

func (s *resumeScenarioService) ListResumes(_ context.Context, in resume.ListRequest) (api.PaginatedResumeAsset, error) {
	items := make([]api.ResumeAsset, 0, len(s.assets))
	for _, asset := range s.assets {
		items = append(items, asset)
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = sharedtypes.DefaultPageSize
	}
	return api.PaginatedResumeAsset{
		Items:    items,
		PageInfo: api.PageInfo{PageSize: pageSize, HasMore: false},
	}, nil
}
