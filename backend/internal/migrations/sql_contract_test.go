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
		"file_objects",
		"resume_assets",
		"target_jobs",
		"target_job_requirements",
		"practice_plans",
		"idempotency_records",
		"practice_sessions",
		"practice_session_events",
		"practice_messages",
		"feedback_reports",
		"resume_tailor_runs",
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
		t.Fatalf("baseline migration must not create out-of-scope mistake_entries table")
	}
	if strings.Contains(up, "create table target_job_sources ") {
		t.Fatalf("paste-only baseline must not create target_job_sources")
	}
	for _, required := range []string{
		"open_question_issue_count integer not null default 0",
		"dimension_assessments jsonb not null default '[]'::jsonb",
		"retry_focus_competency_codes text[] not null default '{}'::text[]",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("baseline migration missing product-scope v1.2 column/check %q", required)
		}
	}
	for _, outOfScope := range []string{
		"open_mistake_count",
		"written_to_mistake_book",
		"single_drill",
		"core_interview",
		"fix_mistake",
		"counter_questions",
	} {
		if strings.Contains(up, outOfScope) {
			t.Fatalf("baseline migration still contains out-of-scope token %q", outOfScope)
		}
	}
}

func TestTargetJobPasteOnlyMigrationNetState(t *testing.T) {
	root := repoRoot(t)
	allUp := strings.ToLower(readAllUpMigrations(t, filepath.Join(root, "migrations")))
	baseline := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	enumSources := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "enum-sources.yaml")))

	for _, removed := range []string{
		"target_job_attachment",
		"create table target_job_sources",
		"source_file_object_id",
		"source_refresh",
	} {
		if strings.Contains(allUp, removed) || strings.Contains(enumSources, removed) {
			t.Fatalf("paste-only migration surface still contains %q", removed)
		}
	}
	targetStart := strings.Index(baseline, "create table target_jobs")
	targetEnd := strings.Index(baseline, "create table target_job_requirements")
	if targetStart < 0 || targetEnd <= targetStart {
		t.Fatal("could not isolate target_jobs baseline DDL")
	}
	targetDDL := baseline[targetStart:targetEnd]
	for _, removed := range []string{"source_type", "source_url"} {
		if strings.Contains(targetDDL, removed) {
			t.Fatalf("target_jobs still contains %q", removed)
		}
	}
	for _, preserved := range []string{
		"raw_jd_text text",
		"create table source_records",
		"'resume'",
		"'privacy_export'",
	} {
		if !strings.Contains(baseline, preserved) {
			t.Fatalf("paste-only baseline must preserve %q", preserved)
		}
	}
	for _, preserved := range []string{
		"table: source_records",
		"table: file_objects",
		"resume",
		"privacy_export",
	} {
		if !strings.Contains(enumSources, preserved) {
			t.Fatalf("enum sources must preserve %q", preserved)
		}
	}
}

func TestTargetJobReportPointerRemovedMigrationContract(t *testing.T) {
	root := repoRoot(t)
	baseline := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	targetJobs := tableBlock(t, baseline, "target_jobs")
	feedbackReports := tableBlock(t, baseline, "feedback_reports")
	groundedReportContext := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000018_grounded_report_context.up.sql")))

	if strings.Contains(targetJobs, "latest_report_id") {
		t.Fatal("target_jobs must not retain the denormalized latest_report_id pointer")
	}
	for _, required := range []string{
		"target_job_id uuid not null references target_jobs(id) on delete cascade",
		"generated_at timestamptz",
		"created_at timestamptz not null default now()",
	} {
		if !strings.Contains(feedbackReports, required) {
			t.Fatalf("feedback_reports must preserve canonical report history field %q", required)
		}
	}
	if !strings.Contains(groundedReportContext, "add column generation_context jsonb not null default '{}'::jsonb") {
		t.Fatal("grounded report context migration must preserve feedback_reports.generation_context")
	}
}

