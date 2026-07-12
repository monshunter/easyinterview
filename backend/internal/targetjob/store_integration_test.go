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
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestSQLStoreIntegration_PracticeProgressProjectionPersistsAcrossGetAndList(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for TargetJob practice-progress persistence integration gate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("postgres ping failed for TargetJob practice-progress persistence integration gate: %v", err)
	}

	const (
		userID        = "019f50d0-0000-7000-8000-000000000001"
		resumeID      = "019f50d0-0000-7000-8000-000000000002"
		targetID      = "019f50d0-0000-7000-8000-000000000003"
		wrongResumeID = "019f50d0-0000-7000-8000-000000000004"
		round1PlanID  = "019f50d0-0000-7000-8000-000000000010"
		round1RetryID = "019f50d0-0000-7000-8000-000000000011"
		round2PlanID  = "019f50d0-0000-7000-8000-000000000012"
		round4PlanID  = "019f50d0-0000-7000-8000-000000000013"
		legacyPlanID  = "019f50d0-0000-7000-8000-000000000014"
		wrongPlanID   = "019f50d0-0000-7000-8000-000000000015"
		round1Session = "019f50d0-0000-7000-8000-000000000020"
		round1Dup     = "019f50d0-0000-7000-8000-000000000021"
		round2Session = "019f50d0-0000-7000-8000-000000000022"
		round4Session = "019f50d0-0000-7000-8000-000000000023"
		wrongSession  = "019f50d0-0000-7000-8000-000000000024"
		reportID      = "019f50d0-0000-7000-8000-000000000040"
	)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from users where id = $1`, userID)
	}
	cleanup()
	t.Cleanup(cleanup)

	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := db.ExecContext(ctx, `
