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
		name       string
		method     string
		path       string
		scenario   string
		wantStatus int
		wantField  [2]string
	}{
		{
			name:       "unauthenticated",
			method:     http.MethodGet,
			path:       "/api/v1/me",
			scenario:   "unauthenticated",
			wantStatus: http.StatusUnauthorized,
			wantField:  [2]string{"error.code", "AUTH_UNAUTHORIZED"},
		},
		{
			name:       "authenticated",
			method:     http.MethodGet,
			path:       "/api/v1/me",
			scenario:   "authenticated",
			wantStatus: http.StatusOK,
			wantField:  [2]string{"displayName", "Alice Example"},
		},
		{
			name:       "missing-session",
			method:     http.MethodGet,
			path:       "/api/v1/practice/sessions/01918fa0-0000-7000-8000-000000005000",
			scenario:   "missing-session",
			wantStatus: http.StatusUnauthorized,
			wantField:  [2]string{"error.code", "AUTH_UNAUTHORIZED"},
		},
		{
			name:       "missing-resume",
			method:     http.MethodPost,
			path:       "/api/v1/practice/plans",
			scenario:   "missing-resume",
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  [2]string{"error.code", "VALIDATION_FAILED"},
		},
		{
			name:       "report-generating",
			method:     http.MethodGet,
			path:       "/api/v1/reports/01918fa0-0000-7000-8000-000000007000",
			scenario:   "report-generating",
			wantStatus: http.StatusOK,
			wantField:  [2]string{"status", "generating"},
		},
		{
			name:       "privacy-delete-requested",
			method:     http.MethodPost,
			path:       "/api/v1/privacy/deletions",
			scenario:   "privacy-delete-requested",
			wantStatus: http.StatusAccepted,
			wantField:  [2]string{"job.jobType", "privacy_delete"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				request := httptest.NewRequest(tc.method, tc.path, nil)
				request.Header.Set("Prefer", "example="+tc.scenario)
				response := httptest.NewRecorder()

				handler.ServeHTTP(response, request)

				if response.Code != tc.wantStatus {
					t.Fatalf("status = %d, want %d; body=%s", response.Code, tc.wantStatus, response.Body.String())
				}
				assertJSONField(t, response.Body.Bytes(), tc.wantField[0], tc.wantField[1])
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
	defaultScenario := fixture.Scenarios["default"]
	return defaultScenario.Response.Status, defaultScenario.Response.Body
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

func assertJSONField(t *testing.T, body []byte, path string, want string) {
	t.Helper()
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("parse response body: %v\n%s", err, body)
	}
	got, ok := lookupJSONPath(decoded, path)
	if !ok {
		t.Fatalf("path %s not found in %s", path, body)
	}
	if got != want {
		t.Fatalf("path %s = %q, want %q", path, got, want)
	}
}

func lookupJSONPath(value any, path string) (string, bool) {
	cursor := value
	for _, part := range strings.Split(path, ".") {
		obj, ok := cursor.(map[string]any)
		if !ok {
			return "", false
		}
		cursor, ok = obj[part]
		if !ok {
			return "", false
		}
	}
	got, ok := cursor.(string)
	return got, ok
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
