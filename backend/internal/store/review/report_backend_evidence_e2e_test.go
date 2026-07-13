package review

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestE2EP0056ReportBackendEvidence(t *testing.T) {
	requireStoreCompletionOwnerEvidence(t)
	t.Log("REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS")

	fixture := newStoreReportEvidenceFixture(t, "0056")
	fixture.seedBase(t)
	sessionID, reportID, jobID := fixture.seedGeneratingReport(t, 10, storeReportSnapshotOptions{})
	repository := NewRepository(fixture.db)

	before, err := repository.LoadReportContext(fixture.ctx, reportID)
	if err != nil {
		t.Fatalf("load frozen report context before mutable edits: %v", err)
	}
	fixture.exec(t, `update target_jobs set title='Mutable title changed', company_name='Mutable company changed', raw_jd_text='Mutable JD changed', updated_at=$2 where id=$1`, fixture.targetID, fixture.now.Add(time.Second))
	fixture.exec(t, `update resumes set display_name='Mutable resume changed', parsed_text_snapshot='Mutable resume body changed', updated_at=$2 where id=$1`, fixture.resumeID, fixture.now.Add(time.Second))
	after, err := repository.LoadReportContext(fixture.ctx, reportID)
	if err != nil {
		t.Fatalf("load frozen report context after mutable edits: %v", err)
	}
	if !reflect.DeepEqual(before.FrozenContext, after.FrozenContext) || !reflect.DeepEqual(before.Messages, after.Messages) {
		t.Fatal("report input changed after mutable TargetJob/Resume edits")
	}

	content := storeValidDirectReportContent()
	err = repository.PersistReportResult(fixture.ctx, reviewdomain.ReportResultPersistence{
		UserID: fixture.userID, ReportID: reportID, SessionID: sessionID, TargetJobID: fixture.targetID, AsyncJobID: jobID,
		ClaimedAttempts: 1,
		OutboxEventID:   fixture.id(40), AuditEventID: fixture.id(41),
		PromptVersion: "v0.2.0", RubricVersion: "v0.2.0", ModelID: "evidence-model", Provider: "evidence-provider",
		Language: "en", FeatureFlag: "none", DataSourceVersion: practicedomain.ReportContextSchemaVersion,
		Now: fixture.now.Add(2 * time.Second), Content: content,
	})
	if err != nil {
		t.Fatalf("persist direct ready report: %v", err)
	}
	stored, err := repository.GetFeedbackReport(fixture.ctx, fixture.userID, reportID)
	if err != nil {
		t.Fatalf("read direct ready report: %v", err)
	}
	if stored.Status != sharedtypes.ReportStatusReady || stored.Summary == nil || *stored.Summary != content.Summary || stored.PreparednessLevel == nil || *stored.PreparednessLevel != content.PreparednessLevel ||
		!reflect.DeepEqual(stored.Context, reviewdomain.ProjectFrozenReportContext(before.FrozenContext)) || len(stored.DimensionAssessments) != 1 || len(stored.Issues) != 1 || len(stored.NextActions) != 1 ||
		!reflect.DeepEqual(stored.Issues[0].SourceMessageSeqNos, []int32{2}) || !reflect.DeepEqual(stored.RetryFocusDimensionCodes, []string{"technical_depth"}) {
		t.Fatal("stored direct ready report drifted")
	}
	if stored.Provenance == nil || stored.Provenance.PromptVersion != "v0.2.0" || stored.Provenance.RubricVersion != "v0.2.0" || stored.Provenance.ModelID != "evidence-model" || stored.Provenance.DataSourceVersion != practicedomain.ReportContextSchemaVersion {
		t.Fatal("stored report provenance drifted")
	}

	t.Log("direct_ready_status=ready")
	t.Log("frozen_context_read_equal=true")
	t.Log("REPORT_DIRECT_READY_PASS")
	t.Log("REPORT_FROZEN_CONTEXT_READ_PASS")
}

