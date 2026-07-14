//go:build integration

package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
)

func TestIntegrationE2EP0047RejectsZeroAnswerCompletion(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for the reportable-completion integration gate")
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
		userID            = "019f56e5-0000-7000-8000-000000000101"
		resumeID          = "019f56e5-0000-7000-8000-000000000102"
		otherResumeID     = "019f56e5-0000-7000-8000-000000000117"
		targetID          = "019f56e5-0000-7000-8000-000000000103"
		zeroPlanID        = "019f56e5-0000-7000-8000-000000000104"
		pendingPlanID     = "019f56e5-0000-7000-8000-000000000115"
		readyPlanID       = "019f56e5-0000-7000-8000-000000000116"
		mismatchPlanID    = "019f56e5-0000-7000-8000-000000000118"
		zeroSessionID     = "019f56e5-0000-7000-8000-000000000105"
		pendingSessionID  = "019f56e5-0000-7000-8000-000000000106"
		readySessionID    = "019f56e5-0000-7000-8000-000000000107"
		mismatchSessionID = "019f56e5-0000-7000-8000-000000000119"
		reportID          = "019f56e5-0000-7000-8000-000000000110"
		jobID             = "019f56e5-0000-7000-8000-000000000111"
		sessionEventID    = "019f56e5-0000-7000-8000-000000000112"
		outboxEventID     = "019f56e5-0000-7000-8000-000000000113"
		auditEventID      = "019f56e5-0000-7000-8000-000000000114"
		completionGateKey = int64(9047001)
	)
	cleanup := func() {
		dropCompletionIntegrationGate(db)
		_, _ = db.ExecContext(context.Background(), `delete from async_jobs where id=$1 or resource_id=$2`, jobID, reportID)
		_, _ = db.ExecContext(context.Background(), `delete from outbox_events where id=$1`, outboxEventID)
		_, _ = db.ExecContext(context.Background(), `delete from audit_events where id=$1`, auditEventID)
		_, _ = db.ExecContext(context.Background(), `delete from users where id=$1`, userID)
	}
	cleanup()
	t.Cleanup(cleanup)
	if err := installCompletionIntegrationGate(ctx, db, reportID, completionGateKey); err != nil {
		t.Fatalf("install completion concurrency gate: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := db.ExecContext(ctx, `insert into users (id,email,display_name,created_at,updated_at) values ($1,$2,$3,$4,$4)`, userID, "reportable-completion@example.test", "Reportable Completion", now); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into resumes (
  id,user_id,title,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Reportable Resume','en','ready','{}'::jsonb,'resume body','paste','resume body','resume body','{}'::jsonb,$3,$3)`, resumeID, userID, now); err != nil {
		t.Fatalf("insert resume: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into resumes (
  id,user_id,title,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Other Resume','en','ready','{}'::jsonb,'other resume','paste','other resume','other resume','{}'::jsonb,$3,$3)`, otherResumeID, userID, now); err != nil {
		t.Fatalf("insert other resume: %v", err)
	}
	summary := `{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","durationMinutes":30,"focus":"system design"},{"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`
	if _, err := db.ExecContext(ctx, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,target_language,
  raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','en','complete jd',$4::jsonb,'{}'::jsonb,$5,$5)`, targetID, userID, resumeID, summary, now); err != nil {
		t.Fatalf("insert target job: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into target_job_requirements (id,target_job_id,kind,label,description,evidence_level,display_order,created_at)
values ('019f56e5-0000-7000-8000-000000000120',$1,'must_have','Go','production Go','explicit',1,$2)`, targetID, now); err != nil {
		t.Fatalf("insert target requirement: %v", err)
	}
	for _, planID := range []string{zeroPlanID, pendingPlanID, readyPlanID, mismatchPlanID} {
		if _, err := db.ExecContext(ctx, `
insert into practice_plans (
  id,user_id,target_job_id,goal,interviewer_persona,difficulty,language,time_budget_minutes,
  resume_id,status,round_id,round_sequence,created_at,updated_at
) values ($1,$2,$3,'baseline','hiring_manager','standard','en',30,$4,'ready','round-1-technical',1,$5,$5)`, planID, userID, targetID, resumeID, now); err != nil {
			t.Fatalf("insert plan %s: %v", planID, err)
		}
	}
	for _, item := range []struct{ sessionID, planID string }{
		{zeroSessionID, zeroPlanID},
		{pendingSessionID, pendingPlanID},
		{readySessionID, readyPlanID},
		{mismatchSessionID, mismatchPlanID},
	} {
		if _, err := db.ExecContext(ctx, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,started_at,created_at,updated_at)
values ($1,$2,$3,$4,'running','en',$5,$5,$5)`, item.sessionID, userID, item.planID, targetID, now); err != nil {
			t.Fatalf("insert session %s: %v", item.sessionID, err)
		}
		if _, err := db.ExecContext(ctx, `
insert into practice_messages (id,session_id,seq_no,role,content,created_at)
values ($1,$2,1,'assistant','Tell me about a project.',$3)`, integrationMessageID(item.sessionID, 1), item.sessionID, now); err != nil {
			t.Fatalf("insert opening %s: %v", item.sessionID, err)
		}
	}
	if _, err := db.ExecContext(ctx, `
	insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_generation,reply_lease_expires_at,created_at)
	values ($1,$2,2,'user','I led a migration.',$3,'pending',1,$4,$5)`, integrationMessageID(pendingSessionID, 2), pendingSessionID, integrationMessageID(pendingSessionID, 9), now.Add(domain.PracticeReplyLeaseDuration), now); err != nil {
		t.Fatalf("insert pending user message: %v", err)
	}
	readyUserID := integrationMessageID(readySessionID, 2)
	if _, err := db.ExecContext(ctx, `
	insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_generation,created_at)
	values ($1,$2,2,'user','I led a migration.',$3,'complete',1,$4)`, readyUserID, readySessionID, integrationMessageID(readySessionID, 9), now); err != nil {
		t.Fatalf("insert answered user message: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into practice_messages (id,session_id,seq_no,role,content,reply_to_message_id,created_at)
values ($1,$2,3,'assistant','What tradeoff did you make?',$3,$4)`, integrationMessageID(readySessionID, 3), readySessionID, readyUserID, now); err != nil {
		t.Fatalf("insert assistant reply: %v", err)
	}
	mismatchUserID := integrationMessageID(mismatchSessionID, 2)
	if _, err := db.ExecContext(ctx, `
	insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_generation,created_at)
	values ($1,$2,2,'user','I kept the scope bounded.',$3,'complete',1,$4)`, mismatchUserID, mismatchSessionID, integrationMessageID(mismatchSessionID, 9), now); err != nil {
		t.Fatalf("insert mismatch user message: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into practice_messages (id,session_id,seq_no,role,content,reply_to_message_id,created_at)
values ($1,$2,3,'assistant','Thanks.',$3,$4)`, integrationMessageID(mismatchSessionID, 3), mismatchSessionID, mismatchUserID, now); err != nil {
		t.Fatalf("insert mismatch assistant reply: %v", err)
	}

	repo := NewSQLRepository(db)
	for _, sessionID := range []string{zeroSessionID, pendingSessionID} {
		_, err := repo.CompleteSession(ctx, domain.CompleteSessionStoreInput{UserID: userID, SessionID: sessionID, Now: now})
		if !errors.Is(err, domain.ErrSessionNotReportable) {
			t.Fatalf("session %s error=%v want ErrSessionNotReportable", sessionID, err)
		}
		var sideEffects int
		if err := db.QueryRowContext(ctx, `
select (select count(*) from practice_session_events where session_id=$1 and event_type='session_completed')
     + (select count(*) from feedback_reports where session_id=$1)`, sessionID).Scan(&sideEffects); err != nil {
			t.Fatalf("count invalid completion side effects: %v", err)
		}
		if sideEffects != 0 {
			t.Fatalf("session %s produced %d completion/report side effects", sessionID, sideEffects)
		}
	}

	gateTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin completion concurrency gate: %v", err)
	}
	if _, err := gateTx.ExecContext(ctx, `select pg_advisory_xact_lock($1)`, completionGateKey); err != nil {
		_ = gateTx.Rollback()
		t.Fatalf("lock completion concurrency gate: %v", err)
	}
	type completionCallResult struct {
		result domain.CompleteSessionResult
		err    error
	}
	completionResult := make(chan completionCallResult, 1)
	go func() {
		result, callErr := repo.CompleteSession(ctx, domain.CompleteSessionStoreInput{
			UserID: userID, SessionID: readySessionID,
			ReportID:          reportID,
			JobID:             jobID,
			SessionEventID:    sessionEventID,
			OutboxEventID:     outboxEventID,
			AuditEventID:      auditEventID,
			ClientCompletedAt: now, Now: now,
		})
		completionResult <- completionCallResult{result: result, err: callErr}
	}()
	if err := waitForAdvisoryWaiter(ctx, db, completionGateKey); err != nil {
		_ = gateTx.Rollback()
		t.Fatal(err)
	}
	mutationConn, err := db.Conn(ctx)
	if err != nil {
		_ = gateTx.Rollback()
		t.Fatalf("open concurrent mutation connection: %v", err)
	}
	defer mutationConn.Close()
	var mutationPID int
	if err := mutationConn.QueryRowContext(ctx, `select pg_backend_pid()`).Scan(&mutationPID); err != nil {
		_ = gateTx.Rollback()
		t.Fatalf("select concurrent mutation pid: %v", err)
	}
	mutationResult := make(chan error, 1)
	go func() {
		_, updateErr := mutationConn.ExecContext(ctx, `update target_jobs set raw_jd_text='concurrent mutated jd', updated_at=$2 where id=$1`, targetID, now.Add(time.Second))
		mutationResult <- updateErr
	}()
	if err := waitForBackendLockWait(ctx, db, mutationPID); err != nil {
		_ = gateTx.Rollback()
		t.Fatal(err)
	}
	if err := gateTx.Commit(); err != nil {
		t.Fatalf("release completion concurrency gate: %v", err)
	}
	completed := <-completionResult
	result, err := completed.result, completed.err
	if err != nil {
		t.Fatalf("complete answered session: %v", err)
	}
	if err := <-mutationResult; err != nil {
		t.Fatalf("concurrent target mutation: %v", err)
	}
	if result.ReportID == "" || result.Job.ID == "" {
		t.Fatalf("answered completion result=%+v", result)
	}
	var reportJobMaxAttempts int32
	if err := db.QueryRowContext(ctx, `select max_attempts from async_jobs where id=$1`, result.Job.ID).Scan(&reportJobMaxAttempts); err != nil {
		t.Fatalf("read report job max_attempts: %v", err)
	}
	if reportJobMaxAttempts != 5 {
		t.Fatalf("report job max_attempts=%d want generic infrastructure default 5", reportJobMaxAttempts)
	}
	if result.GenerationContext.SchemaVersion != domain.ReportContextSchemaVersion ||
		result.GenerationContext.TargetJob.RawJD != "complete jd" ||
		result.GenerationContext.Resume.SourceSnapshot != "resume body" ||
		result.GenerationContext.Conversation.MessageCount != 3 ||
		result.GenerationContext.Conversation.LastMessageSeqNo != 3 ||
		!result.GenerationContext.HasNextRound {
		t.Fatalf("answered completion context=%+v", result.GenerationContext)
	}
	var persistedRaw []byte
	if err := db.QueryRowContext(ctx, `select generation_context from feedback_reports where id=$1`, result.ReportID).Scan(&persistedRaw); err != nil {
		t.Fatalf("load persisted generation context: %v", err)
	}
	var persisted domain.ReportContextSnapshot
	if err := json.Unmarshal(persistedRaw, &persisted); err != nil {
		t.Fatalf("decode persisted generation context: %v", err)
	}
	if !reflect.DeepEqual(persisted, result.GenerationContext) {
		t.Fatalf("persisted context changed inside completion transaction: got=%+v want=%+v", persisted, result.GenerationContext)
	}

	if _, err := db.ExecContext(ctx, `update target_jobs set resume_id=$1, raw_jd_text='mutated jd', updated_at=$3 where id=$2`, otherResumeID, targetID, now.Add(time.Second)); err != nil {
		t.Fatalf("mutate target after completion: %v", err)
	}
	if _, err := db.ExecContext(ctx, `update resumes set parsed_text_snapshot='mutated resume', updated_at=$2 where id=$1`, resumeID, now.Add(time.Second)); err != nil {
		t.Fatalf("mutate resume after completion: %v", err)
	}
	replay, err := repo.CompleteSession(ctx, domain.CompleteSessionStoreInput{UserID: userID, SessionID: readySessionID, Now: now.Add(2 * time.Second)})
	if err != nil {
		t.Fatalf("replay answered completion: %v", err)
	}
	if !replay.Replay || !reflect.DeepEqual(replay.GenerationContext, result.GenerationContext) {
		t.Fatalf("replay rebuilt mutable context: replay=%+v original=%+v", replay, result.GenerationContext)
	}

	_, err = repo.CompleteSession(ctx, domain.CompleteSessionStoreInput{UserID: userID, SessionID: mismatchSessionID, Now: now.Add(3 * time.Second)})
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("mismatched target/resume completion error=%v want ErrSessionConflict", err)
	}
	var mismatchSideEffects int
	if err := db.QueryRowContext(ctx, `
select (select count(*) from practice_session_events where session_id=$1 and event_type='session_completed')
     + (select count(*) from feedback_reports where session_id=$1)`, mismatchSessionID).Scan(&mismatchSideEffects); err != nil {
		t.Fatalf("count mismatch completion side effects: %v", err)
	}
	if mismatchSideEffects != 0 {
		t.Fatalf("mismatched completion produced %d completion/report side effects", mismatchSideEffects)
	}
	t.Log("zero_answer_side_effect_count=0")
	t.Log("pending_reply_side_effect_count=0")
	t.Log("snapshot_schema_version=report-context.v1")
	t.Log("concurrent_mutation_blocked=true")
	t.Log("snapshot_replay_equal=true")
	t.Log("mismatch_side_effect_count=0")
	t.Log("ZERO_ANSWER_COMPLETION_REJECTED_PASS")
	t.Log("REPORT_CONTEXT_SNAPSHOT_PASS")
	t.Log("REPORT_CONTEXT_REPLAY_PASS")
}

