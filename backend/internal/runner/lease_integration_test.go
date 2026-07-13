//go:build integration

package runner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func openRunnerTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping runner store integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("db not reachable: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func insertAsyncJob(t *testing.T, db *sql.DB, id, jobType, status string, attempts int32, availableAt, lockedAt time.Time, locked bool) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `delete from async_jobs where id = $1`, id)
	var lockedArg any
	if locked {
		lockedArg = lockedAt
	}
	_, err := db.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, status, attempts, max_attempts,
  payload, available_at, locked_at, created_at, updated_at
) values ($1,$2,'target_job',$1,$3,$4,5,'{}'::jsonb,$5,$6,$5,$5)`,
		id, jobType, status, attempts, availableAt, lockedArg)
	if err != nil {
		t.Fatalf("insert async_jobs: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `delete from async_jobs where id = $1`, id) })
}

func TestLeaseAsyncJob_ClaimsQueuedRow(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	id := "0197d130-0000-7000-8000-000000000001"
	insertAsyncJob(t, db, id, "target_import", "queued", 0, now.Add(-time.Minute), time.Time{}, false)

	job, ok, err := store.LeaseAsyncJob(context.Background(), []string{"target_import"}, now)
	if err != nil {
		t.Fatalf("LeaseAsyncJob: %v", err)
	}
	if !ok {
		t.Fatalf("expected to claim a row")
	}
	if job.JobID != id || job.JobType != "target_import" {
		t.Fatalf("claimed %+v, want id=%s target_import", job, id)
	}
	if job.Attempts != 1 {
		t.Fatalf("attempts = %d, want 1 after claim", job.Attempts)
	}
	var status string
	var lockedAt sql.NullTime
	if err := db.QueryRow(`select status, locked_at from async_jobs where id=$1`, id).Scan(&status, &lockedAt); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != "running" || !lockedAt.Valid {
		t.Fatalf("row status=%s locked=%v, want running + locked_at set", status, lockedAt.Valid)
	}
}

func TestLeaseAsyncJob_ColumnNames(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	id := "0197d130-0000-7000-8000-000000000002"
	insertAsyncJob(t, db, id, "target_import", "queued", 2, now.Add(-time.Minute), time.Time{}, false)

	job, ok, err := store.LeaseAsyncJob(context.Background(), []string{"target_import"}, now)
	if err != nil || !ok {
		t.Fatalf("LeaseAsyncJob: ok=%v err=%v", ok, err)
	}
	// The B4 baseline column projection must populate every kernel field.
	if job.ResourceType != "target_job" || job.ResourceID != id {
		t.Fatalf("resource columns = %s/%s, want target_job/%s", job.ResourceType, job.ResourceID, id)
	}
	if job.MaxAttempts != 5 {
		t.Fatalf("max_attempts = %d, want 5", job.MaxAttempts)
	}
	if job.AvailableAt.IsZero() {
		t.Fatalf("available_at not populated")
	}
}

func TestOrdinaryBusinessJobStillUsesSchemaDefaultFiveAttempts(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	id := "0197d130-0000-7000-8000-000000000099"
	_, _ = db.ExecContext(context.Background(), `delete from async_jobs where id=$1`, id)
	if _, err := db.ExecContext(context.Background(), `
