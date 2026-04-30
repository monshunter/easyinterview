package main

import (
	"os"
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

func TestGeneratorEntrypointAndB1Boundary(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()

	err := Run(
		filepath.Join(repoRoot, "shared", "events.yaml"),
		filepath.Join(repoRoot, "shared", "jobs.yaml"),
		tmp,
		false,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

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
	if err := Run(
		filepath.Join(repoRoot, "shared", "events.yaml"),
		filepath.Join(repoRoot, "shared", "jobs.yaml"),
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

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
		"SourceType",
		"TargetImportSourceType",
	} {
		if !strings.Contains(events, want) {
			t.Errorf("events.go missing %q", want)
		}
	}

	jobs := readFile(t, filepath.Join(tmp, "backend/internal/shared/jobs/jobs.go"))
	for _, want := range []string{
		`JobTypeTargetImport`,
		`= "target_import"`,
		`AsynqTaskEmailDispatch`,
		`= "email.dispatch"`,
		"var APIFacingJobTypes = []JobType{",
		"JobTypePrivacyDelete",
		"var EmailDispatchRedactedFields = []string{",
	} {
		if !strings.Contains(jobs, want) {
			t.Errorf("jobs.go missing %q", want)
		}
	}
}

func TestGenerateTSOutputs(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()
	if err := Run(
		filepath.Join(repoRoot, "shared", "events.yaml"),
		filepath.Join(repoRoot, "shared", "jobs.yaml"),
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	envelope := readFile(t, filepath.Join(tmp, "frontend/src/lib/events/envelope.ts"))
	for _, want := range []string{
		"export type Producer =",
		"export interface EventEnvelope<TPayload>",
		"traceId?: string;",
	} {
		if !strings.Contains(envelope, want) {
			t.Errorf("envelope.ts missing %q", want)
		}
	}

	events := readFile(t, filepath.Join(tmp, "frontend/src/lib/events/events.ts"))
	for _, want := range []string{
		"import type {",
		"export type TargetImportSourceType =",
		`export const EVENT_NAME_REPORT_GENERATED = "report.generated" as const;`,
		"export interface ReportGeneratedPayload",
		"preparednessLevel: ReadinessTier;",
		"export interface EventNameToPayload",
	} {
		if !strings.Contains(events, want) {
			t.Errorf("events.ts missing %q", want)
		}
	}

	jobs := readFile(t, filepath.Join(tmp, "frontend/src/lib/jobs/jobs.ts"))
	for _, want := range []string{
		`export const JOB_TYPE_TARGET_IMPORT = "target_import" as const;`,
		`export const ASYNQ_TASK_EMAIL_DISPATCH = "email.dispatch" as const;`,
		"export const API_FACING_JOB_TYPES = [",
		"export const EMAIL_DISPATCH_REDACTED_FIELDS = [",
	} {
		if !strings.Contains(jobs, want) {
			t.Errorf("jobs.ts missing %q", want)
		}
	}
}

func TestGenerateJSONSchemas(t *testing.T) {
	repoRoot := repoRootForTest(t)
	tmp := t.TempDir()
	if err := Run(
		filepath.Join(repoRoot, "shared", "events.yaml"),
		filepath.Join(repoRoot, "shared", "jobs.yaml"),
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	schemaDir := filepath.Join(tmp, "shared/events/schemas")
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		t.Fatalf("ReadDir schemas: %v", err)
	}
	if got, want := len(entries), 18; got != want {
		t.Fatalf("schema file count = %d, want %d", got, want)
	}

	reportGenerated := readFile(t, filepath.Join(schemaDir, "report.generated.v1.json"))
	for _, want := range []string{
		`"$schema": "https://json-schema.org/draft/2020-12/schema"`,
		`"const": "report.generated"`,
		`"mistakeCount"`,
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
	sourceTypeRef := readFile(t, filepath.Join(tmp, "shared/events/refs/TargetImportSourceType.json"))
	if !strings.Contains(sourceTypeRef, `"file"`) {
		t.Fatalf("event-local enum ref must include values, got:\n%s", sourceTypeRef)
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
	if err := Run(
		filepath.Join(repoRoot, "shared", "events.yaml"),
		filepath.Join(repoRoot, "shared", "jobs.yaml"),
		tmp,
		false,
	); err != nil {
		t.Fatalf("Run: %v", err)
	}

	eventsBaseline := readFile(t, filepath.Join(tmp, "shared/events/baseline/events.v1.json"))
	for _, want := range []string{
		`"name": "report.generated"`,
		`"mistakeCount"`,
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
