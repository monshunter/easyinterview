package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repoRootForTest(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", "..", ".."))
}

func runGenerator(t *testing.T, repoRoot, outputRoot string) {
	t.Helper()
	if err := RunWithConventions(
		filepath.Join(repoRoot, "shared", "events.yaml"),
		filepath.Join(repoRoot, "shared", "jobs.yaml"),
		filepath.Join(repoRoot, "shared", "conventions.yaml"),
		outputRoot,
		false,
	); err != nil {
		t.Fatalf("RunWithConventions: %v", err)
	}
}

func TestGeneratorEntrypointAndB1Boundary(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()

	runGenerator(t, repoRoot, tmp)

	src := readFile(t, filepath.Join(tmp, "backend/internal/shared/events/events.go"))
	if !strings.Contains(src, `sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"`) {
		t.Fatalf("events.go must import B1 generated shared types, got:\n%s", src)
	}

	b1Generator := readFile(t, filepath.Join(repoRoot, "backend/cmd/codegen/conventions/render.go"))
	if strings.Contains(b1Generator, "shared/events.yaml") || strings.Contains(b1Generator, "shared/jobs.yaml") {
		t.Fatalf("B3 event generator must not be merged into B1 conventions generator")
	}
}

func TestGenerateGoOutputs(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()
	runGenerator(t, repoRoot, tmp)

	envelope := readFile(t, filepath.Join(tmp, "backend/internal/shared/events/envelope.go"))
	for _, want := range []string{
		"package events",
		"type EventName string",
		"type Producer string",
		"type Envelope struct",
		"EventID",
		"TraceID",
		"*string",
		"Payload",
		"json.RawMessage",
	} {
		if !strings.Contains(envelope, want) {
			t.Errorf("envelope.go missing %q", want)
		}
	}

	events := readFile(t, filepath.Join(tmp, "backend/internal/shared/events/events.go"))
	for _, want := range []string{
		`EventNameReportGenerated`,
		`= "report.generated"`,
		"type TargetParsedPayload struct",
		"AnalysisStatus",
		"sharedtypes.TargetJobParseStatus",
		"PreparednessLevel",
		"sharedtypes.ReadinessTier",
	} {
		if !strings.Contains(events, want) {
			t.Errorf("events.go missing %q", want)
		}
	}
	for _, stale := range []string{"PracticeTurnCompleted", "PracticeMode", "TurnCount", "QuestionIssueCount", "TargetImportSourceType", "SourceFreshnessStatus", "SourceRefreshedPayload"} {
		if strings.Contains(events, stale) {
			t.Errorf("events.go contains stale practice question contract %q", stale)
		}
	}

	jobs := readFile(t, filepath.Join(tmp, "backend/internal/shared/jobs/jobs.go"))
	for _, want := range []string{
		`JobTypeTargetImport`,
		`= "target_import"`,
		`AsynqTaskEmailDispatch`,
		`= "email.dispatch"`,
		`JobTriggerEventSemanticSourceEventOnly`,
		`= "source_event_only"`,
		`JobTriggerEventSemanticTriggerCreatesJob`,
		`var JobTriggerEventSemantics = map[JobType]JobTriggerEventSemantic{`,
		`JobTypeReportGenerate: JobTriggerEventSemanticSourceEventOnly`,
		`func IsSourceEventOnly(jobType JobType) bool`,
		"var APIFacingJobTypes = []JobType{",
		"JobTypePrivacyDelete",
		"var EmailDispatchRedactedFields = []string{",
		"func BuildEmailDispatchPayload(",
	} {
		if !strings.Contains(jobs, want) {
			t.Errorf("jobs.go missing %q", want)
		}
	}
}

