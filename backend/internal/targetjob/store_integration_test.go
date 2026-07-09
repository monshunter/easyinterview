//go:build integration

package targetjob_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestSQLStoreIntegration_CompleteParseFailureDeletesTargetAndSources(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for TargetJob parse-failure persistence integration gate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("postgres ping failed for TargetJob parse-failure persistence integration gate: %v", err)
	}

	userID := "019f40d0-0000-7000-8000-000000000001"
	targetID := "019f40d0-0000-7000-8000-000000000002"
	sourceID := "019f40d0-0000-7000-8000-000000000003"
	eventID := "019f40d0-0000-7000-8000-000000000004"
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from outbox_events where id = $1 or aggregate_id = $2`, eventID, targetID)
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

	store := targetjob.NewSQLStore(db)
	if err := store.CompleteParseFailure(ctx, targetjob.CompleteParseFailureInput{
		TargetJobID:        targetID,
		FailedEventID:      eventID,
		FailedEventPayload: []byte(`{"targetJobId":"019f40d0-0000-7000-8000-000000000002","errorCode":"AI_OUTPUT_INVALID","retryable":false}`),
		Now:                now.Add(time.Second),
	}); err != nil {
		t.Fatalf("CompleteParseFailure: %v", err)
	}

	_, _, _, err = store.GetTargetJobByUser(ctx, userID, targetID)
	if !errors.Is(err, targetjob.ErrTargetJobNotFound) {
		t.Fatalf("failed target must not be readable after failure commit, got %v", err)
	}

	var targetCount int
	if err := db.QueryRowContext(ctx, `select count(*) from target_jobs where id = $1`, targetID).Scan(&targetCount); err != nil {
		t.Fatalf("count target_jobs: %v", err)
	}
	if targetCount != 0 {
		t.Fatalf("failed target row persisted, count=%d", targetCount)
	}
	var sourceCount int
	if err := db.QueryRowContext(ctx, `select count(*) from target_job_sources where target_job_id = $1`, targetID).Scan(&sourceCount); err != nil {
		t.Fatalf("count target_job_sources: %v", err)
	}
	if sourceCount != 0 {
		t.Fatalf("failed target sources were not cascade-deleted, count=%d", sourceCount)
	}
	var eventName string
	if err := db.QueryRowContext(ctx, `select event_name from outbox_events where id = $1`, eventID).Scan(&eventName); err != nil {
		t.Fatalf("select outbox event: %v", err)
	}
	if eventName != string(events.EventNameTargetAnalysisFailed) {
		t.Fatalf("outbox event_name = %q, want %q", eventName, events.EventNameTargetAnalysisFailed)
	}
}