insert into users (id, email, display_name, created_at, updated_at)
values ($1, $2, $3, $4, $4)`,
		userID, "targetjob-practice-progress@example.test", "TargetJob Practice Progress", now,
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into resumes (id, user_id, title, language, parse_status, parsed_summary, created_at, updated_at)
values
  ($1, $3, 'Practice progress resume', 'en', 'ready', '{}'::jsonb, $4, $4),
  ($2, $3, 'Wrong binding resume', 'en', 'ready', '{}'::jsonb, $4, $4)`,
		resumeID, wrongResumeID, userID, now,
	); err != nil {
		t.Fatalf("insert resume: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into target_jobs (
  id, user_id, status, analysis_status, title, company_name, target_language,
  source_type, summary, fit_summary, resume_id, created_at, updated_at
) values ($1, $2, 'draft', 'ready', 'Backend Engineer', 'Acme', 'en',
          'manual_text', $3::jsonb, '{}'::jsonb, $4, $5, $5)`,
		targetID, userID, string(nonContiguousRoundSummaryJSON()), resumeID, now,
	); err != nil {
		t.Fatalf("insert target job: %v", err)
	}

	insertPlan := func(id, boundResumeID string, roundID any, roundSequence any, createdAt time.Time) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `
insert into practice_plans (
  id, user_id, target_job_id, goal, interviewer_persona, difficulty, language,
  time_budget_minutes, resume_id, status, round_id, round_sequence, created_at, updated_at
) values ($1, $2, $3, 'baseline', 'generalist', 'standard', 'en', 30, $4, 'ready', $5, $6, $7, $7)`,
			id, userID, targetID, boundResumeID, roundID, roundSequence, createdAt,
		); err != nil {
			t.Fatalf("insert practice plan %s: %v", id, err)
		}
	}
	insertPlan(round1PlanID, resumeID, "round-1-hr", 1, now)
	insertPlan(round2PlanID, resumeID, "round-2-technical", 2, now.Add(time.Minute))
	insertPlan(round4PlanID, resumeID, "round-4-manager", 4, now.Add(2*time.Minute))
	insertPlan(round1RetryID, resumeID, "round-1-hr", 1, now.Add(10*time.Minute))
	insertPlan(legacyPlanID, resumeID, nil, nil, now.Add(20*time.Minute))
	insertPlan(wrongPlanID, wrongResumeID, "round-1-hr", 1, now.Add(30*time.Minute))

	insertSession := func(id, planID, status string) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `
insert into practice_sessions (
  id, user_id, plan_id, target_job_id, status, language, started_at, created_at, updated_at
) values ($1, $2, $3, $4, $5, 'en', $6, $6, $6)`,
			id, userID, planID, targetID, status, now,
		); err != nil {
			t.Fatalf("insert practice session %s: %v", id, err)
		}
	}
	insertSession(round1Session, round1PlanID, "completing")
	insertSession(round1Dup, round1PlanID, "failed")
	insertSession(round2Session, round2PlanID, "waiting_user_input")
	insertSession(round4Session, round4PlanID, "waiting_user_input")
	insertSession(wrongSession, wrongPlanID, "completed")
	if _, err := db.ExecContext(ctx, `
insert into practice_session_events (id, session_id, seq_no, event_type, payload, created_at)
values ('019f50d0-0000-7000-8000-000000000034', $1, 1, 'session_completed', '{}'::jsonb, $2)`,
		wrongSession, now.Add(30*time.Minute),
	); err != nil {
		t.Fatalf("insert wrong-resume completion fact: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into feedback_reports (id, user_id, session_id, target_job_id, status, created_at, updated_at)
values ($1, $2, $3, $4, 'queued', $5, $5)`,
		reportID, userID, round1Session, targetID, now,
	); err != nil {
		t.Fatalf("insert feedback report: %v", err)
	}

	store := targetjob.NewSQLStore(db)
	service := targetjob.NewService(targetjob.ServiceOptions{Store: store})
	assertProjection := func(stage, wantStatus string, wantCompleted []string, wantCurrent, wantPlan string) {
		t.Helper()
		detail, err := service.GetTargetJob(ctx, userID, targetID)
		if err != nil {
			t.Fatalf("%s GetTargetJob: %v", stage, err)
		}
		list, err := service.ListTargetJobs(ctx, targetjob.ListRequest{UserID: userID, PageSize: 20})
		if err != nil {
			t.Fatalf("%s ListTargetJobs: %v", stage, err)
		}
		if len(list.Items) != 1 {
			t.Fatalf("%s list items = %d, want 1", stage, len(list.Items))
		}
		for _, observed := range []struct {
			name string
			job  api.TargetJob
		}{{name: "get", job: detail}, {name: "list", job: list.Items[0]}} {
			progress := observed.job.PracticeProgress
			if progress == nil || progress.Status != wantStatus {
				t.Fatalf("%s %s progress = %+v, want status %s", stage, observed.name, progress, wantStatus)
			}
			if len(progress.CompletedRounds) != len(wantCompleted) {
				t.Fatalf("%s %s completed = %+v", stage, observed.name, progress.CompletedRounds)
			}
			for i, roundID := range wantCompleted {
				if progress.CompletedRounds[i].RoundId != roundID {
					t.Fatalf("%s %s completed[%d] = %+v, want %s", stage, observed.name, i, progress.CompletedRounds[i], roundID)
				}
			}
			if wantCurrent == "" {
				if progress.CurrentRound != nil {
					t.Fatalf("%s %s current = %+v, want nil", stage, observed.name, progress.CurrentRound)
				}
			} else if progress.CurrentRound == nil || progress.CurrentRound.RoundId != wantCurrent {
				t.Fatalf("%s %s current = %+v, want %s", stage, observed.name, progress.CurrentRound, wantCurrent)
			}
			if wantPlan == "" {
				if observed.job.CurrentPracticePlanId != nil {
					t.Fatalf("%s %s plan = %v, want nil", stage, observed.name, observed.job.CurrentPracticePlanId)
				}
			} else if observed.job.CurrentPracticePlanId == nil || *observed.job.CurrentPracticePlanId != wantPlan {
				t.Fatalf("%s %s plan = %v, want %s", stage, observed.name, observed.job.CurrentPracticePlanId, wantPlan)
			}
		}
	}

	assertProjection("first", "not_started", nil, "round-1-hr", round1RetryID)
	t.Log("wrong-resume-completion-ignored=PASS")
	if _, err := db.ExecContext(ctx, `
insert into practice_session_events (id, session_id, seq_no, event_type, payload, created_at)
values
  ('019f50d0-0000-7000-8000-000000000030', $1, 1, 'session_completed', '{}'::jsonb, $3),
  ('019f50d0-0000-7000-8000-000000000031', $2, 1, 'session_completed', '{}'::jsonb, $3)`,
		round1Session, round1Dup, now.Add(time.Hour),
	); err != nil {
		t.Fatalf("insert duplicate round-1 completion facts: %v", err)
	}
	assertProjection("next", "in_progress", []string{"round-1-hr"}, "round-2-technical", round2PlanID)
	t.Log("persisted-first-to-next=PASS")

	for _, state := range []struct {
		targetStatus string
		reportStatus string
	}{{targetStatus: "draft", reportStatus: "queued"}, {targetStatus: "interviewing", reportStatus: "ready"}, {targetStatus: "offer", reportStatus: "failed"}} {
		if _, err := db.ExecContext(ctx, `update target_jobs set status = $1 where id = $2`, state.targetStatus, targetID); err != nil {
			t.Fatalf("update target lifecycle status: %v", err)
		}
		if _, err := db.ExecContext(ctx, `update feedback_reports set status = $1 where id = $2`, state.reportStatus, reportID); err != nil {
			t.Fatalf("update report status: %v", err)
		}
		assertProjection("status-report-independent-"+state.targetStatus+"-"+state.reportStatus, "in_progress", []string{"round-1-hr"}, "round-2-technical", round2PlanID)
	}
	t.Log("target-report-status-independent=PASS")

	if _, err := db.ExecContext(ctx, `
insert into practice_session_events (id, session_id, seq_no, event_type, payload, created_at)
values ('019f50d0-0000-7000-8000-000000000033', $1, 1, 'session_completed', '{}'::jsonb, $2)`,
		round4Session, now.Add(2*time.Hour),
	); err != nil {
		t.Fatalf("insert out-of-order round-4 completion fact: %v", err)
	}
	assertProjection("round-4-gap-hidden", "in_progress", []string{"round-1-hr"}, "round-2-technical", round2PlanID)
	t.Log("out-of-order-gap-hidden=PASS")

	if _, err := db.ExecContext(ctx, `
insert into practice_session_events (id, session_id, seq_no, event_type, payload, created_at)
values ('019f50d0-0000-7000-8000-000000000032', $1, 1, 'session_completed', '{}'::jsonb, $2)`,
		round2Session, now.Add(3*time.Hour),
	); err != nil {
		t.Fatalf("insert round-2 completion fact: %v", err)
	}
	assertProjection("final", "completed", []string{"round-1-hr", "round-2-technical", "round-4-manager"}, "", "")
	t.Log("non-contiguous-round-1-2-4=PASS")
	t.Log("get-list-first-next-final-parity=PASS")
}

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
