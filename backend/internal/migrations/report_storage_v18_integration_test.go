//go:build integration

package migrations

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

func TestIntegrationGroundedReportStorageV18(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL must identify an empty disposable PostgreSQL database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := repoRoot(t)
	cmd := Command{
		DatabaseURL:      dsn,
		MigrationsDir:    filepath.Join(root, "migrations"),
		BackfillManifest: filepath.Join(root, "migrations", "backfill", "manifest.yaml"),
	}
	migrator, err := newMigrate(cmd)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	defer closeMigrate(migrator)
	if _, _, err := migrator.Version(); !errors.Is(err, migrate.ErrNilVersion) {
		t.Fatalf("integration database must be empty, version error=%v", err)
	}
	if err := migrator.Migrate(17); err != nil {
		t.Fatalf("migrate clean database to v17: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	const (
		userID    = "019f5700-0000-7000-8000-000000000001"
		resumeID  = "019f5700-0000-7000-8000-000000000002"
		targetID  = "019f5700-0000-7000-8000-000000000003"
		planID    = "019f5700-0000-7000-8000-000000000004"
		sessionID = "019f5700-0000-7000-8000-000000000005"
		reportID  = "019f5700-0000-7000-8000-000000000006"
	)
	now := time.Now().UTC().Truncate(time.Microsecond)
	mustExecIntegration(t, ctx, db,
		`insert into users (id,email,display_name,created_at,updated_at) values ($1,'migration-v18@example.test','Migration V18',$2,$2)`,
		userID, now,
	)
	mustExecIntegration(t, ctx, db, `
insert into resumes (
  id,user_id,title,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Migration Resume','en','ready','{}'::jsonb,'resume body','paste','resume body','resume body','{}'::jsonb,$3,$3)`,
		resumeID, userID, now,
	)
	mustExecIntegration(t, ctx, db, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,target_language,source_type,
  raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','en','manual_text','job body','{}'::jsonb,'{}'::jsonb,$4,$4)`,
		targetID, userID, resumeID, now,
	)
	mustExecIntegration(t, ctx, db, `
insert into practice_plans (
  id,user_id,target_job_id,goal,interviewer_persona,difficulty,language,time_budget_minutes,
  resume_id,focus_competency_codes,status,round_id,round_sequence,created_at,updated_at
) values ($1,$2,$3,'baseline','technical_manager','standard','en',30,$4,array['technical_tradeoffs'],'ready','round-1-technical',1,$5,$5)`,
		planID, userID, targetID, resumeID, now,
	)
	mustExecIntegration(t, ctx, db, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,started_at,created_at,updated_at)
values ($1,$2,$3,$4,'completed','en',$5,$5,$5)`,
		sessionID, userID, planID, targetID, now,
	)
	mustExecIntegration(t, ctx, db, `
insert into feedback_reports (
  id,user_id,session_id,target_job_id,status,retry_focus_competency_codes,created_at,updated_at
) values ($1,$2,$3,$4,'queued',array['technical_tradeoffs'],$5,$5)`,
		reportID, userID, sessionID, targetID, now,
	)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v17 to v18: %v", err)
	}
	assertMigrationVersion(t, migrator, 18)
	assertV18ColumnContract(t, ctx, db)

	var defaultsAndRenamesOK bool
	if err := db.QueryRowContext(ctx, `
select summary is null
   and generation_context = '{}'::jsonb
   and generation_context->>'schemaVersion' is distinct from 'report-context.v1'
   and retry_focus_dimension_codes = array['technical_tradeoffs']::text[]
   and (select focus_dimension_codes = array['technical_tradeoffs']::text[] from practice_plans where id=$2)
from feedback_reports where id=$1`, reportID, planID).Scan(&defaultsAndRenamesOK); err != nil {
		t.Fatalf("probe populated v18 defaults and renames: %v", err)
	}
	if !defaultsAndRenamesOK {
		t.Fatal("legacy row must receive only empty invalid context/default repair state while focus values survive rename")
	}

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("migrate populated database v18 down to v17: %v", err)
	}
	assertMigrationVersion(t, migrator, 17)
	assertColumnAbsent(t, ctx, db, "feedback_reports", "summary")
	assertColumnAbsent(t, ctx, db, "feedback_reports", "generation_context")
	assertColumnAbsent(t, ctx, db, "feedback_reports", "llm_attempt_count")
	assertColumnAbsent(t, ctx, db, "feedback_reports", "retry_focus_dimension_codes")
	assertColumnAbsent(t, ctx, db, "practice_plans", "focus_dimension_codes")
	var downRestored bool
	if err := db.QueryRowContext(ctx, `
select retry_focus_competency_codes = array['technical_tradeoffs']::text[]
   and (select focus_competency_codes = array['technical_tradeoffs']::text[] from practice_plans where id=$2)
from feedback_reports where id=$1`, reportID, planID).Scan(&downRestored); err != nil {
		t.Fatalf("probe v17 restored names: %v", err)
	}
	if !downRestored {
		t.Fatal("down migration must restore competency column names without losing values")
	}

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v17 back to v18: %v", err)
	}
	assertMigrationVersion(t, migrator, 18)
	assertV18ColumnContract(t, ctx, db)

	const sentinel = "migration-v18-sensitive-report-content"
	mustExecIntegration(t, ctx, db, `
update feedback_reports
set summary=$2::text,
    generation_context=jsonb_build_object('schemaVersion','report-context.v1','sensitive',$2::text)
where id=$1`, reportID, sentinel)
	assertNoReportContentLeak(t, ctx, db, sentinel)
	mustExecIntegration(t, ctx, db, `delete from users where id=$1`, userID)
	var remainingReports int
	if err := db.QueryRowContext(ctx, `select count(*) from feedback_reports where id=$1`, reportID).Scan(&remainingReports); err != nil {
		t.Fatalf("count privacy-deleted report: %v", err)
	}
	if remainingReports != 0 {
		t.Fatalf("feedback report survived user hard deletion: count=%d", remainingReports)
	}
	assertNoReportContentLeak(t, ctx, db, sentinel)

	var tableCount int
	if err := db.QueryRowContext(ctx, `
select count(*) from information_schema.tables
where table_schema='public' and table_type='BASE TABLE'`).Scan(&tableCount); err != nil {
		t.Fatalf("count final public tables: %v", err)
	}
	if tableCount != 26 {
		t.Fatalf("final public table inventory=%d, want exactly 21 app + 3 auth + 2 metadata = 26", tableCount)
	}

	t.Log("REPORT_STORAGE_V18_POPULATED_MIGRATION_PASS")
	t.Log("REPORT_STORAGE_V18_INVALID_CONTEXT_PASS")
	t.Log("REPORT_STORAGE_V18_RENAME_ROLLBACK_PASS")
	t.Log("REPORT_STORAGE_V18_PRIVACY_PROBE_PASS")
	t.Log("REPORT_STORAGE_V18_PASS")
}

func mustExecIntegration(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec integration SQL: %v", err)
	}
}

func assertMigrationVersion(t *testing.T, migrator *migrate.Migrate, want uint) {
	t.Helper()
	version, dirty, err := migrator.Version()
	if err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if version != want || dirty {
		t.Fatalf("migration version=%d dirty=%t, want %d clean", version, dirty, want)
	}
}

func assertV18ColumnContract(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	checks := []struct {
		table, column, dataType, nullable, defaultContains string
	}{
		{"feedback_reports", "summary", "text", "YES", ""},
		{"feedback_reports", "generation_context", "jsonb", "NO", "'{}'::jsonb"},
		{"feedback_reports", "retry_focus_dimension_codes", "ARRAY", "NO", "'{}'::text[]"},
		{"practice_plans", "focus_dimension_codes", "ARRAY", "NO", "'{}'::text[]"},
	}
	for _, check := range checks {
		var dataType, nullable string
		var columnDefault sql.NullString
		if err := db.QueryRowContext(ctx, `
select data_type,is_nullable,column_default
from information_schema.columns
where table_schema='public' and table_name=$1 and column_name=$2`,
			check.table, check.column).Scan(&dataType, &nullable, &columnDefault); err != nil {
			t.Fatalf("read column %s.%s: %v", check.table, check.column, err)
		}
		if dataType != check.dataType || nullable != check.nullable {
			t.Fatalf("column %s.%s type/nullability=%s/%s, want %s/%s", check.table, check.column, dataType, nullable, check.dataType, check.nullable)
		}
		if check.defaultContains == "" {
			if columnDefault.Valid {
				t.Fatalf("column %s.%s default=%q, want null", check.table, check.column, columnDefault.String)
			}
		} else if !columnDefault.Valid || !strings.Contains(columnDefault.String, check.defaultContains) {
			t.Fatalf("column %s.%s default=%q, want fragment %q", check.table, check.column, columnDefault.String, check.defaultContains)
		}
	}
	for _, old := range []struct{ table, column string }{
		{"feedback_reports", "llm_attempt_count"},
		{"feedback_reports", "retry_focus_competency_codes"},
		{"practice_plans", "focus_competency_codes"},
	} {
		assertColumnAbsent(t, ctx, db, old.table, old.column)
	}
}

func assertColumnAbsent(t *testing.T, ctx context.Context, db *sql.DB, table, column string) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
select count(*) from information_schema.columns
where table_schema='public' and table_name=$1 and column_name=$2`, table, column).Scan(&count); err != nil {
		t.Fatalf("probe absent column %s.%s: %v", table, column, err)
	}
	if count != 0 {
		t.Fatalf("compatibility column %s.%s remains present", table, column)
	}
}

func assertNoReportContentLeak(t *testing.T, ctx context.Context, db *sql.DB, sentinel string) {
	t.Helper()
	var leaked int
	if err := db.QueryRowContext(ctx, `
select
  (select count(*) from audit_events row where to_jsonb(row)::text like '%' || $1 || '%')
  + (select count(*) from async_jobs row where to_jsonb(row)::text like '%' || $1 || '%')
  + (select count(*) from outbox_events row where to_jsonb(row)::text like '%' || $1 || '%')`, sentinel).Scan(&leaked); err != nil {
		t.Fatalf("probe report content leakage: %v", err)
	}
	if leaked != 0 {
		t.Fatalf("report summary/context leaked into audit/job/outbox rows: count=%d", leaked)
	}
}
