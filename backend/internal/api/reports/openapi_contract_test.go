package reports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetFeedbackReportNotFoundResponseContract(t *testing.T) {
	for _, rel := range []string{
		"openapi/openapi.yaml",
		"openapi/baseline/openapi-v1.0.0.yaml",
	} {
		t.Run(rel, func(t *testing.T) {
			doc := loadOpenAPI(t, rel)
			responses := lookupMap(t, doc, "paths", "/reports/{reportId}", "get", "responses")
			notFound := mapValue(t, responses, "404")

			description, _ := notFound["description"].(string)
			if !strings.Contains(description, "REPORT_NOT_FOUND") {
				t.Fatalf("404 description = %q, want REPORT_NOT_FOUND", description)
			}

			ref := lookupString(t, notFound, "content", "application/json", "schema", "$ref")
			if ref != "#/components/schemas/ApiErrorResponse" {
				t.Fatalf("404 schema ref = %q, want ApiErrorResponse envelope", ref)
			}

			code := lookupString(t, notFound, "content", "application/json", "examples", "reportNotFound", "value", "error", "code")
			if code != "REPORT_NOT_FOUND" {
				t.Fatalf("404 example error.code = %q, want REPORT_NOT_FOUND", code)
			}
		})
	}
}

func loadOpenAPI(t *testing.T, rel string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	return doc
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "openapi", "openapi.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}

func lookupMap(t *testing.T, root map[string]any, path ...string) map[string]any {
	t.Helper()
	current := root
	for _, key := range path {
		current = mapValue(t, current, key)
	}
	return current
}

func mapValue(t *testing.T, root map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := root[key]
	if !ok {
		t.Fatalf("missing key %q in %#v", key, root)
	}
	mapped, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("key %q has type %T, want map", key, value)
	}
	return mapped
}

func lookupString(t *testing.T, root map[string]any, path ...string) string {
	t.Helper()
	current := root
	for _, key := range path[:len(path)-1] {
		current = mapValue(t, current, key)
	}
	value, ok := current[path[len(path)-1]]
	if !ok {
		t.Fatalf("missing key %q in %#v", path[len(path)-1], current)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("key %q has type %T, want string", path[len(path)-1], value)
	}
	return text
}
