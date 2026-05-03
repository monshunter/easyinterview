package migrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBaselineMigrationEnablesVectorAndKeepsDownSafe(t *testing.T) {
	root := repoRoot(t)
	up := readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql"))
	down := readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.down.sql"))

	if !strings.Contains(strings.ToLower(up), "create extension if not exists vector") {
		t.Fatalf("baseline up migration must enable vector extension idempotently")
	}
	if strings.Contains(strings.ToLower(down), "drop extension") {
		t.Fatalf("baseline down migration must not drop vector by default")
	}
}

func TestBaselineMigrationDefinesAllOwnedTables(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	for _, table := range []string{
		"users",
		"user_settings",
		"candidate_profiles",
		"experience_cards",
		"file_objects",
		"resume_assets",
		"target_jobs",
		"target_job_requirements",
		"target_job_sources",
		"practice_plans",
		"practice_sessions",
		"practice_session_events",
		"practice_turns",
		"feedback_reports",
		"question_assessments",
		"resume_tailor_runs",
		"debriefs",
		"source_records",
		"retrieval_chunks",
		"prompt_versions",
		"rubric_versions",
		"ai_task_runs",
		"async_jobs",
		"outbox_events",
		"privacy_requests",
		"audit_events",
		"auth_challenges",
		"sessions",
		"external_identities",
		"schema_backfills",
	} {
		if !strings.Contains(up, "create table "+table+" ") {
			t.Fatalf("baseline migration missing create table %s", table)
		}
	}
	if strings.Contains(up, "create table mistake_entries ") {
		t.Fatalf("baseline migration must not create removed mistake_entries table")
	}
	for _, required := range []string{
		"open_question_issue_count integer not null default 0",
		"review_status text not null check (review_status in ('open', 'queued_for_retry', 'resolved'))",
		"included_in_retry_plan boolean not null default false",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("baseline migration missing product-scope v1.2 column/check %q", required)
		}
	}
	for _, removed := range []string{
		"open_mistake_count",
		"written_to_mistake_book",
		"single_drill",
		"core_interview",
		"fix_mistake",
		"counter_questions",
	} {
		if strings.Contains(up, removed) {
			t.Fatalf("baseline migration still contains removed token %q", removed)
		}
	}
}

func TestBaselineMigrationDoesNotStoreRawAuthSecrets(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	for _, forbidden := range []string{"raw_token", "session_cookie", "api_key", "provider_token"} {
		if strings.Contains(up, forbidden) {
			t.Fatalf("baseline migration must not contain plaintext-secret column marker %q", forbidden)
		}
	}
	for _, required := range []string{"challenge_token_hash", "session_hash", "provider_subject_hash"} {
		if !strings.Contains(up, required) {
			t.Fatalf("baseline migration must contain safe hash field %q", required)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
