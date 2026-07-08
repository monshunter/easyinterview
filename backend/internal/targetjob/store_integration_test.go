//go:build integration

package targetjob_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestSQLStoreIntegration_GetTargetJobByUser_AllowsFailedJobWithoutRequirements(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping targetjob store integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("postgres ping failed (%v); skipping targetjob store integration test", err)
	}

	userID := "019f40d0-0000-7000-8000-000000000001"
	targetID := "019f40d0-0000-7000-8000-000000000002"
	sourceID := "019f40d0-0000-7000-8000-000000000003"
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from target_job_sources where target_job_id = $1`, targetID)
		_, _ = db.ExecContext(context.Background(), `delete from target_job_requirements where target_job_id = $1`, targetID)
		_, _ = db.ExecContext(context.Background(), `delete from target_jobs where id = $1`, targetID)
		_, _ = db.ExecContext(context.Background(), `delete from users where id = $1`, userID)
	}
	cleanup()
	t.Cleanup(cleanup)

	now := time.Now().UTC()
	if _, err := db.ExecContext(ctx, `
insert into users (id, email, display_name, created_at, updated_at)
values ($1, $2, $3, $4, $4)`,
		userID, "targetjob-failed-detail@example.test", "TargetJob Detail Regression", now,
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into target_jobs (
  id, user_id, status, analysis_status, target_language, source_type,
  raw_jd_text, summary, fit_summary, created_at, updated_at
) values ($1, $2, 'draft', 'failed', 'zh-CN', 'manual_text', $3, '{}'::jsonb, '{}'::jsonb, $4, $4)`,
		targetID, userID, "integration jd text", now,
	); err != nil {
		t.Fatalf("insert target job: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into target_job_sources (id, target_job_id, source_type, snapshot_text, created_at)
values ($1, $2, 'manual_text', $3, $4)`,
		sourceID, targetID, "integration jd text", now,
	); err != nil {
		t.Fatalf("insert target job source: %v", err)
	}

	got, reqs, sources, err := targetjob.NewSQLStore(db).GetTargetJobByUser(ctx, userID, targetID)
	if err != nil {
		t.Fatalf("GetTargetJobByUser failed target: %v", err)
	}
	if got.AnalysisStatus != sharedtypes.TargetJobParseStatusFailed {
		t.Fatalf("analysisStatus = %q, want failed", got.AnalysisStatus)
	}
	if len(reqs) != 0 {
		t.Fatalf("expected no requirements for failed parse, got %+v", reqs)
	}
	if len(sources) != 1 || sources[0].SourceType != targetjob.SourceTypeManualText {
		t.Fatalf("sources unexpected: %+v", sources)
	}
}
