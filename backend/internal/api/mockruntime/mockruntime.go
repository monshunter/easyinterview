package mockruntime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
)

type Registry struct {
	fixtures map[string]Fixture
}

type Fixture struct {
	OperationID string
	Scenarios   map[string]Scenario
}

type Scenario struct {
	Response FixtureResponse
}

type FixtureResponse struct {
	Status  int
	Headers map[string]string
	Body    json.RawMessage
}

func LoadRegistry(fixturesRoot string) (*Registry, error) {
	entries, err := os.ReadDir(fixturesRoot)
	if err != nil {
		return nil, err
	}
	registry := &Registry{fixtures: map[string]Fixture{}}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		tagDir := filepath.Join(fixturesRoot, entry.Name())
		files, err := os.ReadDir(tagDir)
		if err != nil {
			return nil, err
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
				continue
			}
			path := filepath.Join(tagDir, file.Name())
			fixture, err := readFixture(path)
			if err != nil {
				return nil, err
			}
			if fixture.OperationID != strings.TrimSuffix(file.Name(), ".json") {
				return nil, fmt.Errorf("%s: operationId %q does not match filename", path, fixture.OperationID)
			}
			if _, exists := registry.fixtures[fixture.OperationID]; exists {
				return nil, fmt.Errorf("%s: duplicate operationId %q", path, fixture.OperationID)
			}
			registry.fixtures[fixture.OperationID] = fixture
		}
	}
	return registry, nil
}

func readFixture(path string) (Fixture, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Fixture{}, err
	}
	var disk struct {
		OperationID string `json:"operationId"`
		Scenarios   map[string]struct {
			Response struct {
				Status  int               `json:"status"`
				Headers map[string]string `json:"headers"`
				Body    json.RawMessage   `json:"body"`
			} `json:"response"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(raw, &disk); err != nil {
		return Fixture{}, fmt.Errorf("%s: %w", path, err)
	}
	if disk.OperationID == "" {
		return Fixture{}, fmt.Errorf("%s: operationId missing", path)
	}
	fixture := Fixture{OperationID: disk.OperationID, Scenarios: map[string]Scenario{}}
	for name, scenario := range disk.Scenarios {
		fixture.Scenarios[name] = Scenario{
			Response: FixtureResponse{
				Status:  scenario.Response.Status,
				Headers: scenario.Response.Headers,
				Body:    scenario.Response.Body,
			},
		}
	}
	if _, ok := fixture.Scenarios["default"]; !ok {
		return Fixture{}, fmt.Errorf("%s: scenarios.default missing", path)
	}
	return fixture, nil
}

func NewHandler(registry *Registry) http.Handler {
	return &handler{registry: registry}
}

type handler struct {
	registry *Registry
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, ok := matchRoute(r.Method, stripBasePath(r.URL.Path))
	if !ok {
		http.NotFound(w, r)
		return
	}
	fixture, ok := h.registry.fixtures[route.OperationID]
	if !ok {
		http.Error(w, fmt.Sprintf("missing fixture for operationId: %s", route.OperationID), http.StatusInternalServerError)
		return
	}
	scenarioName := selectScenario(r.Header)
	scenario, ok := fixture.Scenarios[scenarioName]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown fixture scenario %q for operationId: %s", scenarioName, route.OperationID), http.StatusBadRequest)
		return
	}
	for key, value := range scenario.Response.Headers {
		w.Header().Set(key, value)
	}
	if len(scenario.Response.Body) > 0 {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(scenario.Response.Status)
	if len(scenario.Response.Body) > 0 && scenario.Response.Status != http.StatusNoContent {
		_, _ = w.Write(scenario.Response.Body)
	}
}

func stripBasePath(path string) string {
	if strings.HasPrefix(path, "/api/v1/") {
		return strings.TrimPrefix(path, "/api/v1")
	}
	return path
}

func matchRoute(method string, path string) (generated.Route, bool) {
	for _, route := range generated.AllRoutes {
		if !strings.EqualFold(route.Method, method) {
			continue
		}
		if pathTemplateMatches(route.Path, path) {
			return route, true
		}
	}
	return generated.Route{}, false
}

func pathTemplateMatches(template string, actual string) bool {
	templateParts := splitPath(template)
	actualParts := splitPath(actual)
	if len(templateParts) != len(actualParts) {
		return false
	}
	for i, part := range templateParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != actualParts[i] {
			return false
		}
	}
	return true
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func selectScenario(header http.Header) string {
	prefer := header.Get("Prefer")
	for _, part := range strings.Split(prefer, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "example=") {
			scenario := strings.TrimSpace(strings.TrimPrefix(part, "example="))
			if scenario != "" {
				return scenario
			}
		}
	}
	return "default"
}