func TestGenerateTSOutputs(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()
	runGenerator(t, repoRoot, tmp)

	envelope := readFile(t, filepath.Join(tmp, "frontend/src/lib/events/envelope.ts"))
	for _, want := range []string{
		"export type Producer =",
		"export interface EventEnvelope<TPayload>",
		"traceId?: string;",
		"export function validateEnvelopeForPublish",
	} {
		if !strings.Contains(envelope, want) {
			t.Errorf("envelope.ts missing %q", want)
		}
	}

	events := readFile(t, filepath.Join(tmp, "frontend/src/lib/events/events.ts"))
	for _, want := range []string{
		"import type {",
		`export const EVENT_NAME_REPORT_GENERATED = "report.generated" as const;`,
		"export interface ReportGeneratedPayload",
		"preparednessLevel: ReadinessTier;",
		"export interface EventNameToPayload",
	} {
		if !strings.Contains(events, want) {
			t.Errorf("events.ts missing %q", want)
		}
	}
	for _, stale := range []string{"PracticeTurnCompleted", "PracticeMode", "turnCount", "questionIssueCount", "TargetImportSourceType", "SourceFreshnessStatus", "SourceRefreshedPayload"} {
		if strings.Contains(events, stale) {
			t.Errorf("events.ts contains stale practice question contract %q", stale)
		}
	}

	jobs := readFile(t, filepath.Join(tmp, "frontend/src/lib/jobs/jobs.ts"))
	for _, want := range []string{
		`export const JOB_TYPE_TARGET_IMPORT = "target_import" as const;`,
		`export const ASYNQ_TASK_EMAIL_DISPATCH = "email.dispatch" as const;`,
		`export const JOB_TRIGGER_EVENT_SEMANTIC_SOURCE_EVENT_ONLY = "source_event_only" as const;`,
		`export const JOB_TRIGGER_EVENT_SEMANTICS = {`,
		`report_generate: JOB_TRIGGER_EVENT_SEMANTIC_SOURCE_EVENT_ONLY`,
		`export function isSourceEventOnly(jobType: JobType): boolean`,
		"export const API_FACING_JOB_TYPES = [",
		"export const EMAIL_DISPATCH_REDACTED_FIELDS = [",
		"export function buildEmailDispatchPayload(",
	} {
		if !strings.Contains(jobs, want) {
			t.Errorf("jobs.ts missing %q", want)
		}
	}
}

func TestGenerateJSONSchemas(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()
	runGenerator(t, repoRoot, tmp)

	schemaDir := filepath.Join(tmp, "shared/events/schemas")
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		t.Fatalf("ReadDir schemas: %v", err)
	}
	if got, want := len(entries), 12; got != want {
		t.Fatalf("schema file count = %d, want %d", got, want)
	}

	reportGenerated := readFile(t, filepath.Join(schemaDir, "report.generated.v1.json"))
	for _, want := range []string{
		`"$schema": "https://json-schema.org/draft/2020-12/schema"`,
		`"const": "report.generated"`,
		`"preparednessLevel"`,
		`"$ref": "../refs/ReadinessTier.json"`,
	} {
		if !strings.Contains(reportGenerated, want) {
			t.Errorf("report.generated schema missing %q", want)
		}
	}

	readinessRef := readFile(t, filepath.Join(tmp, "shared/events/refs/ReadinessTier.json"))
	if !strings.Contains(readinessRef, `"basically_ready"`) {
		t.Fatalf("ReadinessTier ref must include current B1 enum values, got:\n%s", readinessRef)
	}
	for _, removed := range []string{
		"TargetImportSourceType.json",
		"SourceFreshnessStatus.json",
	} {
		if _, err := os.Stat(filepath.Join(tmp, "shared/events/refs", removed)); !os.IsNotExist(err) {
			t.Fatalf("removed event-local ref %s still exists: %v", removed, err)
		}
	}
	if _, err := os.Stat(filepath.Join(schemaDir, "source.refreshed.v1.json")); !os.IsNotExist(err) {
		t.Fatalf("removed source.refreshed schema still exists: %v", err)
	}
}

