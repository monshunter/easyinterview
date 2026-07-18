//go:build integration

package practice

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryIntegration_StartRecoversSamePlanActiveSession(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for the active-session recovery integration gate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	const (
		userID          = "019f7600-0000-7000-8000-000000000001"
		otherUserID     = "019f7600-0000-7000-8000-000000000002"
		resumeID        = "019f7600-0000-7000-8000-000000000003"
		targetID        = "019f7600-0000-7000-8000-000000000004"
		runningPlanID   = "019f7600-0000-7000-8000-000000000005"
		concurrentPlan  = "019f7600-0000-7000-8000-000000000006"
		runningSession  = "019f7600-0000-7000-8000-000000000007"
		openingID       = "019f7600-0000-7000-8000-000000000008"
		recoveryRecord  = "019f7600-0000-7000-8000-000000000009"
		unusedSession   = "019f7600-0000-7000-8000-00000000000a"
		otherRecord     = "019f7600-0000-7000-8000-00000000000b"
		otherSession    = "019f7600-0000-7000-8000-00000000000c"
		concurrentRecA  = "019f7600-0000-7000-8000-00000000000d"
		concurrentRecB  = "019f7600-0000-7000-8000-00000000000e"
		concurrentSessA = "019f7600-0000-7000-8000-00000000000f"
		concurrentSessB = "019f7600-0000-7000-8000-000000000010"
	)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from audit_events where user_id in ($1,$2) or actor_id in ($1,$2)`, userID, otherUserID)
		_, _ = db.ExecContext(context.Background(), `delete from users where id in ($1,$2)`, userID, otherUserID)
	}
	cleanup()
	t.Cleanup(cleanup)

	now := time.Now().UTC().Truncate(time.Microsecond)
	mustExecActiveRecovery(t, ctx, db, `insert into users (id,email,display_name,created_at,updated_at) values
($1,'active-recovery@example.test','Active Recovery',$3,$3),
($2,'active-recovery-other@example.test','Active Recovery Other',$3,$3)`, userID, otherUserID, now)
	mustExecActiveRecovery(t, ctx, db, `
insert into resumes (
  id,user_id,title,display_name,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Active Recovery Resume','Active Recovery Resume','zh-CN','ready','{}'::jsonb,'完整简历','paste','完整简历','完整简历','{}'::jsonb,$3,$3)`, resumeID, userID, now)
	summary := `{"interviewRounds":[{"sequence":1,"type":"technical","name":"技术面","durationMinutes":45,"focus":"系统设计"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"zh-CN","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`
	mustExecActiveRecovery(t, ctx, db, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,target_language,raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','zh-CN','完整 JD',$4::jsonb,'{}'::jsonb,$5,$5)`, targetID, userID, resumeID, summary, now)
	for _, planID := range []string{runningPlanID, concurrentPlan} {
		mustExecActiveRecovery(t, ctx, db, `
insert into practice_plans (
  id,user_id,target_job_id,goal,interviewer_persona,difficulty,language,time_budget_minutes,
  resume_id,focus_dimension_codes,status,round_id,round_sequence,created_at,updated_at
) values ($1,$2,$3,'baseline','hiring_manager','standard','zh-CN',45,$4,'{}'::text[],'ready','round-1-technical',1,$5,$5)`, planID, userID, targetID, resumeID, now)
	}
	mustExecActiveRecovery(t, ctx, db, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,started_at,created_at,updated_at)
values ($1,$2,$3,$4,'running','zh-CN',$5,$5,$5)`, runningSession, userID, runningPlanID, targetID, now)
	mustExecActiveRecovery(t, ctx, db, `
insert into practice_messages (id,session_id,seq_no,role,content,created_at)
values ($1,$2,1,'assistant','原开场消息',$3)`, openingID, runningSession, now)

	repo := NewSQLRepository(db)
	before := readActiveRecoveryCounts(t, ctx, db, runningPlanID, runningSession)
	input := domain.StartSessionReservationInput{
		IdempotencyRecordID: recoveryRecord,
		SessionID:           unusedSession,
		UserID:              userID,
		PlanID:              runningPlanID,
		IdempotencyKeyHash:  "active-recovery-new-key",
		RequestFingerprint:  "active-recovery-fingerprint",
		ExpiresAt:           now.Add(time.Hour),
		Now:                 now,
	}
	reservation, err := repo.ReserveSessionStart(ctx, input)
	if err != nil {
		t.Fatalf("reserve active-session recovery: %v", err)
	}
	if reservation.RecoverSession == nil || reservation.RecoverSession.ID != runningSession || reservation.ReplaySession != nil {
		t.Fatalf("active recovery reservation = %+v", reservation)
	}
	afterReserve := readActiveRecoveryCounts(t, ctx, db, runningPlanID, runningSession)
	if afterReserve != before {
		t.Fatalf("recovery reservation changed opening side effects: before=%+v after=%+v", before, afterReserve)
	}
	var pendingStatus string
	if err := db.QueryRowContext(ctx, `select status from idempotency_records where id=$1`, recoveryRecord).Scan(&pendingStatus); err != nil || pendingStatus != "pending" {
		t.Fatalf("recovery key before finalization status=%q err=%v", pendingStatus, err)
	}

	recovered, err := repo.CommitSessionStartRecovery(ctx, domain.CommitSessionStartRecoveryInput{
		IdempotencyRecordID: recoveryRecord,
		SessionID:           runningSession,
		UserID:              userID,
		RecoveredAt:         now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("finalize active-session recovery: %v", err)
	}
	if recovered.ID != runningSession || recovered.Status != sharedtypes.SessionStatusRunning || len(recovered.Messages) != 1 || recovered.Messages[0].ID != openingID {
		t.Fatalf("recovered session = %+v", recovered)
	}
	afterFinalize := readActiveRecoveryCounts(t, ctx, db, runningPlanID, runningSession)
	if afterFinalize != before {
		t.Fatalf("recovery finalization changed opening side effects: before=%+v after=%+v", before, afterFinalize)
	}

	replayInput := input
	replayInput.IdempotencyRecordID = "019f7600-0000-7000-8000-000000000011"
	replayInput.SessionID = "019f7600-0000-7000-8000-000000000012"
	replayInput.Now = now.Add(2 * time.Second)
	replayed, err := repo.ReserveSessionStart(ctx, replayInput)
	if err != nil || replayed.ReplaySession == nil || replayed.ReplaySession.ID != runningSession || replayed.RecoverSession != nil {
		t.Fatalf("same-key recovered replay=%+v err=%v", replayed, err)
	}
	mismatch := replayInput
	mismatch.RequestFingerprint = "changed-fingerprint"
	if _, err := repo.ReserveSessionStart(ctx, mismatch); !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("same-key fingerprint mismatch error=%v want ErrSessionConflict", err)
	}

	if _, err := repo.ReserveSessionStart(ctx, domain.StartSessionReservationInput{
		IdempotencyRecordID: otherRecord,
		SessionID:           otherSession,
		UserID:              otherUserID,
		PlanID:              runningPlanID,
		IdempotencyKeyHash:  "cross-user-key",
		RequestFingerprint:  "cross-user-fingerprint",
		ExpiresAt:           now.Add(time.Hour),
		Now:                 now,
	}); !errors.Is(err, domain.ErrPlanNotFound) {
		t.Fatalf("cross-user recovery error=%v want ErrPlanNotFound", err)
	}

	concurrentInputs := []domain.StartSessionReservationInput{
		{
			IdempotencyRecordID: concurrentRecA, SessionID: concurrentSessA, UserID: userID, PlanID: concurrentPlan,
			IdempotencyKeyHash: "concurrent-key-a", RequestFingerprint: "concurrent-fingerprint",
			ExpiresAt: now.Add(time.Hour), Now: now,
		},
		{
			IdempotencyRecordID: concurrentRecB, SessionID: concurrentSessB, UserID: userID, PlanID: concurrentPlan,
			IdempotencyKeyHash: "concurrent-key-b", RequestFingerprint: "concurrent-fingerprint",
			ExpiresAt: now.Add(time.Hour), Now: now,
		},
	}
	type reserveOutcome struct {
		reservation domain.SessionReservation
		err         error
	}
	outcomes := make(chan reserveOutcome, len(concurrentInputs))
	start := make(chan struct{})
	var workers sync.WaitGroup
	for _, concurrentInput := range concurrentInputs {
		concurrentInput := concurrentInput
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			reservation, reserveErr := repo.ReserveSessionStart(ctx, concurrentInput)
			outcomes <- reserveOutcome{reservation: reservation, err: reserveErr}
		}()
	}
	close(start)
	workers.Wait()
	close(outcomes)
	var created, recoveredActive int
	var selectedSessionID string
	for outcome := range outcomes {
		if outcome.err != nil {
			t.Fatalf("concurrent start reservation: %v", outcome.err)
		}
		if outcome.reservation.RecoverSession != nil {
			recoveredActive++
			selectedSessionID = outcome.reservation.RecoverSession.ID
		} else {
			created++
			selectedSessionID = outcome.reservation.SessionID
		}
	}
	if created != 1 || recoveredActive != 1 {
		t.Fatalf("concurrent decisions created=%d recovered=%d", created, recoveredActive)
	}
	var activeCount int
	if err := db.QueryRowContext(ctx, `select count(*) from practice_sessions where user_id=$1 and plan_id=$2 and status in ('queued','running')`, userID, concurrentPlan).Scan(&activeCount); err != nil {
		t.Fatalf("count concurrent active sessions: %v", err)
	}
	if activeCount != 1 || selectedSessionID == runningSession {
		t.Fatalf("concurrent active count=%d selected=%s runningPlanSession=%s", activeCount, selectedSessionID, runningSession)
	}
	if _, err := repo.ReserveSessionStart(ctx, concurrentInputs[0]); !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("same pending key error=%v want ErrSessionConflict", err)
	}
	t.Log("active-session-start-recovery=PASS")
}

type activeRecoveryCounts struct {
	Sessions int
	Messages int
	Events   int
	Outbox   int
	Audit    int
	AITasks  int
}

func readActiveRecoveryCounts(t *testing.T, ctx context.Context, db *sql.DB, planID, sessionID string) activeRecoveryCounts {
	t.Helper()
	var counts activeRecoveryCounts
	err := db.QueryRowContext(ctx, `
select
  (select count(*) from practice_sessions where plan_id=$1),
  (select count(*) from practice_messages where session_id=$2),
  (select count(*) from practice_session_events where session_id=$2),
  (select count(*) from outbox_events where aggregate_type='practice_session' and aggregate_id=$2),
  (select count(*) from audit_events where resource_type='practice_session' and resource_id=$2),
  (select count(*) from ai_task_runs where resource_type='practice_session' and resource_id=$2)`, planID, sessionID).Scan(
		&counts.Sessions, &counts.Messages, &counts.Events, &counts.Outbox, &counts.Audit, &counts.AITasks,
	)
	if err != nil {
		t.Fatalf("read active recovery side-effect counts: %v", err)
	}
	return counts
}

func mustExecActiveRecovery(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec active recovery fixture: %v", err)
	}
}
