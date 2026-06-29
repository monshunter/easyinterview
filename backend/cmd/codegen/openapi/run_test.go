package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRun_Idempotent verifies that two consecutive `Run` invocations produce
// byte-identical generated files. This is what `make codegen-check` relies on
// to detect drift via `git diff --exit-code`.
func TestRun_Idempotent(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()

	// Mirror the relevant inputs into a writable temp tree so the test does
	// not modify the repo's openapi.yaml on disk.
	openapiSrc := filepath.Join(repoRoot, "openapi", "openapi.yaml")
	openapiDst := filepath.Join(tmp, "openapi", "openapi.yaml")
	if err := os.MkdirAll(filepath.Dir(openapiDst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustCopy(t, openapiSrc, openapiDst)

	conventionsPath := filepath.Join(repoRoot, "shared", "conventions.yaml")
	templatesDir := filepath.Join(repoRoot, "openapi", "templates")
	mirrorTemplates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), mirrorTemplates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}
	_ = templatesDir // unused; mirror serves the test

	// First run.
	if err := Run(openapiDst, conventionsPath, mirrorTemplates, tmp, false); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	hashes1 := snapshotHashes(t, tmp)

	// Second run.
	if err := Run(openapiDst, conventionsPath, mirrorTemplates, tmp, false); err != nil {
		t.Fatalf("second Run: %v", err)
	}
	hashes2 := snapshotHashes(t, tmp)

	for path, h1 := range hashes1 {
		if h2 := hashes2[path]; h1 != h2 {
			t.Errorf("non-idempotent output for %s: %s vs %s", path, h1, h2)
		}
	}
}

// TestRun_DriftPropagatesFromConventions verifies that adding a new enum
// value to the loaded Conventions changes the openapi.yaml B1-AUTO block.
// We exercise the rewrite path (not the YAML on disk) by feeding a synthetic
// conventions struct into `syncB1AutoBlock` directly.
func TestRun_DriftPropagatesFromConventions(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "openapi.yaml")
	mustCopy(t, filepath.Join(repoRoot, "openapi", "openapi.yaml"), dst)

	original, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	// Re-run the canonical sync; output should be byte-stable.
	conv, err := loadConventions(filepath.Join(repoRoot, "shared", "conventions.yaml"))
	if err != nil {
		t.Fatalf("loadConventions: %v", err)
	}
	if err := syncB1AutoBlock(dst, conv); err != nil {
		t.Fatalf("sync 1: %v", err)
	}
	stable, _ := os.ReadFile(dst)
	if string(original) != string(stable) {
		t.Fatalf("first sync produced drift even though conventions.yaml is unchanged")
	}

	// Mutate the conventions struct (one of the B1 enums) and verify the
	// next sync rewrites openapi.yaml.
	for i := range conv.Enums {
		if conv.Enums[i].Name == "QuestionReviewStatus" {
			conv.Enums[i].Values = append(conv.Enums[i].Values, "x_test_drift_value")
			break
		}
	}
	if err := syncB1AutoBlock(dst, conv); err != nil {
		t.Fatalf("sync 2: %v", err)
	}
	drifted, _ := os.ReadFile(dst)
	if string(stable) == string(drifted) {
		t.Fatal("expected drift after mutating QuestionReviewStatus values")
	}
}

func TestRun_ApiErrorInnerObjectAndResponseEnvelope(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()

	openapiSrc := filepath.Join(repoRoot, "openapi", "openapi.yaml")
	openapiDst := filepath.Join(tmp, "openapi", "openapi.yaml")
	mustCopy(t, openapiSrc, openapiDst)

	mirrorTemplates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), mirrorTemplates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}

	if err := Run(
		openapiDst,
		filepath.Join(repoRoot, "shared", "conventions.yaml"),
		mirrorTemplates,
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	openapiBytes, err := os.ReadFile(openapiDst)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	openapiText := string(openapiBytes)
	mustContain(t, openapiText, "    ApiError:\n      type: object\n      required: [code, message, requestId, retryable]")
	mustContain(t, openapiText, "    ApiErrorResponse:\n      type: object\n      required: [error]")
	mustContain(t, openapiText, "          $ref: '#/components/schemas/ApiError'")

	goTypes := readFile(t, filepath.Join(tmp, "backend/internal/api/generated/types.gen.go"))
	mustContain(t, goTypes, "type ApiError = sharederrors.APIError")
	mustContain(t, goTypes, "type ApiErrorResponse struct {")
	mustContain(t, goTypes, "Error ApiError `json:\"error\"`")
	mustNotContain(t, goTypes, "type ApiError struct {\n\tError any `json:\"error\"`")

	tsTypes := readFile(t, filepath.Join(tmp, "frontend/src/api/generated/types.ts"))
	mustContain(t, tsTypes, "export type ApiError = ApiErrorAlias;")
	mustContain(t, tsTypes, "export interface ApiErrorResponse {")
	mustContain(t, tsTypes, "\terror: ApiError;")

	tsClient := readFile(t, filepath.Join(tmp, "frontend/src/api/generated/client.ts"))
	mustContain(t, tsClient, "async requestPrivacyExport(opts?: RequestOptions): Promise<Types.ApiErrorResponse>")
	mustContain(t, tsClient, "async listResumes(opts?: RequestOptions): Promise<Types.PaginatedResume>")
		mustContain(t, tsClient, "if (!response.ok && !okStatuses.includes(response.status))")
		mustContain(t, tsClient, "const text = await response.text()")
		mustContain(t, tsClient, "if (response.status === 204 || text.trim() === \"\")")
	mustContain(t, tsClient, "async createPracticeVoiceTurn(sessionId: string, body: Types.CreatePracticeVoiceTurnRequest, opts?: RequestOptions): Promise<Types.PracticeVoiceTurnResult>")

	goServer := readFile(t, filepath.Join(tmp, "backend/internal/api/generated/server.gen.go"))
	mustContain(t, goServer, "// 35-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
	mustNotContain(t, goServer, "// 43-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
	mustNotContain(t, goServer, "// 59-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
}

func mustFindRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	cur := wd
	for cur != "/" {
		if _, err := os.Stat(filepath.Join(cur, "go.work")); err == nil {
			return cur
		}
		if _, err := os.Stat(filepath.Join(cur, "openapi", "openapi.yaml")); err == nil {
			return cur
		}
		cur = filepath.Dir(cur)
	}
	t.Fatalf("could not locate repo root from %s", wd)
	return ""
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q", needle)
	}
}

func mustNotContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Fatalf("expected output not to contain %q", needle)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mustCopy(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func mirrorDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

func snapshotHashes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[path] = sha256hex(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}
