package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
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
	mirrorTemplates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), mirrorTemplates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}

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

func TestRun_DoesNotEmitUnusedFrontendSpecSnapshot(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()
	openapiDst := filepath.Join(tmp, "openapi", "openapi.yaml")
	if err := os.MkdirAll(filepath.Dir(openapiDst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustCopy(t, filepath.Join(repoRoot, "openapi", "openapi.yaml"), openapiDst)

	templates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), templates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}
	if err := Run(
		openapiDst,
		filepath.Join(repoRoot, "shared", "conventions.yaml"),
		templates,
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	generated := filepath.Join(tmp, "frontend", "src", "api", "generated")
	for _, name := range []string{"client.ts", "types.ts"} {
		if _, err := os.Stat(filepath.Join(generated, name)); err != nil {
			t.Fatalf("expected %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(generated, "spec.ts")); !os.IsNotExist(err) {
		t.Fatalf("unexpected frontend raw spec snapshot: %v", err)
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
		if conv.Enums[i].Name == "PracticeGoal" {
			conv.Enums[i].Values = append(conv.Enums[i].Values, "x_test_drift_value")
			break
		}
	}
	if err := syncB1AutoBlock(dst, conv); err != nil {
		t.Fatalf("sync 2: %v", err)
	}
	drifted, _ := os.ReadFile(dst)
	if string(stable) == string(drifted) {
		t.Fatal("expected drift after mutating PracticeGoal values")
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
	mustContain(t, tsClient, "async createPracticeVoiceTurn(sessionId: string, body: Types.CreatePracticeVoiceTurnRequest, opts?: RequestOptions): Promise<unknown>")

	goServer := readFile(t, filepath.Join(tmp, "backend/internal/api/generated/server.gen.go"))
	mustContain(t, goServer, "// 37-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
	mustNotContain(t, goServer, "// 35-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
	mustNotContain(t, goServer, "// 43-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
	mustNotContain(t, goServer, "// 59-row table in `docs/spec/openapi-v1-contract/spec.md` §3.1.1.")
}

func TestRun_PracticeRoundIdentityAndProgressTypes(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()
	openapiDst := filepath.Join(tmp, "openapi", "openapi.yaml")
	if err := os.MkdirAll(filepath.Dir(openapiDst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustCopy(t, filepath.Join(repoRoot, "openapi", "openapi.yaml"), openapiDst)

	templates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), templates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}
	if err := Run(
		openapiDst,
		filepath.Join(repoRoot, "shared", "conventions.yaml"),
		templates,
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	goTypes := readFile(t, filepath.Join(tmp, "backend/internal/api/generated/types.gen.go"))
	mustContain(t, goTypes, "type PracticeRoundRef struct {")
	mustMatch(t, goTypes, `(?m)^\s*RoundId\s+string\s+`+"`json:\"roundId\"`"+`$`)
	mustMatch(t, goTypes, `(?m)^\s*RoundSequence\s+int32\s+`+"`json:\"roundSequence\"`"+`$`)
	mustContain(t, goTypes, "type PracticeProgress struct {")
	mustMatch(t, goTypes, `(?m)^\s*CurrentRound\s+\*PracticeRoundRef\s+`+"`json:\"currentRound\"`"+`$`)
	mustMatch(t, goTypes, `(?m)^\s*PracticeProgress\s+\*PracticeProgress\s+`+"`json:\"practiceProgress,omitempty\"`"+`$`)
	mustMatch(t, goTypes, `(?m)^\s*RoundId\s+\*string\s+`+"`json:\"roundId,omitempty\"`"+`$`)
	mustMatch(t, goTypes, `(?m)^\s*RoundSequence\s+\*int32\s+`+"`json:\"roundSequence,omitempty\"`"+`$`)

	tsTypes := readFile(t, filepath.Join(tmp, "frontend/src/api/generated/types.ts"))
	mustContain(t, tsTypes, "export interface PracticeRoundRef {")
	mustContain(t, tsTypes, "export interface PracticeProgress {")
	mustContain(t, tsTypes, "\tcurrentRound: PracticeRoundRef | null;")
	mustContain(t, tsTypes, "\tpracticeProgress?: PracticeProgress;")
	mustContain(t, tsTypes, "\troundId?: string | null;")
	mustContain(t, tsTypes, "\troundSequence?: number | null;")
}

func TestRun_GroundedReportAndTypedDerivedPlanTypes(t *testing.T) {
	repoRoot := mustFindRepoRoot(t)
	tmp := t.TempDir()
	openapiDst := filepath.Join(tmp, "openapi", "openapi.yaml")
	if err := os.MkdirAll(filepath.Dir(openapiDst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustCopy(t, filepath.Join(repoRoot, "openapi", "openapi.yaml"), openapiDst)
	templates := filepath.Join(tmp, "openapi", "templates")
	if err := mirrorDir(filepath.Join(repoRoot, "openapi", "templates"), templates); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}
	if err := Run(
		openapiDst,
		filepath.Join(repoRoot, "shared", "conventions.yaml"),
		templates,
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	goTypes := readFile(t, filepath.Join(tmp, "backend/internal/api/generated/types.gen.go"))
	mustContain(t, goTypes, "type CreatePracticePlanRequest struct {")
	mustNotContain(t, goTypes, "type CreatePracticePlanRequest = any")
	mustContain(t, goTypes, "SourceReportId     *string")
	mustContain(t, goTypes, "type ReportContextSnapshot struct {")
	mustMatch(t, goTypes, `(?s)type FeedbackReport struct \{.*?RetryFocusDimensionCodes\s+\[\]string\s+`+"`json:\"retryFocusDimensionCodes\"`"+`.*?Summary\s+\*string\s+`+"`json:\"summary\"`"+`.*?\n\}`)
	mustNotContain(t, goTypes, "type FeedbackReport struct {\n\tItems")
	mustNotContain(t, goTypes, "RetryFocusCompetencyCodes")
	mustNotContain(t, goTypes, "FocusCompetencyCodes")
	mustNotContain(t, goTypes, "type DimensionResult")

	tsTypes := readFile(t, filepath.Join(tmp, "frontend/src/api/generated/types.ts"))
	mustContain(t, tsTypes, "export interface CreatePracticePlanRequest {")
	mustNotContain(t, tsTypes, "export type CreatePracticePlanRequest = any")
	mustContain(t, tsTypes, "\tsourceReportId?: string;")
	mustContain(t, tsTypes, "export interface ReportContextSnapshot {")
	mustMatch(t, tsTypes, `(?s)export interface FeedbackReport \{.*?\tretryFocusDimensionCodes: string\[\];.*?\tsummary: string \| null;.*?\n\}`)
	mustNotContain(t, tsTypes, "export interface FeedbackReport {\n\titems:")
	mustNotContain(t, tsTypes, "retryFocusCompetencyCodes")
	mustNotContain(t, tsTypes, "focusCompetencyCodes")
	mustNotContain(t, tsTypes, "export interface DimensionResult")
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

func mustMatch(t *testing.T, haystack, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(haystack) {
		t.Fatalf("expected output to match %q", pattern)
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
		sum := sha256.Sum256(data)
		out[path] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}