func TestE2EP0058ReportFailureBackendEvidence(t *testing.T) {
	requireStoreCompletionOwnerEvidence(t)
	fixture := newStoreReportEvidenceFixture(t, "0058")
	fixture.seedBase(t)
	repository := NewRepository(fixture.db)

	t.Run("context mismatch is terminal and empty", func(t *testing.T) {
		_, reportID, jobID := fixture.seedGeneratingReport(t, 10, storeReportSnapshotOptions{MessageCount: 4, LastMessageSeqNo: 4})
		ai := &storeReportEvidenceAI{}
		service := fixture.service(repository, ai, 100)
		outcome := service.GenerateReport(fixture.ctx, reviewdomain.AsyncJob{JobID: jobID, ResourceID: reportID, Attempts: 1, MaxAttempts: 4})
		if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiOutputInvalid || ai.calls != 0 {
			t.Fatalf("context mismatch did not fail closed: succeeded=%t retryable=%t code=%s providerCalls=%d", outcome.Succeeded, outcome.Retryable, outcome.ErrorCode, ai.calls)
		}
		fixture.assertFailedReadyColumnsEmpty(t, reportID, sharederrors.CodeAiOutputInvalid)
		t.Log("context_mismatch_fail_closed=true")
		t.Log("REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS")
	})

	t.Run("oversized context persists terminal failure without provider", func(t *testing.T) {
		_, reportID, jobID := fixture.seedGeneratingReport(t, 20, storeReportSnapshotOptions{RawJDBytes: 60_000})
		ai := &storeReportEvidenceAI{}
		service := fixture.service(repository, ai, 120)
		outcome := service.GenerateReport(fixture.ctx, reviewdomain.AsyncJob{JobID: jobID, ResourceID: reportID, Attempts: 1, MaxAttempts: 4})
		if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeReportContextTooLarge || ai.calls != 0 {
			t.Fatalf("oversized context did not fail terminal: succeeded=%t retryable=%t code=%s providerCalls=%d", outcome.Succeeded, outcome.Retryable, outcome.ErrorCode, ai.calls)
		}
		fixture.assertFailedReadyColumnsEmpty(t, reportID, sharederrors.CodeReportContextTooLarge)
		t.Log("context_too_large_status=failed")
		t.Log("context_too_large_provider_calls=0")
	})

	t.Run("action local retry persists direct ready", func(t *testing.T) {
		_, reportID, jobID := fixture.seedGeneratingReport(t, 30, storeReportSnapshotOptions{})
		ai := &storeReportEvidenceAI{results: []storeReportEvidenceAIResult{
			{content: `{"summary":"bad"}`},
			{content: storeValidDirectReportJSON()},
		}}
		var waits []time.Duration
		service := fixture.serviceWithWaits(repository, ai, 140, &waits)
		outcome := service.GenerateReport(fixture.ctx, reviewdomain.AsyncJob{JobID: jobID, ResourceID: reportID, Attempts: 1, MaxAttempts: 4})
		if !outcome.Succeeded || !outcome.AsyncJobFinalized || ai.calls != 2 {
			t.Fatalf("single repair failed: succeeded=%t finalized=%t code=%s providerCalls=%d", outcome.Succeeded, outcome.AsyncJobFinalized, outcome.ErrorCode, ai.calls)
		}
		var status string
		if err := fixture.db.QueryRowContext(fixture.ctx, `select status from feedback_reports where id=$1`, reportID).Scan(&status); err != nil {
			t.Fatalf("read repaired report state: %v", err)
		}
		if status != "ready" || !reflect.DeepEqual(waits, []time.Duration{10 * time.Second}) {
			t.Fatalf("repaired report state=%s waits=%v", status, waits)
		}
		t.Log("action_retry_call_count=2")
		t.Log("action_retry_waits=10s")
	})

	t.Run("four invalid outputs end one action and a new action resets", func(t *testing.T) {
		_, reportID, jobID := fixture.seedGeneratingReport(t, 50, storeReportSnapshotOptions{})
		ai := &storeReportEvidenceAI{results: []storeReportEvidenceAIResult{
			{content: `{"summary":"bad"}`},
			{content: `{"summary":"still bad"}`},
			{content: `{"summary":"bad again"}`},
			{content: `{"summary":"still invalid"}`},
		}}
		var waits []time.Duration
		service := fixture.serviceWithWaits(repository, ai, 160, &waits)
		outcome := service.GenerateReport(fixture.ctx, reviewdomain.AsyncJob{JobID: jobID, ResourceID: reportID, Attempts: 1, MaxAttempts: 4})
		if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiOutputInvalid || ai.calls != 4 {
			t.Fatalf("four invalid outputs did not fail terminal: succeeded=%t retryable=%t code=%s providerCalls=%d", outcome.Succeeded, outcome.Retryable, outcome.ErrorCode, ai.calls)
		}
		fixture.assertFailedReadyColumnsEmpty(t, reportID, sharederrors.CodeAiOutputInvalid)
		if !reflect.DeepEqual(waits, []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}) {
			t.Fatalf("action retry waits=%v", waits)
		}

		_, resetReportID, resetJobID := fixture.seedGeneratingReport(t, 70, storeReportSnapshotOptions{})
		fixture.exec(t, `update async_jobs set attempts=3, max_attempts=5, updated_at=$2 where id=$1`, resetJobID, fixture.now.Add(3*time.Second))
		resetAI := &storeReportEvidenceAI{results: []storeReportEvidenceAIResult{{content: storeValidDirectReportJSON()}}}
		resetOutcome := fixture.service(repository, resetAI, 180).GenerateReport(fixture.ctx, reviewdomain.AsyncJob{
			JobID: resetJobID, ResourceID: resetReportID, Attempts: 3, MaxAttempts: 5,
		})
		if !resetOutcome.Succeeded || resetAI.calls != 1 {
			t.Fatalf("new action did not reset: outcome=%+v providerCalls=%d", resetOutcome, resetAI.calls)
		}
		t.Log("four_invalid_status=failed")
		t.Log("action_retry_call_count=4")
		t.Log("action_retry_waits=10s,20s,40s")
		t.Log("new_action_reset_call_count=1")
	})

	t.Log("failed_ready_columns_empty=true")
}

