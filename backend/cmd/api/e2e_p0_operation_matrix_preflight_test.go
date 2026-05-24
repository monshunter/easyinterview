package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
)

type e2eP0OperationMatrixEntry struct {
	operationID string
	method      string
	path        string
	fixture     string
	mainRoute   string
	mainHandler string
	handlerFile string
	handlerDecl string
}

func TestE2EP0OperationMatrixPreflight(t *testing.T) {
	root := scenarioRepoRoot(t)
	mainSource := readE2EP0PreflightFile(t, root, "backend/cmd/api/main.go")
	routes := map[string]api.Route{}
	for _, route := range api.AllRoutes {
		routes[route.OperationID] = route
	}

	matrix := []e2eP0OperationMatrixEntry{
		{
			operationID: "registerResume", method: "post", path: "/resumes",
			fixture:   "openapi/fixtures/Resumes/registerResume.json",
			mainRoute: "POST /api/v1/resumes", mainHandler: "resume.Handler.RegisterResume",
			handlerFile: "backend/internal/resume/handler/register.go", handlerDecl: "func (h *Handler) RegisterResume",
		},
		{
			operationID: "importTargetJob", method: "post", path: "/targets/import",
			fixture:   "openapi/fixtures/TargetJobs/importTargetJob.json",
			mainRoute: "POST /api/v1/targets/import", mainHandler: "targetJobHandler.ImportTargetJob",
			handlerFile: "backend/internal/targetjob/handler.go", handlerDecl: "func (h *Handler) ImportTargetJob",
		},
		{
			operationID: "getTargetJob", method: "get", path: "/targets/{targetJobId}",
			fixture:   "openapi/fixtures/TargetJobs/getTargetJob.json",
			mainRoute: "GET /api/v1/targets/{targetJobId}", mainHandler: "targetJobHandler.GetTargetJob",
			handlerFile: "backend/internal/targetjob/handler.go", handlerDecl: "func (h *Handler) GetTargetJob",
		},
		{
			operationID: "createPracticePlan", method: "post", path: "/practice/plans",
			fixture:   "openapi/fixtures/PracticePlans/createPracticePlan.json",
			mainRoute: "POST /api/v1/practice/plans", mainHandler: "practice.Handler.CreatePracticePlan",
			handlerFile: "backend/internal/api/practice/handler.go", handlerDecl: "func (h *Handler) CreatePracticePlan",
		},
		{
			operationID: "startPracticeSession", method: "post", path: "/practice/sessions",
			fixture:   "openapi/fixtures/PracticeSessions/startPracticeSession.json",
			mainRoute: "POST /api/v1/practice/sessions", mainHandler: "practice.Handler.StartPracticeSession",
			handlerFile: "backend/internal/api/practice/handler.go", handlerDecl: "func (h *Handler) StartPracticeSession",
		},
		{
			operationID: "appendSessionEvent", method: "post", path: "/practice/sessions/{sessionId}/events",
			fixture:   "openapi/fixtures/PracticeSessions/appendSessionEvent.json",
			mainRoute: "POST /api/v1/practice/sessions/{sessionId}/events", mainHandler: "practice.Handler.AppendSessionEvent",
			handlerFile: "backend/internal/api/practice/session_event_handlers.go", handlerDecl: "func (h *Handler) AppendSessionEvent",
		},
		{
			operationID: "completePracticeSession", method: "post", path: "/practice/sessions/{sessionId}/complete",
			fixture:   "openapi/fixtures/PracticeSessions/completePracticeSession.json",
			mainRoute: "POST /api/v1/practice/sessions/{sessionId}/complete", mainHandler: "practice.Handler.CompletePracticeSession",
			handlerFile: "backend/internal/api/practice/session_event_handlers.go", handlerDecl: "func (h *Handler) CompletePracticeSession",
		},
		{
			operationID: "getFeedbackReport", method: "get", path: "/reports/{reportId}",
			fixture:   "openapi/fixtures/Reports/getFeedbackReport.json",
			mainRoute: "GET /api/v1/reports/{reportId}", mainHandler: "reports.Handler.GetFeedbackReport",
			handlerFile: "backend/internal/api/reports/get_feedback_report.go", handlerDecl: "func (h *Handler) GetFeedbackReport",
		},
		{
			operationID: "getJob", method: "get", path: "/jobs/{jobId}",
			fixture:   "openapi/fixtures/Jobs/getJob.json",
			mainRoute: "GET /api/v1/jobs/{jobId}", mainHandler: "jobs.Handler.GetJob",
			handlerFile: "backend/internal/api/jobs/handler.go", handlerDecl: "func (h *Handler) GetJob",
		},
	}

	if len(matrix) != 9 {
		t.Fatalf("operation matrix size=%d, want 9", len(matrix))
	}
	for _, entry := range matrix {
		t.Run(entry.operationID, func(t *testing.T) {
			route, ok := routes[entry.operationID]
			if !ok {
				t.Fatalf("generated route for %s missing", entry.operationID)
			}
			if route.Method != entry.method || route.Path != entry.path {
				t.Fatalf("generated route for %s = %s %s, want %s %s", entry.operationID, route.Method, route.Path, entry.method, entry.path)
			}
			if _, err := os.Stat(filepath.Join(root, entry.fixture)); err != nil {
				t.Fatalf("fixture %s missing: %v", entry.fixture, err)
			}
			if !strings.Contains(mainSource, entry.mainRoute) {
				t.Fatalf("cmd/api route %q missing for %s", entry.mainRoute, entry.operationID)
			}
			if !strings.Contains(mainSource, entry.mainHandler) {
				t.Fatalf("cmd/api handler wiring %q missing for %s", entry.mainHandler, entry.operationID)
			}
			handlerSource := readE2EP0PreflightFile(t, root, entry.handlerFile)
			if !strings.Contains(handlerSource, entry.handlerDecl) {
				t.Fatalf("handler declaration %q missing in %s", entry.handlerDecl, entry.handlerFile)
			}
		})
	}
}

func readE2EP0PreflightFile(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}