func TestUserSettingsDisplayPreferencesPruningMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000020_remove_user_settings_display_preferences.up.sql")))
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000020_remove_user_settings_display_preferences.down.sql")))

	for _, removed := range []string{
		"drop column ui_language",
		"drop column preferred_practice_language",
		"drop column region",
		"drop column timezone",
	} {
		if !strings.Contains(up, removed) {
			t.Fatalf("user_settings display-preference up migration missing %q", removed)
		}
	}
	for _, forbidden := range []string{
		"drop column user_id",
		"drop column analytics_opt_in",
		"drop column created_at",
		"drop column updated_at",
		"create table",
		"create view",
	} {
		if strings.Contains(up, forbidden) {
			t.Fatalf("user_settings display-preference up migration must not contain %q", forbidden)
		}
	}
	for _, restored := range []string{
		"add column ui_language text not null default 'zh-cn'",
		"add column preferred_practice_language text not null default 'en'",
		"add column region text",
		"add column timezone text not null default 'utc'",
	} {
		if !strings.Contains(down, restored) {
			t.Fatalf("user_settings display-preference down migration missing %q", restored)
		}
	}
}

func TestPracticeIdempotencyMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readAllUpMigrations(t, filepath.Join(root, "migrations")))
	outOfScopePracticeMode := "debrief" + "_replay"

	for _, required := range []string{
		"create table idempotency_records",
		"user_id uuid not null references users(id) on delete cascade",
		"unique (user_id, domain, operation, idempotency_key_hash)",
		"create index idx_idempotency_records_expires_at on idempotency_records (expires_at)",
		"create table practice_messages",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("practice idempotency migration contract missing %q", required)
		}
	}
	if strings.Contains(up, outOfScopePracticeMode) {
		t.Fatalf("practice migrations must not accept out-of-scope mode %s", outOfScopePracticeMode)
	}
}

func TestPracticeReplyStatusMigrationContract(t *testing.T) {
	root := repoRoot(t)
	baseline := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	enumSources := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "enum-sources.yaml")))
	block := tableBlock(t, baseline, "practice_messages")

	for _, required := range []string{
		"reply_status text",
		"reply_status in ('pending', 'retryable_failed', 'terminal_failed', 'complete')",
		"role = 'user' and client_message_id is not null and reply_to_message_id is null and reply_status is not null",
		"role = 'assistant' and client_message_id is null and reply_status is null",
		"unique (session_id, client_message_id)",
		"unique (reply_to_message_id)",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("practice reply-status migration contract missing %q", required)
		}
	}
	for _, required := range []string{
		"table: practice_messages",
		"column: reply_status",
		"values: [pending, retryable_failed, terminal_failed, complete]",
	} {
		if !strings.Contains(enumSources, required) {
			t.Fatalf("practice reply-status enum source missing %q", required)
		}
	}
}

func TestPracticeReplyLeaseGenerationMigrationContract(t *testing.T) {
	root := repoRoot(t)
	baseline := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	block := tableBlock(t, baseline, "practice_messages")

	for _, required := range []string{
		"reply_generation bigint",
		"reply_lease_expires_at timestamptz",
		"reply_generation is not null and reply_generation > 0",
		"reply_status = 'pending' and reply_lease_expires_at is not null",
		"reply_status in ('retryable_failed', 'terminal_failed', 'complete') and reply_lease_expires_at is null",
		"role = 'assistant' and client_message_id is null and reply_status is null and reply_generation is null and reply_lease_expires_at is null",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("practice reply lease/generation migration contract missing %q", required)
		}
	}
	if strings.Contains(block, "now() + interval '90 seconds'") || strings.Contains(block, "current_timestamp + interval '90 seconds'") {
		t.Fatal("practice reply lease must use the service clock, not a database clock")
	}
}

func TestBaselinePracticePlansDerivedSourceContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))
	block := tableBlock(t, up, "practice_plans")

	for _, required := range []string{
		"source_report_id uuid",
		"goal text not null check (goal in ('baseline', 'retry_current_round', 'next_round'))",
		"source_report_id is null",
		"source_report_id is not null",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("practice_plans source contract missing %q", required)
		}
	}
	for _, required := range []string{
		"foreign key (source_report_id) references feedback_reports(id) on delete set null",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("practice_plans source FK missing %q", required)
		}
	}
}

func TestPracticePlanRoundIdentityMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000017_practice_plan_round_identity.up.sql")))
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000017_practice_plan_round_identity.down.sql")))

	for _, required := range []string{
		"add column round_id text",
		"add column round_sequence integer",
		"constraint practice_plans_round_identity_pair_check",
		"round_id is null and round_sequence is null",
		"round_id is not null and round_sequence is not null",
		"btrim(round_id) <> ''",
		"constraint practice_plans_round_sequence_positive_check",
		"round_sequence is null or round_sequence > 0",
		"create index idx_practice_plans_target_job_round_created",
		"on practice_plans (user_id, target_job_id, status, round_sequence, round_id, created_at desc, id desc)",
		"where round_id is not null and round_sequence is not null",
		"create index idx_practice_sessions_plan_user_target",
		"on practice_sessions (plan_id, user_id, target_job_id)",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("practice plan round identity up migration missing %q", required)
		}
	}

	for _, required := range []string{
		"drop index if exists idx_practice_plans_target_job_round_created",
		"drop index if exists idx_practice_sessions_plan_user_target",
		"drop constraint if exists practice_plans_round_sequence_positive_check",
		"drop constraint if exists practice_plans_round_identity_pair_check",
		"drop column if exists round_sequence",
		"drop column if exists round_id",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("practice plan round identity down migration missing %q", required)
		}
	}

	for _, forbidden := range []string{
		"add column practice_progress",
		"add column current_round",
		"add column completed_round",
	} {
		if strings.Contains(up, forbidden) {
			t.Fatalf("round progress must be projected from backend facts, migration contains %q", forbidden)
		}
	}
}

func TestPracticePlanRoundIdentityBackfillIsManifestRegistered(t *testing.T) {
	root := repoRoot(t)
	manifest := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "backfill", "manifest.yaml")))
	main := readFile(t, filepath.Join(root, "backend", "cmd", "migrate", "main.go"))

	for _, required := range []string{
		"version: 17",
		"name: practice_plan_round_identity",
		"checksum: sha256:",
		"dryrun: true",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("practice plan round identity backfill manifest missing %q", required)
		}
	}
	if !strings.Contains(main, `_ "github.com/monshunter/easyinterview/backend/internal/migrations/backfills/v000017"`) {
		t.Fatal("cmd/migrate must blank-import the v000017 backfill registration")
	}

	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000017_practice_plan_round_identity.up.sql")))
	if strings.Contains(up, "update practice_plans") {
		t.Fatal("row-level legacy backfill must run through the Go registry, not migration SQL")
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

func TestBaselineMigrationAcceptsConversationTaskTypes(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	if !strings.Contains(up, "task_type in ('jd_parse', 'resume_parse', 'practice_chat', 'report_generate', 'resume_tailor')") {
		t.Fatalf("ai_task_runs.task_type CHECK must match the conversation-level task set")
	}
	for _, stale := range []string{"question_generate", "followup_generate", "report_assessment", "hint_generate"} {
		if strings.Contains(up, stale) {
			t.Fatalf("ai_task_runs.task_type CHECK must not contain stale task %s", stale)
		}
	}
}

func TestBaselinePracticeSessionEventsOnlyStoreLifecycleFacts(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	for _, required := range []string{"'session_started'", "'session_completed'"} {
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
		"dimension_assessments jsonb not null default '[]'::jsonb",
		"retry_focus_competency_codes text[] not null default '{}'::text[]",
	} {
		if !strings.Contains(block, required) {
			t.Fatalf("feedback_reports missing provenance persistence column %q", required)
		}
	}
}

func TestGroundedReportContextMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000018_grounded_report_context.up.sql")))
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000018_grounded_report_context.down.sql")))

	for _, required := range []string{
		"alter table feedback_reports",
		"add column summary text",
		"add column generation_context jsonb not null default '{}'::jsonb",
		"rename column retry_focus_competency_codes to retry_focus_dimension_codes",
		"alter table practice_plans",
		"rename column focus_competency_codes to focus_dimension_codes",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("grounded report context up migration missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"summary text not null",
		"add column retry_focus_dimension_codes",
		"add column focus_dimension_codes",
		"report-context.v1",
		"update feedback_reports",
		"create trigger",
		"audit_events",
		"async_jobs",
		"outbox_events",
		"llm_attempt_count",
	} {
		if strings.Contains(up, forbidden) {
			t.Fatalf("grounded report context up migration must not contain %q", forbidden)
		}
	}

	for _, required := range []string{
		"rename column retry_focus_dimension_codes to retry_focus_competency_codes",
		"rename column focus_dimension_codes to focus_competency_codes",
		"drop column if exists generation_context",
		"drop column if exists summary",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("grounded report context down migration missing %q", required)
		}
	}
	if strings.Contains(down, "llm_attempt_count") {
		t.Fatal("grounded report context down migration must not contain product retry attempt persistence")
	}
}