func TestReportPersistenceFencesStaleLeaseBeforeAnyDomainSideEffect(t *testing.T) {
	fixture := newStoreReportEvidenceFixture(t, "0fce")
	fixture.seedBase(t)
	repository := NewRepository(fixture.db)
	kernelStore := runner.NewSQLStore(fixture.db)

	t.Run("stale worker cannot pass provider admission lease fence", func(t *testing.T) {
		_, _, jobID := fixture.seedGeneratingReport(t, 5, storeReportSnapshotOptions{})
		claimed := fixture.reapAndTakeOver(t, kernelStore, jobID)
		if err := repository.AssertCurrentReportJobLease(fixture.ctx, jobID, 1); !errors.Is(err, runner.ErrStaleLease) {
			t.Fatalf("stale provider admission err=%v want ErrStaleLease", err)
		}
		if err := repository.AssertCurrentReportJobLease(fixture.ctx, jobID, claimed.Attempts); err != nil {
			t.Fatalf("current provider admission: %v", err)
		}
		if err := kernelStore.FinalizeAsyncJob(fixture.ctx, jobID, claimed.Attempts, runner.JobOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid}, fixture.now.Add(4*time.Second), fixture.now.Add(4*time.Second)); err != nil {
			t.Fatalf("finalize current reservation test job: %v", err)
		}
	})

	t.Run("stale success cannot write report outbox audit or job", func(t *testing.T) {
		sessionID, reportID, jobID := fixture.seedGeneratingReport(t, 10, storeReportSnapshotOptions{})
		claimed := fixture.reapAndTakeOver(t, kernelStore, jobID)
		staleOutboxID, staleAuditID := fixture.id(80), fixture.id(81)
		in := reviewdomain.ReportResultPersistence{
			UserID: fixture.userID, ReportID: reportID, SessionID: sessionID, TargetJobID: fixture.targetID,
			AsyncJobID: jobID, ClaimedAttempts: 1, OutboxEventID: staleOutboxID, AuditEventID: staleAuditID,
			PromptVersion: "v0.2.0", RubricVersion: "v0.2.0", ModelID: "evidence-model", Provider: "evidence-provider",
			Language: "en", FeatureFlag: "none", DataSourceVersion: practicedomain.ReportContextSchemaVersion,
			Now: fixture.now.Add(4 * time.Second), Content: storeValidDirectReportContent(),
		}
		if err := repository.PersistReportResult(fixture.ctx, in); !errors.Is(err, runner.ErrStaleLease) {
			t.Fatalf("stale success err=%v want ErrStaleLease", err)
		}
		fixture.assertLeaseFenceState(t, reportID, jobID, staleOutboxID, staleAuditID, "generating", "running", claimed.Attempts, 0)

		in.ClaimedAttempts = claimed.Attempts
		in.OutboxEventID, in.AuditEventID = fixture.id(82), fixture.id(83)
		if err := repository.PersistReportResult(fixture.ctx, in); err != nil {
			t.Fatalf("current lease success: %v", err)
		}
		fixture.assertLeaseFenceState(t, reportID, jobID, in.OutboxEventID, in.AuditEventID, "ready", "succeeded", claimed.Attempts, 2)
	})

	t.Run("stale failure cannot write report outbox audit or job", func(t *testing.T) {
		sessionID, reportID, jobID := fixture.seedGeneratingReport(t, 30, storeReportSnapshotOptions{})
		claimed := fixture.reapAndTakeOver(t, kernelStore, jobID)
		staleOutboxID, staleAuditID := fixture.id(90), fixture.id(91)
		in := reviewdomain.ReportFailurePersistence{
			UserID: fixture.userID, ReportID: reportID, SessionID: sessionID,
			AsyncJobID: jobID, ClaimedAttempts: 1, OutboxEventID: staleOutboxID, AuditEventID: staleAuditID,
			ErrorCode: sharederrors.CodeAiProviderTimeout, Retryable: true,
			MaxAttempts: 4, Now: fixture.now.Add(4 * time.Second),
		}
		if err := repository.PersistReportFailure(fixture.ctx, in); !errors.Is(err, runner.ErrStaleLease) {
			t.Fatalf("stale failure err=%v want ErrStaleLease", err)
		}
		fixture.assertLeaseFenceState(t, reportID, jobID, staleOutboxID, staleAuditID, "generating", "running", claimed.Attempts, 0)

		in.ClaimedAttempts = claimed.Attempts
		in.OutboxEventID, in.AuditEventID = fixture.id(92), fixture.id(93)
		if err := repository.PersistReportFailure(fixture.ctx, in); err != nil {
			t.Fatalf("current lease failure persistence: %v", err)
		}
		if err := kernelStore.FinalizeAsyncJob(fixture.ctx, jobID, claimed.Attempts, runner.JobOutcome{Retryable: true, ErrorCode: in.ErrorCode}, fixture.now.Add(24*time.Second), fixture.now.Add(4*time.Second)); err != nil {
			t.Fatalf("current lease kernel finalize: %v", err)
		}
		fixture.assertLeaseFenceState(t, reportID, jobID, in.OutboxEventID, in.AuditEventID, "queued", "queued", claimed.Attempts, 2)
	})

	t.Run("failure persistence renews an expired lease before kernel finalize", func(t *testing.T) {
		sessionID, reportID, jobID := fixture.seedGeneratingReport(t, 50, storeReportSnapshotOptions{})
		failureNow := fixture.now.Add(10 * time.Minute)
		in := reviewdomain.ReportFailurePersistence{
			UserID: fixture.userID, ReportID: reportID, SessionID: sessionID,
			AsyncJobID: jobID, ClaimedAttempts: 1, OutboxEventID: fixture.id(100), AuditEventID: fixture.id(101),
			ErrorCode: sharederrors.CodeAiProviderTimeout, Retryable: true, MaxAttempts: 4, Now: failureNow,
		}
		if err := repository.PersistReportFailure(fixture.ctx, in); err != nil {
			t.Fatalf("persist failure with expired locked_at: %v", err)
		}
		if reclaimed, err := kernelStore.ReclaimExpiredLeases(fixture.ctx, []string{string(sharedjobs.JobTypeReportGenerate)}, failureNow.Add(-5*time.Minute), failureNow); err != nil || reclaimed != 0 {
			t.Fatalf("reaper interposed after failure persistence: reclaimed=%d err=%v", reclaimed, err)
		}
		if err := kernelStore.FinalizeAsyncJob(fixture.ctx, jobID, 1, runner.JobOutcome{Retryable: true, ErrorCode: in.ErrorCode}, failureNow.Add(10*time.Second), failureNow); err != nil {
			t.Fatalf("finalize renewed failure lease: %v", err)
		}
		fixture.assertLeaseFenceState(t, reportID, jobID, in.OutboxEventID, in.AuditEventID, "queued", "queued", 1, 2)
	})
}

