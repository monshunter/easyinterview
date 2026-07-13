//go:build integration

package practice

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for practice round persistence integration gate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("postgres ping: %v", err)
	}

	const (
		userID        = "019f5585-0000-7000-8000-000000000001"
		resumeID      = "019f5585-0000-7000-8000-000000000002"
		targetID      = "019f5585-0000-7000-8000-000000000003"
		otherResumeID = "019f5585-0000-7000-8000-000000000004"
	)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from audit_events where user_id=$1 or actor_id=$1`, userID)
		_, _ = db.ExecContext(context.Background(), `delete from users where id=$1`, userID)
	}
	cleanup()
	t.Cleanup(cleanup)

	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := db.ExecContext(ctx, `insert into users (id,email,display_name,created_at,updated_at) values ($1,$2,$3,$4,$4)`, userID, "practice-round-integration@example.test", "Practice Round Integration", now); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into resumes (
  id,user_id,title,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,$3,'zh-CN','ready','{}'::jsonb,$4,'paste',$4,$4,'{}'::jsonb,$5,$5)`, resumeID, userID, "Integration resume", "complete resume", now); err != nil {
		t.Fatalf("insert resume: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into resumes (
  id,user_id,title,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,$3,'zh-CN','ready','{}'::jsonb,$4,'paste',$4,$4,'{}'::jsonb,$5,$5)`, otherResumeID, userID, "Other integration resume", "other complete resume", now); err != nil {
		t.Fatalf("insert other resume: %v", err)
	}
	summary := `{"interviewRounds":[` +
		`{"sequence":1,"type":"hr","name":"HR","durationMinutes":60,"focus":"motivation"},` +
		`{"sequence":2,"type":"technical","name":"Technical","durationMinutes":60,"focus":"system design"},` +
		`{"sequence":4,"type":"manager","name":"Manager","durationMinutes":45,"focus":"ownership"}` +
		`],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"zh-CN","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`
	if _, err := db.ExecContext(ctx, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,target_language,source_type,
  summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','zh-CN','manual_text',$4::jsonb,'{}'::jsonb,$5,$5)`, targetID, userID, resumeID, summary, now); err != nil {
		t.Fatalf("insert target job: %v", err)
	}

	repo := NewSQLRepository(db)
	createWithResume := func(planID, auditID, sourceReportID string, goal sharedtypes.PracticeGoal, roundID string, budget int32, requestedResumeID string) (domain.PlanRecord, error) {
		return repo.CreatePlan(ctx, domain.CreatePlanStoreInput{
			PlanID: planID, AuditEventID: auditID, UserID: userID, TargetJobID: targetID,
			ResumeID: requestedResumeID, SourceReportID: sourceReportID, RoundID: roundID, Goal: goal,
			InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
			Difficulty:         "standard", Language: "zh-CN", TimeBudgetMinutes: budget, Now: now,
		})
	}
	create := func(planID, auditID, sourceReportID string, goal sharedtypes.PracticeGoal, roundID string, budget int32) (domain.PlanRecord, error) {
		return createWithResume(planID, auditID, sourceReportID, goal, roundID, budget, resumeID)
	}
	complete := func(planID, sessionID, eventID, reportID string) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,completed_at,created_at,updated_at)
values ($1,$2,$3,$4,'completed','zh-CN',$5,$5,$5)`, sessionID, userID, planID, targetID, now); err != nil {
			t.Fatalf("insert completed session: %v", err)
		}
		if _, err := db.ExecContext(ctx, `
insert into practice_session_events (id,session_id,seq_no,event_type,payload,created_at)
values ($1,$2,1,'session_completed','{}'::jsonb,$3)`, eventID, sessionID, now); err != nil {
			t.Fatalf("insert completion fact: %v", err)
		}
		if reportID != "" {
			if _, err := db.ExecContext(ctx, `
insert into feedback_reports (id,user_id,session_id,target_job_id,status,created_at,updated_at)
values ($1,$2,$3,$4,'ready',$5,$5)`, reportID, userID, sessionID, targetID, now); err != nil {
				t.Fatalf("insert ready report: %v", err)
			}
		}
	}

	if _, err := createWithResume("019f5585-0000-7000-8000-000000000005", "019f5585-0000-7000-8000-000000000006", "", sharedtypes.PracticeGoalBaseline, "", 60, otherResumeID); !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
		t.Fatalf("same-user wrong-resume plan error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `update target_jobs set summary = summary - 'provenance' where id = $1`, targetID); err != nil {
		t.Fatalf("remove provenance: %v", err)
	}
	if _, err := create("019f5585-0000-7000-8000-000000000007", "019f5585-0000-7000-8000-000000000008", "", sharedtypes.PracticeGoalBaseline, "", 60); !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
		t.Fatalf("missing-provenance plan error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `update target_jobs set summary = $1::jsonb where id = $2`, summary, targetID); err != nil {
		t.Fatalf("restore summary provenance: %v", err)
	}
	t.Log("target-resume-binding-and-provenance=PASS")
	if _, err := db.ExecContext(ctx, `update target_jobs set summary = jsonb_set(summary, '{interviewRounds,1,type}', '"Technical"'::jsonb) where id = $1`, targetID); err != nil {
		t.Fatalf("set uppercase round type: %v", err)
	}
	if _, err := create("019f5585-0000-7000-8000-000000000009", "019f5585-0000-7000-8000-00000000000a", "", sharedtypes.PracticeGoalBaseline, "", 60); !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
		t.Fatalf("uppercase round-type plan error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `update target_jobs set summary = $1::jsonb where id = $2`, summary, targetID); err != nil {
		t.Fatalf("restore lowercase round type: %v", err)
	}
	t.Log("canonical-round-type-case-sensitive=PASS")

	first, err := create("019f5585-0000-7000-8000-000000000010", "019f5585-0000-7000-8000-000000000011", "", sharedtypes.PracticeGoalBaseline, "", 60)
	if err != nil {
		t.Fatalf("create first baseline: %v", err)
	}
	if first.RoundID != "round-1-hr" || first.RoundSequence != 1 {
		t.Fatalf("first round = %+v", first)
	}
	reservation, err := repo.ReserveSessionStart(ctx, domain.StartSessionReservationInput{
		IdempotencyRecordID: "019f5585-0000-7000-8000-000000000040",
		SessionID:           "019f5585-0000-7000-8000-000000000041",
		UserID:              userID,
		PlanID:              first.ID,
		IdempotencyKeyHash:  "practice-round-integration-start",
		RequestFingerprint:  "practice-round-integration-fingerprint",
		ExpiresAt:           now.Add(time.Hour),
		Now:                 now,
	})
	if err != nil {
		t.Fatalf("reserve round-aware session: %v", err)
	}
	if reservation.RoundID != "round-1-hr" || reservation.RoundSequence != 1 || reservation.RoundName != "HR" || reservation.RoundFocus != "motivation" {
		t.Fatalf("reserved round context = %+v", reservation)
	}
	t.Log("canonical-round-prompt-context=PASS")
	if _, err := db.ExecContext(ctx, `update target_jobs set resume_id=$1 where id=$2`, otherResumeID, targetID); err != nil {
		t.Fatalf("rebind target to other resume: %v", err)
	}
	if _, err := repo.ReserveSessionStart(ctx, domain.StartSessionReservationInput{
		IdempotencyRecordID: "019f5585-0000-7000-8000-000000000042",
		SessionID:           "019f5585-0000-7000-8000-000000000043",
		UserID:              userID,
		PlanID:              first.ID,
		IdempotencyKeyHash:  "practice-round-stale-resume-start",
		RequestFingerprint:  "practice-round-stale-resume-start-fingerprint",
		ExpiresAt:           now.Add(time.Hour),
		Now:                 now,
	}); !errors.Is(err, domain.ErrPlanNotFound) {
		t.Fatalf("stale-resume plan start error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `update practice_sessions set status='waiting_user_input' where id=$1`, reservation.SessionID); err != nil {
		t.Fatalf("make probe session message-ready: %v", err)
	}
	if _, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID:   "019f5585-0000-7000-8000-000000000044",
		UserID:          userID,
		SessionID:       reservation.SessionID,
		ClientMessageID: "019f5585-0000-7000-8000-000000000045",
		Text:            "continue",
		Now:             now,
	}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("stale-resume session message error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `update target_jobs set resume_id=$1 where id=$2`, resumeID, targetID); err != nil {
		t.Fatalf("restore target resume binding: %v", err)
	}
	t.Log("start-and-send-bound-resume-fail-closed=PASS")
	if _, err := db.ExecContext(ctx, `delete from practice_sessions where id=$1`, reservation.SessionID); err != nil {
		t.Fatalf("delete reservation probe session: %v", err)
	}
	if _, err := db.ExecContext(ctx, `delete from idempotency_records where id=$1`, reservation.IdempotencyRecordID); err != nil {
		t.Fatalf("delete reservation probe idempotency: %v", err)
	}
	complete(first.ID, "019f5585-0000-7000-8000-000000000012", "019f5585-0000-7000-8000-000000000013", "")

	second, err := create("019f5585-0000-7000-8000-000000000020", "019f5585-0000-7000-8000-000000000021", "", sharedtypes.PracticeGoalBaseline, "round-2-technical", 60)
	if err != nil {
		t.Fatalf("create completed-ledger successor: %v", err)
	}
	if second.RoundID != "round-2-technical" || second.RoundSequence != 2 {
		t.Fatalf("second round = %+v", second)
	}
	t.Log("completed-ledger-successor=PASS")

	complete(second.ID, "019f5585-0000-7000-8000-000000000024", "019f5585-0000-7000-8000-000000000025", "019f5585-0000-7000-8000-000000000026")
	fourth, err := create("019f5585-0000-7000-8000-000000000030", "019f5585-0000-7000-8000-000000000031", "", sharedtypes.PracticeGoalBaseline, "round-4-manager", 45)
	if err != nil {
		t.Fatalf("create non-contiguous successor: %v", err)
	}
	if fourth.RoundID != "round-4-manager" || fourth.RoundSequence != 4 {
		t.Fatalf("non-contiguous successor = %+v", fourth)
	}
	t.Log("non-contiguous-successor=PASS")

	if _, err := create("019f5585-0000-7000-8000-000000000034", "019f5585-0000-7000-8000-000000000035", "", sharedtypes.PracticeGoalBaseline, "round-2-technical", 45); !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
		t.Fatalf("mismatched round/budget error = %v", err)
	}
	t.Log("round-budget-mismatch=PASS")

	complete(fourth.ID, "019f5585-0000-7000-8000-000000000036", "019f5585-0000-7000-8000-000000000037", "")
	if _, err := create("019f5585-0000-7000-8000-000000000038", "019f5585-0000-7000-8000-000000000039", "", sharedtypes.PracticeGoalBaseline, "", 45); !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
		t.Fatalf("all-complete baseline error = %v", err)
	}
	t.Log("all-rounds-complete-fail-closed=PASS")
}