func integrationMessageID(sessionID string, suffix int) string {
	return fmt.Sprintf("019f56e5-0000-7000-8000-000000%s%03d", sessionID[len(sessionID)-3:], suffix)
}

func installCompletionIntegrationGate(ctx context.Context, db *sql.DB, reportID string, lockKey int64) error {
	dropCompletionIntegrationGate(db)
	ddl := fmt.Sprintf(`
create function test_e2ep0047_completion_gate() returns trigger language plpgsql as $$
begin
  if new.id = '%s'::uuid then
    perform pg_advisory_xact_lock(%d);
  end if;
  return new;
end
$$;
create trigger test_e2ep0047_completion_gate
before insert on feedback_reports
for each row execute function test_e2ep0047_completion_gate()`, reportID, lockKey)
	_, err := db.ExecContext(ctx, ddl)
	return err
}

func dropCompletionIntegrationGate(db *sql.DB) {
	if db == nil {
		return
	}
	_, _ = db.ExecContext(context.Background(), `drop trigger if exists test_e2ep0047_completion_gate on feedback_reports`)
	_, _ = db.ExecContext(context.Background(), `drop function if exists test_e2ep0047_completion_gate()`)
}

func waitForAdvisoryWaiter(ctx context.Context, db *sql.DB, lockKey int64) error {
	return waitForIntegrationCondition(ctx, func() (bool, error) {
		var waiting bool
		err := db.QueryRowContext(ctx, `
select exists(
  select 1 from pg_locks
  where locktype='advisory' and classid=0 and objid=$1 and not granted
)`, lockKey).Scan(&waiting)
		return waiting, err
	}, "completion did not reach the frozen-context transaction gate")
}

func waitForBackendLockWait(ctx context.Context, db *sql.DB, pid int) error {
	return waitForIntegrationCondition(ctx, func() (bool, error) {
		var waiting bool
		err := db.QueryRowContext(ctx, `
select coalesce((select wait_event_type='Lock' from pg_stat_activity where pid=$1), false)`, pid).Scan(&waiting)
		return waiting, err
	}, "concurrent target mutation was not blocked by completion locks")
}

func waitForIntegrationCondition(ctx context.Context, probe func() (bool, error), message string) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		ok, err := probe()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", message, ctx.Err())
		case <-ticker.C:
		}
	}
}