type storeReportEvidenceFixture struct {
	db       *sql.DB
	ctx      context.Context
	cancel   context.CancelFunc
	prefix   string
	userID   string
	resumeID string
	targetID string
	planID   string
	now      time.Time
}

func newStoreReportEvidenceFixture(t *testing.T, prefix string) *storeReportEvidenceFixture {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping isolated report evidence database test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open report evidence postgres: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		_ = db.Close()
		t.Fatalf("ping report evidence postgres: %v", err)
	}
	fixture := &storeReportEvidenceFixture{
		db: db, ctx: ctx, cancel: cancel, prefix: prefix,
		userID: storeReportEvidenceUUID(prefix, 1), resumeID: storeReportEvidenceUUID(prefix, 2),
		targetID: storeReportEvidenceUUID(prefix, 3), planID: storeReportEvidenceUUID(prefix, 4),
		now: time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC),
	}
	fixture.cleanup(t)
	t.Cleanup(func() {
		fixture.cleanup(t)
		cancel()
		_ = db.Close()
	})
	return fixture
}

func (f *storeReportEvidenceFixture) id(index int) string {
	return storeReportEvidenceUUID(f.prefix, index)
}

func (f *storeReportEvidenceFixture) cleanup(t *testing.T) {
	t.Helper()
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, statement := range []string{
		`delete from outbox_events where aggregate_id in (select id from feedback_reports where user_id=$1)`,
		`delete from audit_events where user_id=$1 or resource_id in (select id from feedback_reports where user_id=$1)`,
		`delete from async_jobs where resource_id in (select id from feedback_reports where user_id=$1)`,
		`delete from users where id=$1`,
	} {
		if _, err := f.db.ExecContext(cleanupCtx, statement, f.userID); err != nil {
			t.Fatalf("cleanup isolated report evidence rows: %v", err)
		}
	}
}