insert into async_jobs (
  id,job_type,resource_type,resource_id,status,attempts,payload,available_at,created_at,updated_at
) values ($1,'target_import','target_job',$1,'queued',0,'{}'::jsonb,$2,$2,$2)`, id, now); err != nil {
		t.Fatalf("insert ordinary business job: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `delete from async_jobs where id=$1`, id) })
	job, ok, err := store.LeaseAsyncJob(context.Background(), []string{"target_import"}, now)
	if err != nil || !ok {
		t.Fatalf("lease ordinary business job ok=%t err=%v", ok, err)
	}
	if job.MaxAttempts != MaxAttempts {
		t.Fatalf("ordinary business max_attempts=%d want=%d", job.MaxAttempts, MaxAttempts)
	}
}

func TestFinalizeAsyncJob_Succeeded(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	id := "0197d130-0000-7000-8000-000000000003"
	insertAsyncJob(t, db, id, "target_import", "running", 1, now, now, true)

	if err := store.FinalizeAsyncJob(context.Background(), id, 1, JobOutcome{Succeeded: true}, now, now); err != nil {
		t.Fatalf("FinalizeAsyncJob: %v", err)
	}
	var status string
	var lockedAt sql.NullTime
	if err := db.QueryRow(`select status, locked_at from async_jobs where id=$1`, id).Scan(&status, &lockedAt); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != "succeeded" || lockedAt.Valid {
		t.Fatalf("status=%s locked=%v, want succeeded + locked_at cleared", status, lockedAt.Valid)
	}
}

func TestFinalizeAsyncJob_RetryableRequeues(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	id := "0197d130-0000-7000-8000-000000000004"
	insertAsyncJob(t, db, id, "target_import", "running", 1, now, now, true)

	available := now.Add(30 * time.Second)
	if err := store.FinalizeAsyncJob(context.Background(), id, 1, JobOutcome{Retryable: true, ErrorCode: "TRANSIENT"}, available, now); err != nil {
		t.Fatalf("FinalizeAsyncJob: %v", err)
	}
	var status string
	var availableAt time.Time
	if err := db.QueryRow(`select status, available_at from async_jobs where id=$1`, id).Scan(&status, &availableAt); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != "queued" {
		t.Fatalf("status=%s, want queued (retryable below max)", status)
	}
	if availableAt.Sub(now) < 29*time.Second {
		t.Fatalf("available_at not advanced by backoff: %s", availableAt.Sub(now))
	}
}

func TestFinalizeAsyncJob_RetryableDeadAtMax(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	id := "0197d130-0000-7000-8000-000000000005"
	insertAsyncJob(t, db, id, "target_import", "running", 5, now, now, true)

	if err := store.FinalizeAsyncJob(context.Background(), id, 5, JobOutcome{Retryable: true, ErrorCode: "TRANSIENT"}, now.Add(time.Hour), now); err != nil {
		t.Fatalf("FinalizeAsyncJob: %v", err)
	}
	var status string
	var availableAt time.Time
	if err := db.QueryRow(`select status,available_at from async_jobs where id=$1`, id).Scan(&status, &availableAt); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != "dead" {
		t.Fatalf("status=%s, want dead at max attempts", status)
	}
	if !availableAt.Equal(now) {
		t.Fatalf("dead job scheduled another retry at %s want unchanged %s", availableAt, now)
	}
}

func TestFinalizeAsyncJob_RejectsStaleLeaseGenerationAfterTakeover(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	outcomes := []struct {
		name    string
		outcome JobOutcome
	}{
		{name: "success", outcome: JobOutcome{Succeeded: true}},
		{name: "retry", outcome: JobOutcome{Retryable: true, ErrorCode: "TRANSIENT"}},
		{name: "failure", outcome: JobOutcome{ErrorCode: "PERMANENT"}},
	}
	for index, tc := range outcomes {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprintf("0197d130-0000-7000-8000-%012d", 100+index)
			insertAsyncJob(t, db, id, "target_import", "running", 1, now.Add(-time.Hour), now.Add(-time.Hour), true)

			if reclaimed, err := store.ReclaimExpiredLeases(context.Background(), []string{"target_import"}, now.Add(-time.Minute), now); err != nil || reclaimed != 1 {
				t.Fatalf("reclaim stale attempt1: reclaimed=%d err=%v", reclaimed, err)
			}
			claimed, ok, err := store.LeaseAsyncJob(context.Background(), []string{"target_import"}, now.Add(time.Second))
			if err != nil || !ok || claimed.Attempts != 2 {
				t.Fatalf("claim attempt2: claimed=%+v ok=%t err=%v", claimed, ok, err)
			}

			err = store.FinalizeAsyncJob(context.Background(), id, 1, tc.outcome, now.Add(10*time.Second), now.Add(2*time.Second))
			if !errors.Is(err, ErrStaleLease) {
				t.Fatalf("stale attempt1 finalize err=%v want ErrStaleLease", err)
			}
			var status string
			var attempts int32
			if err := db.QueryRow(`select status,attempts from async_jobs where id=$1`, id).Scan(&status, &attempts); err != nil {
				t.Fatalf("read takeover row: %v", err)
			}
			if status != "running" || attempts != 2 {
				t.Fatalf("stale attempt1 overwrote takeover: status=%s attempts=%d", status, attempts)
			}
		})
	}
}

func TestReclaimExpiredLeases_Integration(t *testing.T) {
	db := openRunnerTestDB(t)
	store := NewSQLStore(db)
	now := time.Now().UTC()
	expired := "0197d130-0000-7000-8000-000000000006"
	fresh := "0197d130-0000-7000-8000-000000000007"
	insertAsyncJob(t, db, expired, "target_import", "running", 2, now.Add(-time.Hour), now.Add(-time.Hour), true)
	insertAsyncJob(t, db, fresh, "target_import", "running", 1, now, now, true)

	reclaimed, err := store.ReclaimExpiredLeases(context.Background(), []string{"target_import"}, now.Add(-5*time.Minute), now)
	if err != nil {
		t.Fatalf("ReclaimExpiredLeases: %v", err)
	}
	if reclaimed < 1 {
		t.Fatalf("reclaimed = %d, want >= 1", reclaimed)
	}
	var status string
	var attempts int32
	if err := db.QueryRow(`select status, attempts from async_jobs where id=$1`, expired).Scan(&status, &attempts); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != "queued" || attempts != 2 {
		t.Fatalf("expired row status=%s attempts=%d, want queued + attempts unchanged (2)", status, attempts)
	}
	if err := db.QueryRow(`select status from async_jobs where id=$1`, fresh).Scan(&status); err != nil {
		t.Fatalf("read back fresh: %v", err)
	}
	if status != "running" {
		t.Fatalf("fresh row status=%s, want still running", status)
	}
}
