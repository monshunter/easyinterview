package mockruntime

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerReturnsFixtureResponses(t *testing.T) {
	repoRoot := findRepoRoot(t)
	registry, err := LoadRegistry(filepath.Join(repoRoot, "openapi", "fixtures"))
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	handler := NewHandler(registry)

	cases := []struct {
		name        string
		method      string
		path        string
		fixturePath string
	}{
		{
			name:        "getMe",
			method:      http.MethodGet,
			path:        "/api/v1/me",
			fixturePath: "openapi/fixtures/Auth/getMe.json",
		},
		{
			name:        "listTargetJobs",
			method:      http.MethodGet,
			path:        "/api/v1/targets",
			fixturePath: "openapi/fixtures/TargetJobs/listTargetJobs.json",
		},
		{
			name:        "getPracticeSession",
			method:      http.MethodGet,
			path:        "/api/v1/practice/sessions/01918fa0-0000-7000-8000-000000005000",
			fixturePath: "openapi/fixtures/PracticeSessions/getPracticeSession.json",
		},
		{
			name:        "requestPrivacyExport",
			method:      http.MethodPost,
			path:        "/api/v1/privacy/exports",
			fixturePath: "openapi/fixtures/Privacy/requestPrivacyExport.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			wantStatus, wantBody := fixtureDefaultResponse(t, filepath.Join(repoRoot, tc.fixturePath))
			if response.Code != wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, wantStatus)
			}
			assertJSONEqual(t, response.Body.Bytes(), wantBody)
		})
	}
}

func TestHandlerSelectsNamedSeedScenariosAndFailsUnknown(t *testing.T) {
	repoRoot := findRepoRoot(t)
	registry, err := LoadRegistry(filepath.Join(repoRoot, "openapi", "fixtures"))
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	handler := NewHandler(registry)

	cases := []struct {
		name        string
		method      string
		path        string
		fixturePath string
		scenario    string
	}{
		{
			name:        "unauthenticated",
			method:      http.MethodGet,
			path:        "/api/v1/me",
			fixturePath: "openapi/fixtures/Auth/getMe.json",
			scenario:    "unauthenticated",
		},
		{
			name:        "authenticated",
			method:      http.MethodGet,
			path:        "/api/v1/me",
			fixturePath: "openapi/fixtures/Auth/getMe.json",
			scenario:    "authenticated",
		},
		{
			name:        "missing-session",
			method:      http.MethodGet,
			path:        "/api/v1/practice/sessions/01918fa0-0000-7000-8000-000000005000",
			fixturePath: "openapi/fixtures/PracticeSessions/getPracticeSession.json",
			scenario:    "missing-session",
		},
		{
			name:        "generating report",
			method:      http.MethodGet,
			path:        "/api/v1/reports/01918fa0-0000-7000-8000-000000007000",
			fixturePath: "openapi/fixtures/Reports/getFeedbackReport.json",
			scenario:    "generating",
		},
		{
			name:        "privacy-delete-requested",
			method:      http.MethodPost,
			path:        "/api/v1/privacy/deletions",
			fixturePath: "openapi/fixtures/Privacy/requestPrivacyDelete.json",
			scenario:    "privacy-delete-requested",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				request := httptest.NewRequest(tc.method, tc.path, nil)
				request.Header.Set("Prefer", "example="+tc.scenario)
				response := httptest.NewRecorder()

				handler.ServeHTTP(response, request)

				wantStatus, wantBody := fixtureScenarioResponse(t, filepath.Join(repoRoot, tc.fixturePath), tc.scenario)
				if response.Code != wantStatus {
					t.Fatalf("status = %d, want %d; body=%s", response.Code, wantStatus, response.Body.String())
				}
				assertJSONEqual(t, response.Body.Bytes(), wantBody)
			}
		})
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	request.Header.Set("Prefer", "example=does-not-exist")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unknown scenario status = %d, want 400", response.Code)
	}
	if !strings.Contains(response.Body.String(), "unknown fixture scenario") {
		t.Fatalf("unknown scenario did not fail loudly: %s", response.Body.String())
	}
}

func fixtureDefaultResponse(t *testing.T, path string) (int, []byte) {
	t.Helper()
	return fixtureScenarioResponse(t, path, "default")
}

func fixtureScenarioResponse(t *testing.T, path string, scenarioName string) (int, []byte) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture struct {
		Scenarios map[string]struct {
			Response struct {
				Status int             `json:"status"`
				Body   json.RawMessage `json:"body"`
			} `json:"response"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	scenario, ok := fixture.Scenarios[scenarioName]
	if !ok {
		t.Fatalf("fixture %s missing scenario %q", path, scenarioName)
	}
	return scenario.Response.Status, scenario.Response.Body
}

func assertJSONEqual(t *testing.T, got []byte, want []byte) {
	t.Helper()
	var gotBuffer bytes.Buffer
	var wantBuffer bytes.Buffer
	if err := json.Compact(&gotBuffer, got); err != nil {
		t.Fatalf("compact got JSON: %v\n%s", err, got)
	}
	if err := json.Compact(&wantBuffer, want); err != nil {
		t.Fatalf("compact want JSON: %v\n%s", err, want)
	}
	if !bytes.Equal(gotBuffer.Bytes(), wantBuffer.Bytes()) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", gotBuffer.String(), wantBuffer.String())
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for dir := wd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "openapi", "openapi.yaml")); err == nil {
			return dir
		}
	}
	t.Fatal("repo root not found")
	return ""
}