func (f *storeReportEvidenceFixture) seedBase(t *testing.T) {
	t.Helper()
	f.exec(t, `insert into users (id,email,display_name,created_at,updated_at) values ($1,$2,'Report Evidence',$3,$3)`, f.userID, "report-evidence-"+f.prefix+"@example.test", f.now)
	f.exec(t, `
insert into resumes (
  id,user_id,title,display_name,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Evidence Resume','Frozen resume','en','ready','{}'::jsonb,'resume source','paste','resume source','resume source','{}'::jsonb,$3,$3)`, f.resumeID, f.userID, f.now)
	f.exec(t, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,company_name,target_language,source_type,
  raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Frozen Platform Engineer','Frozen Company','en','manual_text','frozen jd',$4::jsonb,'{}'::jsonb,$5,$5)`,
		f.targetID, f.userID, f.resumeID, storeReportTargetSummaryJSON(), f.now)
	f.exec(t, `
insert into practice_plans (
  id,user_id,target_job_id,goal,round_id,round_sequence,interviewer_persona,difficulty,language,
  time_budget_minutes,resume_id,focus_dimension_codes,status,created_at,updated_at
) values ($1,$2,$3,'baseline','round-1-technical',1,'technical_manager','standard','en',45,$4,'{}'::text[],'ready',$5,$5)`,
		f.planID, f.userID, f.targetID, f.resumeID, f.now)
}

type storeReportSnapshotOptions struct {
	MessageCount     int32
	LastMessageSeqNo int32
	RawJDBytes       int
}

