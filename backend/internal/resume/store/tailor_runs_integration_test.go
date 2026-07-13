//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	"github.com/monshunter/easyinterview/backend/internal/runner"
)

func TestCompleteTailorRunSuccessFencesStaleTakeover(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for resume tailor lease fencing integration")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	const (
		jobID        = "019f5900-0000-7000-8000-000000000001"
		tailorRunID  = "019f5900-0000-7000-8000-000000000002"
		resumeID     = "019f5900-0000-7000-8000-000000000003"
		staleEvent   = "019f5900-0000-7000-8000-000000000004"
		currentEvent = "019f5900-0000-7000-8000-000000000005"
	)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from outbox_events where id in ($1,$2)`, staleEvent, currentEvent)
		_, _ = db.ExecContext(context.Background(), `delete from async_jobs where id=$1`, jobID)
	}
	cleanup()
	t.Cleanup(cleanup)

	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := db.ExecContext(ctx, `
insert into async_jobs (
  id,job_type,resource_type,resource_id,status,attempts,max_attempts,
  payload,available_at,locked_at,created_at,updated_at
) values ($1,'resume_tailor','resume_tailor_run',$2,'running',1,5,'{}'::jsonb,$3,$3,$3,$3)`, jobID, tailorRunID, now.Add(-10*time.Minute)); err != nil {
		t.Fatalf("seed resume tailor job: %v", err)
	}

	kernelStore := runner.NewSQLStore(db)
	if reclaimed, err := kernelStore.ReclaimExpiredLeases(ctx, []string{"resume_tailor"}, now.Add(-5*time.Minute), now); err != nil || reclaimed != 1 {
		t.Fatalf("reap attempt1: reclaimed=%d err=%v", reclaimed, err)
	}
	claimed, ok, err := kernelStore.LeaseAsyncJob(ctx, []string{"resume_tailor"}, now.Add(time.Second))
	if err != nil || !ok || claimed.JobID != jobID || claimed.Attempts != 2 {
		t.Fatalf("claim attempt2: claimed=%+v ok=%t err=%v", claimed, ok, err)
	}

	repo := resumestore.NewRepository(db)
	stale := resumestore.CompleteTailorRunSuccessInput{
		JobID: jobID, ClaimedAttempts: 1, TailorRunID: tailorRunID, ResumeID: resumeID,
		MatchSummary:  []byte(`{"strengths":[],"gaps":[]}`),
		OutboxEventID: staleEvent, OutboxEventPayload: []byte(`{}`), Now: now.Add(2 * time.Second),
	}
	if err := repo.CompleteTailorRunSuccess(ctx, stale); !errors.Is(err, runner.ErrStaleLease) {
		t.Fatalf("stale tailor success err=%v want ErrStaleLease", err)
	}
	assertTailorFenceState(t, ctx, db, jobID, staleEvent, "running", 2, true, 0)

	current := stale
	current.ClaimedAttempts = claimed.Attempts
	current.OutboxEventID = currentEvent
	current.Now = now.Add(10 * time.Minute)
	if err := repo.CompleteTailorRunSuccess(ctx, current); err != nil {
		t.Fatalf("current tailor success: %v", err)
	}
	if reclaimed, err := kernelStore.ReclaimExpiredLeases(ctx, []string{"resume_tailor"}, current.Now.Add(-5*time.Minute), current.Now); err != nil || reclaimed != 0 {
		t.Fatalf("reaper interposed after tailor success: reclaimed=%d err=%v", reclaimed, err)
	}
	if err := kernelStore.FinalizeAsyncJob(ctx, jobID, claimed.Attempts, runner.JobOutcome{Succeeded: true}, current.Now, current.Now); err != nil {
		t.Fatalf("finalize current tailor lease: %v", err)
	}
	assertTailorFenceState(t, ctx, db, jobID, currentEvent, "succeeded", 2, false, 1)
}

func assertTailorFenceState(t *testing.T, ctx context.Context, db *sql.DB, jobID, eventID, wantStatus string, wantAttempts int32, wantResultEmpty bool, wantEvents int) {
	t.Helper()
	var status string
	var attempts int32
	var resultEmpty bool
	if err := db.QueryRowContext(ctx, `select status,attempts,result = '{}'::jsonb from async_jobs where id=$1`, jobID).Scan(&status, &attempts, &resultEmpty); err != nil {
		t.Fatalf("read tailor fence state: %v", err)
	}
	if status != wantStatus || attempts != wantAttempts || resultEmpty != wantResultEmpty {
		t.Fatalf("tailor fence state status=%s attempts=%d resultEmpty=%t want status=%s attempts=%d resultEmpty=%t", status, attempts, resultEmpty, wantStatus, wantAttempts, wantResultEmpty)
	}
	var events int
	if err := db.QueryRowContext(ctx, `select count(*) from outbox_events where id=$1`, eventID).Scan(&events); err != nil {
		t.Fatalf("count tailor outbox events: %v", err)
	}
	if events != wantEvents {
		t.Fatalf("tailor outbox events=%d want=%d", events, wantEvents)
	}
}
