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
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
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
	authService := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "session-secret",
	})
	resumeSvc := newResumeScenarioService()
	handler := buildAPIHandler(
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
		reportRoutes{},
		jobsRoutes{},
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
	var registered api.ResumeWithJob
	decodeJSON(t, registerRaw, &registered)
	if registered.ResumeId == "" || registered.Job.Status != sharedtypes.JobStatusQueued || registered.Job.JobType != api.JobTypeResumeParse {
		t.Fatalf("unexpected register response: %+v", registered)
	}

	replayRaw := doResumeJSON(t, handler, true, http.MethodPost, "/api/v1/resumes", "resume-scenario-register", api.RegisterResumeRequest{
		Title:      "Scenario Resume",
		Language:   "en",
		SourceType: strPtr("paste"),
		RawText:    strPtr("Private resume body that must not leak."),
	}, http.StatusAccepted)
	var replay api.ResumeWithJob
	decodeJSON(t, replayRaw, &replay)
	if replay.ResumeId != registered.ResumeId || resumeSvc.createCount != 1 {
		t.Fatalf("idempotent replay created duplicate: replay=%+v creates=%d", replay, resumeSvc.createCount)
	}

	detailRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes/"+registered.ResumeId, "", nil, http.StatusOK)
	var detail api.Resume
	decodeJSON(t, detailRaw, &detail)
	if detail.Id != registered.ResumeId || detail.ParseStatus != sharedtypes.TargetJobParseStatusQueued {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	listRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes?pageSize=20", "", nil, http.StatusOK)
	var list api.PaginatedResume
	decodeJSON(t, listRaw, &list)
	if len(list.Items) != 1 || list.Items[0].Id != registered.ResumeId || list.PageInfo.PageSize != 20 {
		t.Fatalf("unexpected list: %+v", list)
	}

	missingRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes/01918fa0-0000-7000-8000-000000009999", "", nil, http.StatusNotFound)
	var missing api.ApiErrorResponse
	decodeJSON(t, missingRaw, &missing)
	if missing.Error.Code != sharederrors.CodeResourceNotFound {
		t.Fatalf("missing error = %+v", missing.Error)
	}
}

func TestResumeRegisterListHTTPValidationScenario(t *testing.T) {
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
	authService := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               authStore,
		SessionCookieSecret: "session-secret",
	})
	handler := buildAPIHandler(
		loader,
		apiRuntimeFlags{},
		authService,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{Handler: resumehandler.New(resumehandler.Options{
			Service: &resumeValidationScenarioService{},
			Session: currentUserFromContext,
		})},
		reportRoutes{},
		jobsRoutes{},
	)

	registerRaw := doResumeJSON(t, handler, true, http.MethodPost, "/api/v1/resumes", "resume-validation", api.RegisterResumeRequest{
		Title:        "Scenario Resume",
		Language:     "en",
		SourceType:   strPtr("upload"),
		FileObjectId: strPtr("01918fa0-0000-7000-8000-000000000301"),
	}, http.StatusUnprocessableEntity)
	var registerErr api.ApiErrorResponse
	decodeJSON(t, registerRaw, &registerErr)
	if registerErr.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("register error = %+v", registerErr.Error)
	}

	listRaw := doResumeJSON(t, handler, true, http.MethodGet, "/api/v1/resumes?cursor=not-a-valid-cursor", "", nil, http.StatusUnprocessableEntity)
	var listErr api.ApiErrorResponse
	decodeJSON(t, listRaw, &listErr)
	if listErr.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("list error = %+v", listErr.Error)
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

type resumeValidationScenarioService struct {
	resumeScenarioService
}

func (s *resumeValidationScenarioService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeWithJob, error) {
	return api.ResumeWithJob{}, resume.ErrValidationFailed
}

func (s *resumeValidationScenarioService) ListResumes(context.Context, resume.ListRequest) (api.PaginatedResume, error) {
	return api.PaginatedResume{}, resumestore.ErrInvalidCursor
}

type resumeScenarioService struct {
	resumes     map[string]api.Resume
	summaries   map[string]api.ResumeSummary
	byKey       map[string]api.ResumeWithJob
	createCount int
}

func newResumeScenarioService() *resumeScenarioService {
	return &resumeScenarioService{
		resumes:   map[string]api.Resume{},
		summaries: map[string]api.ResumeSummary{},
		byKey:     map[string]api.ResumeWithJob{},
	}
}

func (s *resumeScenarioService) RegisterResume(_ context.Context, in resume.RegisterInput) (api.ResumeWithJob, error) {
	key := in.UserID + ":" + in.IdempotencyKey
	if existing, ok := s.byKey[key]; ok {
		return existing, nil
	}
	s.createCount++
	resumeID := "01918fa0-0000-7000-8000-000000009101"
	jobID := "01918fa0-0000-7000-8000-000000009201"
	now := "2026-06-13T09:00:00Z"
	sourceType := in.SourceType
	rawText := in.RawText
	s.resumes[resumeID] = api.Resume{
		Id:           resumeID,
		Title:        in.Title,
		Language:     in.Language,
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceType:   &sourceType,
		OriginalText: &rawText,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.summaries[resumeID] = api.ResumeSummary{
		Id:                 resumeID,
		Title:              in.Title,
		DisplayName:        "",
		Language:           in.Language,
		SourceType:         sourceType,
		ParseStatus:        sharedtypes.TargetJobParseStatusQueued,
		SummaryHeadline:    nil,
		HasReadableContent: sourceType == "paste" && rawText != "",
		UpdatedAt:          now,
	}
	out := api.ResumeWithJob{
		ResumeId: resumeID,
		Job: api.Job{
			Id:           jobID,
			JobType:      api.JobTypeResumeParse,
			ResourceType: api.ResourceTypeResumeAsset,
			ResourceId:   resumeID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	s.byKey[key] = out
	return out, nil
}

func (s *resumeScenarioService) GetResume(_ context.Context, _ string, resumeID string) (api.Resume, error) {
	rec, ok := s.resumes[resumeID]
	if !ok {
		return api.Resume{}, resume.ErrNotFound
	}
	return rec, nil
}

func (s *resumeScenarioService) ListResumes(_ context.Context, in resume.ListRequest) (api.PaginatedResume, error) {
	items := make([]api.ResumeSummary, 0, len(s.summaries))
	for _, rec := range s.summaries {
		items = append(items, rec)
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = sharedtypes.DefaultPageSize
	}
	return api.PaginatedResume{
		Items:    items,
		PageInfo: api.PageInfo{PageSize: pageSize, HasMore: false},
	}, nil
}
