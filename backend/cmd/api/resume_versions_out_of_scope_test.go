package main

// product-scope D-20 collapses resume versions into a flat resume resource.
// This negative gate keeps the out-of-scope version-tree routes and operation IDs
// from being reintroduced by hand-written router drift.

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestResumeVersionRoutesRemainUnmountedPerD20(t *testing.T) {
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
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               &apiAuthStore{},
		SessionCookieSecret: "session-secret",
	})
	handler := buildAPIHandlerWithUploadAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{},
		uploadRoutes{},
		resumeRoutes{Handler: resumehandler.New(resumehandler.Options{
			Service: newResumeScenarioService(),
			Session: currentUserFromContext,
		})},
	)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/resumes/01918fa0-0000-7000-8000-000000001000/structured-master"},
		{http.MethodGet, "/api/v1/resumes/01918fa0-0000-7000-8000-000000001000/versions"},
		{http.MethodPost, "/api/v1/resume-versions"},
		{http.MethodGet, "/api/v1/resume-versions/01918fa0-0000-7000-8000-000000001001"},
		{http.MethodPatch, "/api/v1/resume-versions/01918fa0-0000-7000-8000-000000001001"},
		{http.MethodPost, "/api/v1/resume-versions/01918fa0-0000-7000-8000-000000001001/suggestions/01918fa0-0000-7000-8000-000000001002/accept"},
		{http.MethodPost, "/api/v1/resume-versions/01918fa0-0000-7000-8000-000000001001/suggestions/01918fa0-0000-7000-8000-000000001002/reject"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected 404 for out-of-scope route %s %s, got %d body=%s", tt.method, tt.path, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGeneratedRouteCatalogHasNoResumeVersionOperations(t *testing.T) {
	outOfScopeOperations := []string{
		"confirmResumeStructuredMaster",
		"listResumeVersions",
		"getResumeVersion",
		"updateResumeVersion",
		"branchResumeVersion",
		"acceptResumeTailorSuggestion",
		"rejectResumeTailorSuggestion",
		"exportResumeVersion",
		"archiveResumeAsset",
	}
	for _, route := range generated.AllRoutes {
		lowerOp := strings.ToLower(route.OperationID)
		lowerPath := strings.ToLower(route.Path)
		for _, outOfScopeOperation := range outOfScopeOperations {
			if lowerOp == strings.ToLower(outOfScopeOperation) {
				t.Fatalf("generated route catalog still knows out-of-scope operation %q", outOfScopeOperation)
			}
		}
		if strings.Contains(lowerPath, "resume-versions") || strings.Contains(lowerPath, "structured-master") {
			t.Fatalf("generated route catalog still knows out-of-scope resume version path: %s %s", route.OperationID, route.Path)
		}
	}
	for _, outOfScopeOperation := range outOfScopeOperations {
		if _, ok := auth.SessionPolicyForOperation(outOfScopeOperation); ok {
			t.Fatalf("session policy still classifies out-of-scope operation %q", outOfScopeOperation)
		}
	}
}