func (f *storeReportEvidenceFixture) seedGeneratingReport(t *testing.T, offset int, options storeReportSnapshotOptions) (sessionID, reportID, jobID string) {
	t.Helper()
	sessionID = f.id(offset)
	reportID = f.id(offset + 1)
	jobID = f.id(offset + 2)
	messageCount := options.MessageCount
	if messageCount == 0 {
		messageCount = 3
	}
	lastSeq := options.LastMessageSeqNo
	if lastSeq == 0 {
		lastSeq = 3
	}
	rawJD := "frozen jd"
	if options.RawJDBytes > 0 {
		rawJD = strings.Repeat("x", options.RawJDBytes)
	}
	snapshot, err := practicedomain.BuildReportContextSnapshot(practicedomain.ReportContextSnapshotInput{
		TargetJob: practicedomain.ReportTargetJobSnapshot{
			ID: f.targetID, Title: "Frozen Platform Engineer", Company: "Frozen Company", Language: "en", RawJD: rawJD,
			Summary: json.RawMessage(storeReportTargetSummaryJSON()), Requirements: []practicedomain.ReportRequirementSnapshot{},
		},
		Resume: practicedomain.ReportResumeSnapshot{ID: f.resumeID, DisplayName: "Frozen resume", Language: "en", SourceSnapshot: "resume source", StructuredProfile: json.RawMessage(`{}`)},
		Plan: practicedomain.ReportPlanSnapshot{
			ID: f.planID, Goal: "baseline", InterviewerPersona: "technical_manager", Difficulty: "standard", Language: "en",
			TimeBudgetMinutes: 45, ResumeID: f.resumeID, RoundID: "round-1-technical", RoundSequence: 1, FocusDimensionCodes: []string{},
		},
		Conversation: practicedomain.ReportConversationCoordinate{SessionID: sessionID, Language: "en", MessageCount: messageCount, LastMessageSeqNo: lastSeq},
	})
	if err != nil {
		t.Fatalf("build report evidence snapshot: %v", err)
	}
	snapshotRaw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal report evidence snapshot: %v", err)
	}
	assistant1 := f.id(offset + 3)
	user2 := f.id(offset + 4)
	assistant3 := f.id(offset + 5)
	clientMessageID := f.id(offset + 6)
	f.exec(t, `insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,completed_at,created_at,updated_at) values ($1,$2,$3,$4,'completed','en',$5,$5,$5)`, sessionID, f.userID, f.planID, f.targetID, f.now)
	f.exec(t, `insert into practice_messages (id,session_id,seq_no,role,content,created_at) values ($1,$2,1,'assistant','Describe the migration.',$3)`, assistant1, sessionID, f.now)
	f.exec(t, `insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,created_at) values ($1,$2,2,'user','I used queue backpressure and monitored saturation.',$3,$4)`, user2, sessionID, clientMessageID, f.now)
	f.exec(t, `insert into practice_messages (id,session_id,seq_no,role,content,reply_to_message_id,created_at) values ($1,$2,3,'assistant','What was the rollback plan?',$3,$4)`, assistant3, sessionID, user2, f.now)
	f.exec(t, `insert into feedback_reports (id,user_id,session_id,target_job_id,status,generation_context,created_at,updated_at) values ($1,$2,$3,$4,'generating',$5::jsonb,$6,$6)`, reportID, f.userID, sessionID, f.targetID, snapshotRaw, f.now)
	f.exec(t, `insert into async_jobs (id,job_type,resource_type,resource_id,status,attempts,max_attempts,payload,available_at,locked_at,created_at,updated_at) values ($1,'report_generate','feedback_report',$2,'running',1,4,'{}'::jsonb,$3,$3,$3,$3)`, jobID, reportID, f.now)
	return sessionID, reportID, jobID
}

func (f *storeReportEvidenceFixture) reapAndTakeOver(t *testing.T, store *runner.SQLStore, jobID string) runner.ClaimedJob {
	t.Helper()
	if reclaimed, err := store.ReclaimExpiredLeases(f.ctx, []string{string(sharedjobs.JobTypeReportGenerate)}, f.now.Add(time.Second), f.now.Add(2*time.Second)); err != nil || reclaimed != 1 {
		t.Fatalf("reap attempt1 job=%s reclaimed=%d err=%v", jobID, reclaimed, err)
	}
	claimed, ok, err := store.LeaseAsyncJob(f.ctx, []string{string(sharedjobs.JobTypeReportGenerate)}, f.now.Add(3*time.Second))
	if err != nil || !ok || claimed.JobID != jobID || claimed.Attempts != 2 {
		t.Fatalf("take over attempt2 job=%s claimed=%+v ok=%t err=%v", jobID, claimed, ok, err)
	}
	return claimed
}

func (f *storeReportEvidenceFixture) assertLeaseFenceState(t *testing.T, reportID, jobID, outboxID, auditID, wantReportStatus, wantJobStatus string, wantAttempts int32, wantDomainEffects int) {
	t.Helper()
	var reportStatus, jobStatus string
	var attempts int32
	if err := f.db.QueryRowContext(f.ctx, `select fr.status,aj.status,aj.attempts from feedback_reports fr join async_jobs aj on aj.resource_id=fr.id where fr.id=$1 and aj.id=$2`, reportID, jobID).Scan(&reportStatus, &jobStatus, &attempts); err != nil {
		t.Fatalf("read lease fence state: %v", err)
	}
	if reportStatus != wantReportStatus || jobStatus != wantJobStatus || attempts != wantAttempts {
		t.Fatalf("lease fence state report=%s job=%s attempts=%d want report=%s job=%s attempts=%d", reportStatus, jobStatus, attempts, wantReportStatus, wantJobStatus, wantAttempts)
	}
	var effects int
	if err := f.db.QueryRowContext(f.ctx, `select (select count(*) from outbox_events where id=$1) + (select count(*) from audit_events where id=$2)`, outboxID, auditID).Scan(&effects); err != nil {
		t.Fatalf("count lease fence effects: %v", err)
	}
	if effects != wantDomainEffects {
		t.Fatalf("domain effects=%d want=%d", effects, wantDomainEffects)
	}
}