func TestGeneratorReadsB1EnumRefsFromConventionsYaml(t *testing.T) {
	repoRoot := repoRootForTest(t)
	events := readFile(t, filepath.Join(repoRoot, "shared/events.yaml"))
	jobs := readFile(t, filepath.Join(repoRoot, "shared/jobs.yaml"))
	conventions := strings.Replace(
		readFile(t, filepath.Join(repoRoot, "shared/conventions.yaml")),
		"      - basically_ready\n",
		"      - experimentally_ready\n",
		1,
	)
	tmp := t.TempDir()

	if err := RunFromBytesWithConventions([]byte(events), []byte(jobs), []byte(conventions), tmp, false); err != nil {
		t.Fatalf("RunFromBytesWithConventions: %v", err)
	}

	readinessRef := readFile(t, filepath.Join(tmp, "shared/events/refs/ReadinessTier.json"))
	if !strings.Contains(readinessRef, `"experimentally_ready"`) {
		t.Fatalf("ReadinessTier ref must follow shared/conventions.yaml, got:\n%s", readinessRef)
	}
	if strings.Contains(readinessRef, `"basically_ready"`) {
		t.Fatalf("ReadinessTier ref used stale hard-coded B1 values, got:\n%s", readinessRef)
	}
}

func TestGeneratorFailsWhenB1EnumRefMissingFromConventions(t *testing.T) {
	repoRoot := repoRootForTest(t)
	events := readFile(t, filepath.Join(repoRoot, "shared/events.yaml"))
	jobs := readFile(t, filepath.Join(repoRoot, "shared/jobs.yaml"))
	conventions := strings.Replace(
		readFile(t, filepath.Join(repoRoot, "shared/conventions.yaml")),
		`
  - name: ReadinessTier
    sourceSection: "5.8"
    jsonField: readinessTier
    values:
      - not_ready
      - needs_practice
      - basically_ready
      - well_prepared
`,
		"\n",
		1,
	)
	tmp := t.TempDir()

	err := RunFromBytesWithConventions([]byte(events), []byte(jobs), []byte(conventions), tmp, false)
	if err == nil {
		t.Fatalf("RunFromBytesWithConventions succeeded, want missing B1 enum ref failure")
	}
	if !strings.Contains(err.Error(), "ReadinessTier") {
		t.Fatalf("missing enum error must name ReadinessTier, got: %v", err)
	}
}

func TestMakefileWiresCodegenEvents(t *testing.T) {
	makefile := readFile(t, filepath.Join(repoRootForTest(t), "Makefile"))

	if !strings.Contains(makefile, "codegen-events:") {
		t.Fatalf("Makefile must define codegen-events target")
	}
	if !strings.Contains(makefile, "go run ./cmd/codegen/events") {
		t.Fatalf("codegen-events target must call backend/cmd/codegen/events")
	}
	if !strings.Contains(makefile, "codegen: codegen-conventions codegen-events codegen-openapi") {
		t.Fatalf("codegen aggregate must order B1 conventions before B3 events before B2 openapi")
	}
}

func TestBaselineManifests(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()
	runGenerator(t, repoRoot, tmp)

	eventsBaseline := readFile(t, filepath.Join(tmp, "shared/events/baseline/events.v1.json"))
	for _, want := range []string{
		`"name": "report.generated"`,
		`"$ref:b1.ReadinessTier"`,
	} {
		if !strings.Contains(eventsBaseline, want) {
			t.Errorf("events baseline missing %q", want)
		}
	}

	jobsBaseline := readFile(t, filepath.Join(tmp, "shared/jobs/baseline/jobs.v1.json"))
	for _, want := range []string{
		`"canonical": "email_dispatch"`,
		`"apiFacing": false`,
		`"apiFacingSubset"`,
	} {
		if !strings.Contains(jobsBaseline, want) {
			t.Errorf("jobs baseline missing %q", want)
		}
	}
}

