//go:build integration

package review_test

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestLeaseSkipLocked(t *testing.T) {
	db := openReviewStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Date(2026, 5, 15, 16, 0, 0, 0, time.UTC)
	jobID := "0197d120-0000-7000-8000-000000000030"
	reportID := "0197d120-0000-7000-8000-000000000031"
	cleanupReviewAsyncJob(t, db, jobID, reportID)
	mustExecReview(t, ctx, db, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, status, payload, available_at, created_at, updated_at
) values ($1, 'report_generate', 'feedback_report', $2, 'queued', '{}', $3, $3, $3)`,
		jobID, reportID, now.Add(-time.Minute))

	repo := reviewstore.NewRepository(db)
	start := make(chan struct{})
	results := make(chan bool, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, ok, err := repo.LeaseAsyncJob(ctx, "report_generate", now)
			if err != nil {
				errs <- err
				return
			}
			results <- ok
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		t.Fatalf("LeaseAsyncJob: %v", err)
	}
	claimed := 0
	for ok := range results {
		if ok {
			claimed++
		}
	}
	if claimed != 1 {
		t.Fatalf("claimed workers = %d, want 1", claimed)
	}

	var status string
	var attempts int
	var lockedAt sql.NullTime
	if err := db.QueryRowContext(ctx, `select status, attempts, locked_at from async_jobs where id = $1`, jobID).Scan(&status, &attempts, &lockedAt); err != nil {
		t.Fatalf("select leased job: %v", err)
	}
	if status != "running" || attempts != 1 || !lockedAt.Valid {
		t.Fatalf("leased row status=%s attempts=%d locked_at=%v", status, attempts, lockedAt)
	}

	_, ok, err := repo.LeaseAsyncJob(ctx, "report_generate", now.Add(time.Second))
	if err != nil {
		t.Fatalf("second LeaseAsyncJob: %v", err)
	}
	if ok {
		t.Fatal("second LeaseAsyncJob claimed locked/running row")
	}
	if err := repo.UpdateAsyncJobSucceeded(ctx, jobID, now.Add(time.Minute)); err != nil {
		t.Fatalf("UpdateAsyncJobSucceeded: %v", err)
	}
	if err := db.QueryRowContext(ctx, `select locked_at from async_jobs where id = $1`, jobID).Scan(&lockedAt); err != nil {
		t.Fatalf("select completed job: %v", err)
	}
	if lockedAt.Valid {
		t.Fatalf("locked_at after success = %v, want null", lockedAt)
	}
}

func openReviewStoreTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping review store integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	return db
}

func cleanupReviewAsyncJob(t *testing.T, db *sql.DB, jobID string, reportID string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.Exec(`delete from async_jobs where id = $1 or resource_id = $2`, jobID, reportID)
	})
}

func mustExecReview(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