func (f *storeReportEvidenceFixture) service(repository *Repository, ai *storeReportEvidenceAI, idOffset int) *reviewdomain.Service {
	return f.serviceWithWaits(repository, ai, idOffset, nil)
}

func (f *storeReportEvidenceFixture) serviceWithWaits(repository *Repository, ai *storeReportEvidenceAI, idOffset int, waits *[]time.Duration) *reviewdomain.Service {
	index := idOffset
	return reviewdomain.NewService(reviewdomain.ServiceOptions{
		Registry: storeReportEvidenceResolver{}, AI: ai, Repository: repository,
		WaitBeforeRetry: func(_ context.Context, delay time.Duration) error {
			if waits != nil {
				*waits = append(*waits, delay)
			}
			return nil
		},
		Now: func() time.Time { return f.now.Add(5 * time.Second) },
		NewID: func() string {
			index++
			return f.id(index)
		},
	})
}

func (f *storeReportEvidenceFixture) assertFailedReadyColumnsEmpty(t *testing.T, reportID, wantErrorCode string) {
	t.Helper()
	var status, errorCode string
	var empty bool
	err := f.db.QueryRowContext(f.ctx, `
select status,error_code,
       summary is null
       and preparedness_level is null
       and dimension_assessments = '[]'::jsonb
       and highlights = '[]'::jsonb
       and issues = '[]'::jsonb
       and next_actions = '[]'::jsonb
       and retry_focus_dimension_codes = '{}'::text[]
       and prompt_version is null
       and rubric_version is null
       and model_id is null
from feedback_reports where id=$1`, reportID).Scan(&status, &errorCode, &empty)
	if err != nil {
		t.Fatalf("read failed report state: %v", err)
	}
	if status != "failed" || errorCode != wantErrorCode || !empty {
		t.Fatalf("failed report retained ready state: status=%s errorCode=%s empty=%t", status, errorCode, empty)
	}
}

func (f *storeReportEvidenceFixture) exec(t *testing.T, statement string, args ...any) {
	t.Helper()
	if _, err := f.db.ExecContext(f.ctx, statement, args...); err != nil {
		t.Fatalf("seed isolated report evidence row: %v", err)
	}
}

type storeReportEvidenceResolver struct{}

func (storeReportEvidenceResolver) ResolveActive(context.Context, string, string) (registry.PromptResolution, error) {
	schema := json.RawMessage(`{"type":"object"}`)
	return registry.PromptResolution{
		FeatureKey: "report.generate", PromptVersion: "v0.2.0", RubricVersion: "v0.2.0",
		ModelProfileName: "report.generate.default", FeatureFlag: "none", DataSourceVersion: practicedomain.ReportContextSchemaVersion,
		OutputSchema:        &schema,
		UserMessageTemplate: "Trusted policy {{language}}.\n<untrusted_report_context_json>\n{\"context\":{{frozen_context}},\"messages\":{{conversation_messages}}}\n</untrusted_report_context_json>\nGrounding rules after data.",
	}, nil
}

type storeReportEvidenceAIResult struct {
	content string
	err     error
}

type storeReportEvidenceAI struct {
	results []storeReportEvidenceAIResult
	calls   int
}

func (a *storeReportEvidenceAI) Complete(_ context.Context, _ string, _ aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	a.calls++
	if a.calls > len(a.results) {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("unexpected report evidence provider call")
	}
	result := a.results[a.calls-1]
	meta := aiclient.AICallMeta{
		Provider: "evidence-provider", ModelID: "evidence-model", PromptVersion: "v0.2.0", RubricVersion: "v0.2.0",
		ModelProfileName: "report.generate.default", Language: "en", FeatureKey: "report.generate", FeatureFlag: "none",
		DataSourceVersion: practicedomain.ReportContextSchemaVersion, ValidationStatus: aiclient.ValidationStatusOK,
		InputTokens: 32, OutputTokens: 16,
	}
	return aiclient.CompleteResponse{Content: result.content, FinishReason: "stop"}, meta, result.err
}

