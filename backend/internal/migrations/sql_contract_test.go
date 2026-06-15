package migrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBaselineMigrationDoesNotEnableVectorExtension(t *testing.T) {
	root := repoRoot(t)
	up := readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql"))
	down := readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.down.sql"))

	if strings.Contains(strings.ToLower(up), "create extension if not exists vector") {
		t.Fatalf("baseline up migration must not enable vector extension")
	}
	if strings.Contains(strings.ToLower(down), "drop extension") {
		t.Fatalf("baseline down migration must not manage extensions")
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
		"idempotency_records",
		"practice_sessions",
		"practice_session_events",
		"practice_turns",
		"feedback_reports",
		"question_assessments",
		"resume_tailor_runs",
		"debriefs",
		"source_records",
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

func TestPracticeIdempotencyMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readAllUpMigrations(t, filepath.Join(root, "migrations")))
	removedPracticeMode := "debrief" + "_replay"

	for _, required := range []string{
		"create table idempotency_records",
		"user_id uuid not null references users(id) on delete cascade",
		"unique (user_id, domain, operation, idempotency_key_hash)",
		"create index idx_idempotency_records_expires_at on idempotency_records (expires_at)",
		"mode text not null check (mode in ('assisted', 'strict'))",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("practice idempotency migration contract missing %q", required)
		}
	}
	if strings.Contains(up, removedPracticeMode) {
		t.Fatalf("practice migrations must not accept removed mode %s", removedPracticeMode)
	}
}

func TestBaselinePracticePlansDerivedSourceContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	block := tableBlock(t, up, "practice_plans")

	for _, required := range []string{
		"source_report_id uuid",
		"source_debrief_id uuid",
		"goal text not null check (goal in ('baseline', 'retry_current_round', 'next_round', 'debrief'))",
		"source_report_id is null",
		"source_report_id is not null",
		"source_debrief_id is null",
		"source_debrief_id is not null",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("practice_plans source contract missing %q", required)
		}
	}
	for _, required := range []string{
		"foreign key (source_report_id) references feedback_reports(id) on delete set null",
		"foreign key (source_debrief_id) references debriefs(id) on delete set null",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("practice_plans source FK missing %q", required)
		}
	}
}

func TestPracticeIdempotencyMigrationDownDoesNotDropBaselineOwnedTable(t *testing.T) {
	root := repoRoot(t)
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000003_practice_idempotency_baseline.down.sql")))

	for _, forbidden := range []string{
		"drop index if exists idx_idempotency_records_expires_at",
		"drop table if exists idempotency_records",
	} {
		if strings.Contains(down, forbidden) {
			t.Fatalf("practice idempotency down migration must not contain %q", forbidden)
		}
	}
}

func TestBaselineMigrationAcceptsReportAndDebriefQuestionTaskTypes(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	if !strings.Contains(up, "task_type in ('jd_parse', 'resume_parse', 'question_generate', 'followup_generate', 'report_generate', 'report_assessment', 'resume_tailor', 'debrief_generate', 'debrief_suggest_questions', 'hint_generate')") {
		t.Fatalf("ai_task_runs.task_type CHECK must include report_assessment and debrief_suggest_questions")
	}
	if strings.Contains(up, "'report_assess'") {
		t.Fatalf("ai_task_runs.task_type CHECK must not contain retired report_assess alias")
	}
}

func TestBaselinePracticeSessionEventsAcceptVoicePlaybackKinds(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	for _, required := range []string{
		"'tts_chunk_started'",
		"'tts_chunk_played'",
		"'barge_in_detected'",
		"'assistant_context_committed'",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("practice_session_events.event_type check must include %s", required)
		}
	}
}

func TestFeedbackReportsContainsProvenancePersistenceColumns(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	block := tableBlock(t, up, "feedback_reports")

	for _, required := range []string{
		"language text not null default 'en'",
		"feature_flag text not null default 'none'",
		"data_source_version text not null default 'not_applicable'",
		"retry_focus_turn_ids jsonb not null default '[]'::jsonb",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("feedback_reports missing provenance persistence column %q", required)
		}
	}
}

func TestResumeVersionsAdditiveMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000005_resume_versions.up.sql")))
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000005_resume_versions.down.sql")))

	for _, required := range []string{
		"create table if not exists resume_versions",
		"user_id uuid not null references users(id)",
		"resume_asset_id uuid not null references resume_assets(id)",
		"parent_version_id uuid references resume_versions(id)",
		"version_type text not null check (version_type in ('structured_master', 'targeted'))",
		"seed_strategy text check (seed_strategy is null or seed_strategy in ('copy_master', 'blank', 'ai_select'))",
		"structured_profile jsonb not null default '{}'::jsonb",
		"create index if not exists idx_resume_versions_user_updated on resume_versions (user_id, updated_at desc)",
		"create index if not exists idx_resume_versions_asset_type on resume_versions (resume_asset_id, version_type)",
		"create index if not exists idx_resume_versions_parent on resume_versions (parent_version_id) where parent_version_id is not null",
		"create table if not exists resume_version_suggestions",
		"resume_version_id uuid not null references resume_versions(id) on delete cascade",
		"status text not null default 'pending' check (status in ('pending', 'accepted', 'rejected'))",
		"create index if not exists idx_resume_suggestions_version_status on resume_version_suggestions (resume_version_id, status)",
		"create index if not exists idx_resume_suggestions_tailor_run on resume_version_suggestions (tailor_run_id)",
		"add column if not exists source_type text check (source_type is null or source_type in ('upload', 'paste', 'guided'))",
		"add column if not exists original_text text",
		"add column if not exists guided_answers jsonb",
		"add column if not exists parsed_text_snapshot text",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("resume versions up migration missing %q", required)
		}
	}

	for _, required := range []string{
		"drop table if exists resume_version_suggestions",
		"drop table if exists resume_versions",
		"drop column if exists parsed_text_snapshot",
		"drop column if exists guided_answers",
		"drop column if exists original_text",
		"drop column if exists source_type",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("resume versions down migration missing %q", required)
		}
	}
}

func TestResumeFlattenMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000015_resume_flatten.up.sql")))
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000015_resume_flatten.down.sql")))

	for _, required := range []string{
		"alter table resume_assets rename to resumes",
		"add column if not exists structured_profile jsonb not null default '{}'::jsonb",
		"add column if not exists display_name text",
		"update resumes",
		"version_type = 'structured_master'",
		"where source_type = 'guided'",
		"drop column if exists guided_answers",
		"add constraint resumes_source_type_check",
		"check (source_type is null or source_type in ('upload', 'paste'))",
		"alter table practice_plans rename column resume_asset_id to resume_id",
		"drop table if exists resume_version_suggestions",
		"drop table if exists resume_tailor_runs",
		"drop table if exists resume_versions",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("resume flatten up migration missing %q", required)
		}
	}
	// D-20: create flow is upload/paste only. Existing guided rows must be
	// rewritten before the new check is added, but the check itself must not
	// retain the retired value.
	guidedCleanup := strings.Index(up, "where source_type = 'guided'")
	newCheck := strings.Index(up, "add constraint resumes_source_type_check")
	if guidedCleanup < 0 || newCheck < 0 || guidedCleanup > newCheck {
		t.Fatalf("resume flatten up migration must clean guided rows before adding source_type check")
	}
	if strings.Contains(up[newCheck:], "'guided'") {
		t.Fatalf("resume flatten source_type check must not keep retired 'guided' value")
	}

	for _, required := range []string{
		"alter table resumes rename to resume_assets",
		"drop column if exists structured_profile",
		"drop column if exists display_name",
		"add column if not exists guided_answers",
		"alter table practice_plans rename column resume_id to resume_asset_id",
		"create table if not exists resume_tailor_runs",
		"create table if not exists resume_versions",
		"create table if not exists resume_version_suggestions",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("resume flatten down migration missing %q", required)
		}
	}
	for _, required := range []string{
		"language text not null default 'en'",
		"feature_flag text not null default 'none'",
		"data_source_version text not null default 'not_applicable'",
	} {
		tailorRuns := down[strings.Index(down, "create table if not exists resume_tailor_runs"):]
		tailorRuns = tailorRuns[:strings.Index(tailorRuns, ");")]
		if !strings.Contains(tailorRuns, required) {
			t.Fatalf("resume flatten down migration must restore resume_tailor_runs.%s", required)
		}
	}
}

func TestDropJDMatchMigrationDeletesRetiredAsyncJobsBeforeNarrowingCheck(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000014_drop_jd_match_module.up.sql")))

	deleteJobs := strings.Index(up, "delete from async_jobs")
	addCheck := strings.Index(up, "add constraint async_jobs_job_type_check")
	if deleteJobs < 0 {
		t.Fatalf("jd-match drop migration must delete retired async_jobs rows before narrowing job_type")
	}
	if addCheck < 0 {
		t.Fatalf("jd-match drop migration missing async_jobs_job_type_check")
	}
	if deleteJobs > addCheck {
		t.Fatalf("jd-match async_jobs cleanup must run before narrowed job_type check")
	}
	for _, retired := range []string{"jd_match_agent_scan", "jd_match_search"} {
		if !strings.Contains(up[:addCheck], retired) {
			t.Fatalf("jd-match async_jobs cleanup missing retired job_type %q", retired)
		}
		if strings.Contains(up[addCheck:], retired) {
			t.Fatalf("narrowed async_jobs job_type check must not keep retired value %q", retired)
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

func readAllUpMigrations(t *testing.T, dir string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, path := range matches {
		b.WriteString(readFile(t, path))
		b.WriteString("\n")
	}
	return b.String()
}

func tableBlock(t *testing.T, sql, table string) string {
	t.Helper()
	startMarker := "create table " + table + " ("
	start := strings.Index(sql, startMarker)
	if start == -1 {
		t.Fatalf("missing create table %s", table)
	}
	rest := sql[start:]
	end := strings.Index(rest, ");")
	if end == -1 {
		t.Fatalf("missing end of create table %s", table)
	}
	return rest[:end]
}
