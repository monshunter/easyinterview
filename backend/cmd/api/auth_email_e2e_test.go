//go:build integration

package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

// TestAuthEmailEndToEnd proves the spec D-10 / D-14 contract: starting an email
// challenge enqueues an email_dispatch async_jobs row, and the runner kernel
// EmailDispatchHandler delivers the magic link within a single scan cycle.
func TestAuthEmailEndToEnd(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping auth email end-to-end")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("db not reachable: %v", err)
	}

	challengeID := "0197d150-0000-7000-8000-000000000001"
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from async_jobs where resource_id = $1`, challengeID)
		_, _ = db.ExecContext(context.Background(), `delete from auth_challenges where id = $1`, challengeID)
	}
	cleanup()
	t.Cleanup(cleanup)

	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	enqueuer := auth.NewEmailDispatchEnqueuer(db, func() string { return idx.NewID() }, func() time.Time { return time.Now().UTC() })
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		Dispatcher:          enqueuer,
		DeliverySecrets:     sink,
		ChallengePepper:     "e2e-pepper",
		SessionCookieSecret: "e2e-session-secret",
		NewID:               func() string { return challengeID },
		Now:                 func() time.Time { return time.Now().UTC() },
	})

	res, err := service.StartEmailChallenge(ctx, auth.StartEmailChallengeInput{Email: "auth-e2e@example.test"})
	if err != nil {
		t.Fatalf("StartEmailChallenge: %v", err)
	}
	if !res.Accepted {
		t.Fatalf("challenge not accepted: %+v", res)
	}

	// The async_jobs row exists and is email_dispatch typed.
	var jobType, status string
	if err := db.QueryRowContext(ctx, `select job_type, status from async_jobs where resource_id = $1`, challengeID).Scan(&jobType, &status); err != nil {
		t.Fatalf("read email_dispatch job: %v", err)
	}
	if jobType != string(jobs.JobTypeEmailDispatch) || status != "queued" {
		t.Fatalf("job_type=%s status=%s, want email_dispatch/queued", jobType, status)
	}

	// One kernel scan cycle delivers the magic link.
	kernel := runner.New(runner.Options{Store: runner.NewSQLStore(db), Config: testRunnerConfig()})
	kernel.Register(string(jobs.JobTypeEmailDispatch), auth.NewEmailDispatchHandler(sink))
	processed, err := kernel.RunOnce(ctx)
	if err != nil {
		t.Fatalf("kernel RunOnce: %v", err)
	}
	if !processed {
		t.Fatalf("kernel did not process the email_dispatch job")
	}
	link, ok := sink.MagicLinkForChallenge(challengeID)
	if !ok || link == "" {
		t.Fatalf("magic link not delivered within one scan cycle (ok=%v)", ok)
	}
}