func storeValidDirectReportContent() reviewdomain.ReportContentDraft {
	return reviewdomain.ReportContentDraft{
		Summary:                  "The answer explained tradeoffs, but the rollback plan needs concrete steps.",
		PreparednessLevel:        sharedtypes.ReadinessTierNeedsPractice,
		DimensionAssessments:     []reviewdomain.DimensionAssessmentDraft{{Code: "technical_depth", Label: "Technical depth", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh}},
		Highlights:               []reviewdomain.ReportEvidenceDraft{},
		Issues:                   []reviewdomain.ReportEvidenceDraft{{DimensionCode: "technical_depth", Evidence: "The candidate explained backpressure but did not provide rollback steps.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}}},
		NextActions:              []reviewdomain.ReportNextActionDraft{{Type: "retry_current_round", Label: "Add rollback steps and replay this round"}},
		RetryFocusDimensionCodes: []string{"technical_depth"},
	}
}

func storeValidDirectReportJSON() string {
	raw, _ := json.Marshal(storeValidDirectReportContent())
	return string(raw)
}

func storeReportTargetSummaryJSON() string {
	return `{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"system design"},{"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`
}

func storeReportEvidenceUUID(prefix string, index int) string {
	return fmt.Sprintf("019f5800-%s-7000-8000-%012d", prefix, index)
}

type storeCompletionOwnerEvidence struct {
	SchemaVersion string                             `json:"schemaVersion"`
	ScenarioID    string                             `json:"scenarioId"`
	Command       string                             `json:"command"`
	Tests         []storeCompletionOwnerEvidenceTest `json:"tests"`
	Markers       []string                           `json:"markers"`
	Database      storeCompletionOwnerEvidenceDB     `json:"database"`
	Result        string                             `json:"result"`
}

type storeCompletionOwnerEvidenceTest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type storeCompletionOwnerEvidenceDB struct {
	ZeroAnswerSideEffectCount   int    `json:"zeroAnswerSideEffectCount"`
	PendingReplySideEffectCount int    `json:"pendingReplySideEffectCount"`
	SnapshotSchemaVersion       string `json:"snapshotSchemaVersion"`
	ConcurrentMutationBlocked   bool   `json:"concurrentMutationBlocked"`
	SnapshotReplayEqual         bool   `json:"snapshotReplayEqual"`
	MismatchSideEffectCount     int    `json:"mismatchSideEffectCount"`
}

func requireStoreCompletionOwnerEvidence(t *testing.T) {
	t.Helper()
	path := strings.TrimSpace(os.Getenv("PRACTICE_COMPLETION_EVIDENCE_PATH"))
	if path == "" {
		t.Skip("PRACTICE_COMPLETION_EVIDENCE_PATH is not set; scenario owner evidence is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read practice completion owner evidence: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var evidence storeCompletionOwnerEvidence
	if err := decoder.Decode(&evidence); err != nil {
		t.Fatalf("decode practice completion owner evidence: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatal("practice completion owner evidence has trailing JSON")
	}
	const expectedCommand = "cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^(TestE2EP0047RejectsZeroAnswerCompletion|TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' -count=1 -v"
	if evidence.SchemaVersion != "practice-completion-evidence.v1" || evidence.ScenarioID != "E2E.P0.047" || evidence.Command != expectedCommand || evidence.Result != "PASS" {
		t.Fatal("practice completion owner evidence identity/result mismatch")
	}
	wantTests := []storeCompletionOwnerEvidenceTest{
		{Name: "TestE2EP0047RejectsZeroAnswerCompletion", Status: "PASS"},
		{Name: "TestE2EP0047FreezesReportContext", Status: "PASS"},
		{Name: "TestE2EP0047CompletionReplayPreservesReportContext", Status: "PASS"},
	}
	if !reflect.DeepEqual(evidence.Tests, wantTests) {
		t.Fatal("practice completion owner evidence test set mismatch")
	}
	wantMarkers := []string{"REPORT_CONTEXT_REPLAY_PASS", "REPORT_CONTEXT_SNAPSHOT_PASS", "ZERO_ANSWER_COMPLETION_REJECTED_PASS"}
	gotMarkers := append([]string(nil), evidence.Markers...)
	sort.Strings(gotMarkers)
	if !reflect.DeepEqual(gotMarkers, wantMarkers) {
		t.Fatal("practice completion owner evidence marker set mismatch")
	}
	if evidence.Database.ZeroAnswerSideEffectCount != 0 || evidence.Database.PendingReplySideEffectCount != 0 ||
		evidence.Database.SnapshotSchemaVersion != "report-context.v1" || !evidence.Database.ConcurrentMutationBlocked ||
		!evidence.Database.SnapshotReplayEqual || evidence.Database.MismatchSideEffectCount != 0 {
		t.Fatal("practice completion owner evidence database contract mismatch")
	}
}
