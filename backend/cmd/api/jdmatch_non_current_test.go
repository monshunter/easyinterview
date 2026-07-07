package main

// product-scope v2.1 D-17 keeps JD-Match outside the current product scope.
// This negative gate keeps the non-current surface from coming back: no /api/v1/jd-match/*
// route may be mounted and the generated route catalog must not know any
// JobMatch operation.

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestJDMatchRoutesRemainUnmountedPerD17(t *testing.T) {
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
	handler := buildAPIHandlerWithUploadReportJobsAndHandlers(
		loader,
		apiRuntimeFlags{},
		service,
		targetjob.NewHandler(),
		practiceRoutes{Handler: apipractice.NewHandler(apipractice.HandlerOptions{})},
		uploadRoutes{},
		resumeRoutes{},
		reportRoutes{},
		jobsRoutes{},
	)

	paths := []string{
		"/api/v1/jd-match/profile",
		"/api/v1/jd-match/agent-status",
		"/api/v1/jd-match/recommendations",
		"/api/v1/jd-match/watchlist",
		"/api/v1/jd-match/search",
		"/api/v1/jd-match/saved-searches",
		"/api/v1/jd-match/market-signals",
	}
	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for non-current route %s, got %d body=%s", p, rec.Code, rec.Body.String())
		}
	}
}

func TestGeneratedRouteCatalogHasNoJobMatchOperations(t *testing.T) {
	for _, route := range generated.AllRoutes {
		lowerOp := strings.ToLower(route.OperationID)
		if strings.Contains(lowerOp, "jobmatch") || strings.Contains(lowerOp, "watchlist") ||
			strings.Contains(lowerOp, "savedsearch") || strings.Contains(lowerOp, "marketsignal") ||
			strings.Contains(lowerOp, "agentscan") || strings.Contains(lowerOp, "jobrecommendation") ||
			strings.Contains(strings.ToLower(route.Path), "jd-match") {
			t.Fatalf("generated route catalog still knows non-current JD-Match surface: %s %s", route.OperationID, route.Path)
		}
		if requirement, ok := auth.SessionPolicyForOperation(route.OperationID); !ok || requirement == "" {
			t.Fatalf("session policy cannot classify %s", route.OperationID)
		}
	}
	for _, nonCurrent := range []string{
		"getJobMatchProfile", "getAgentScanStatus", "listJobRecommendations",
		"getJobRecommendation", "markJobNotRelevant", "listWatchlist",
		"addToWatchlist", "removeFromWatchlist", "searchJobs",
		"listSavedSearches", "createSavedSearch", "getMarketSignals",
	} {
		if _, ok := auth.SessionPolicyForOperation(nonCurrent); ok {
			t.Fatalf("session policy still classifies non-current operation %q", nonCurrent)
		}
	}
}