func TestBreakingChangeFixtures(t *testing.T) {
	repoRoot := repoRootForTest(t)
	baseEvents := readFile(t, filepath.Join(repoRoot, "shared/events.yaml"))
	baseJobs := readFile(t, filepath.Join(repoRoot, "shared/jobs.yaml"))

	tests := []struct {
		name   string
		events string
	}{
		{
			name:   "type change",
			events: strings.Replace(baseEvents, "modelId: { type: string, source: spec:3.1.4 }", "modelId: { type: bool, source: spec:3.1.4 }", 1),
		},
		{
			name:   "required field deletion",
			events: strings.Replace(baseEvents, "      modelId: { type: string, source: spec:3.1.4 }\n", "", 1),
		},
		{
			name:   "dot case to snake",
			events: strings.Replace(baseEvents, "  - name: report.generated", "  - name: report_generated", 1),
		},
		{
			name:   "enum member removal",
			events: strings.Replace(baseEvents, "      - bullet_suggestions\n", "", 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			writeLintRepo(t, repoRoot, tmp, tt.events, baseJobs)

			output, err := runLintEvents(t, repoRoot, tmp)
			if err == nil {
				t.Fatalf("lint_events succeeded, want failure; output:\n%s", output)
			}
			if !strings.Contains(output, "breaking change requires eventVersion + 1") {
				t.Fatalf("missing breaking-change message; output:\n%s", output)
			}
		})
	}
}

func TestAdditiveOptionalFieldFixture(t *testing.T) {
	repoRoot := repoRootForTest(t)
	events := strings.Replace(
		readFile(t, filepath.Join(repoRoot, "shared/events.yaml")),
		"    optionalPayload: {}\n    piiBoundary: No report body, answer snippets, raw model response, or prompt body.\n  - name: report.generation.failed",
		"    optionalPayload:\n      reviewerNote: { type: string, source: spec:3.1.4 }\n    piiBoundary: No report body, answer snippets, raw model response, or prompt body.\n  - name: report.generation.failed",
		1,
	)
	jobs := readFile(t, filepath.Join(repoRoot, "shared/jobs.yaml"))
	tmp := t.TempDir()
	conventions := readFile(t, filepath.Join(repoRoot, "shared/conventions.yaml"))
	mustWriteFile(t, filepath.Join(tmp, "shared/conventions.yaml"), conventions)
	if err := RunFromBytesWithConventions([]byte(events), []byte(jobs), []byte(conventions), tmp, false); err != nil {
		t.Fatalf("RunFromBytesWithConventions: %v", err)
	}

	goEvents := readFile(t, filepath.Join(tmp, "backend/internal/shared/events/events.go"))
	if !strings.Contains(goEvents, "ReviewerNote") ||
		!strings.Contains(goEvents, "*string") ||
		!strings.Contains(goEvents, "`json:\"reviewerNote,omitempty\"`") {
		t.Fatalf("Go optional field must be a pointer, got:\n%s", goEvents)
	}
	tsEvents := readFile(t, filepath.Join(tmp, "frontend/src/lib/events/events.ts"))
	if !strings.Contains(tsEvents, "reviewerNote?: string;") {
		t.Fatalf("TS optional field must use ?:, got:\n%s", tsEvents)
	}

	writeLintRepo(t, repoRoot, tmp, events, jobs)
	if output, err := runLintEvents(t, repoRoot, tmp); err != nil {
		t.Fatalf("lint_events additive optional failed: %v\n%s", err, output)
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

func writeLintRepo(t *testing.T, repoRoot, tmp, events, jobs string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(tmp, "shared/events.yaml"), events)
	mustWriteFile(t, filepath.Join(tmp, "shared/jobs.yaml"), jobs)
	mustWriteFile(t, filepath.Join(tmp, "shared/events/baseline/events.v1.json"), readFile(t, filepath.Join(repoRoot, "shared/events/baseline/events.v1.json")))
	mustWriteFile(t, filepath.Join(tmp, "shared/jobs/baseline/jobs.v1.json"), readFile(t, filepath.Join(repoRoot, "shared/jobs/baseline/jobs.v1.json")))
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func runLintEvents(t *testing.T, repoRoot, tmp string) (string, error) {
	t.Helper()
	cmd := exec.Command("python3", filepath.Join(repoRoot, "scripts/lint/lint_events.py"), "--repo-root", tmp)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