func TestReportAndPracticeV020ActivationMigrationContract(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000019_activate_report_and_practice_prompt_rubric_v020.up.sql")))
	down := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000019_activate_report_and_practice_prompt_rubric_v020.down.sql")))

	for _, required := range []string{
		"begin;",
		"insert into prompt_versions",
		"insert into rubric_versions",
		"'report.generate', 'v0.2.0', 'multi'",
		"'practice.session.chat', 'v0.2.0', 'multi'",
		"update prompt_versions",
		"set is_active = (version = 'v0.2.0')",
		"update rubric_versions",
		"where feature_key in ('report.generate', 'practice.session.chat')",
		"activation invariant",
		"commit;",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("v0.2 activation up migration missing %q", required)
		}
	}
	for _, required := range []string{
		"begin;",
		"set is_active = (version = 'v0.1.0')",
		"delete from prompt_versions",
		"delete from rubric_versions",
		"version = 'v0.2.0'",
		"rollback invariant",
		"commit;",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("v0.2 activation down migration missing %q", required)
		}
	}
	if strings.Contains(up, "update prompt_versions set template_body") || strings.Contains(up, "update rubric_versions set schema_json") {
		t.Fatal("v0.2 activation must not mutate immutable version content")
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
	// retain the out-of-scope value.
	guidedCleanup := strings.Index(up, "where source_type = 'guided'")
	newCheck := strings.Index(up, "add constraint resumes_source_type_check")
	if guidedCleanup < 0 || newCheck < 0 || guidedCleanup > newCheck {
		t.Fatalf("resume flatten up migration must clean guided rows before adding source_type check")
	}
	if strings.Contains(up[newCheck:], "'guided'") {
		t.Fatalf("resume flatten source_type check must not keep out-of-scope 'guided' value")
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

func TestDropJDMatchMigrationDeletesOutOfScopeAsyncJobsBeforeNarrowingCheck(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000014_drop_jd_match_module.up.sql")))

	deleteJobs := strings.Index(up, "delete from async_jobs")
	addCheck := strings.Index(up, "add constraint async_jobs_job_type_check")
	if deleteJobs < 0 {
		t.Fatalf("jd-match drop migration must delete out-of-scope async_jobs rows before narrowing job_type")
	}
	if addCheck < 0 {
		t.Fatalf("jd-match drop migration missing async_jobs_job_type_check")
	}
	if deleteJobs > addCheck {
		t.Fatalf("jd-match async_jobs cleanup must run before narrowed job_type check")
	}
	for _, outOfScope := range []string{"jd_match_agent_scan", "jd_match_search"} {
		if !strings.Contains(up[:addCheck], outOfScope) {
			t.Fatalf("jd-match async_jobs cleanup missing out-of-scope job_type %q", outOfScope)
		}
		if strings.Contains(up[addCheck:], outOfScope) {
			t.Fatalf("narrowed async_jobs job_type check must not keep out-of-scope value %q", outOfScope)
		}
	}
}

func TestDropJDMatchMigrationDropsOutOfScopeTablesAndRegistryRows(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000014_drop_jd_match_module.up.sql")))

	addCheck := strings.Index(up, "add constraint async_jobs_job_type_check")
	if addCheck < 0 {
		t.Fatalf("jd-match drop migration missing async_jobs_job_type_check")
	}

	for _, table := range []string{
		"jd_match_search_runs",
		"agent_scans",
		"saved_searches",
		"watchlist_items",
		"jd_match_recommendations",
	} {
		dropTable := "drop table if exists " + table
		dropIndex := strings.Index(up, dropTable)
		if dropIndex < 0 {
			t.Fatalf("jd-match drop migration must remove out-of-scope table %s", table)
		}
		if dropIndex > addCheck {
			t.Fatalf("jd-match table cleanup for %s must run before narrowed async_jobs job_type check", table)
		}
	}

	for _, registryTable := range []string{"rubric_versions", "prompt_versions"} {
		deleteIndex := strings.Index(up, "delete from "+registryTable)
		if deleteIndex < 0 {
			t.Fatalf("jd-match drop migration must delete out-of-scope rows from %s", registryTable)
		}
		if deleteIndex > addCheck {
			t.Fatalf("jd-match %s cleanup must run before narrowed async_jobs job_type check", registryTable)
		}
	}
	for _, featureKey := range []string{"jd_match.recommendation", "jd_match.search"} {
		if !strings.Contains(up[:addCheck], featureKey) {
			t.Fatalf("jd-match registry cleanup missing out-of-scope feature key %q", featureKey)
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
