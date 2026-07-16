package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestRegenerateFeedbackReportAtomicallyResetsAndCreatesJobWithAudit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	in := reviewdomain.RegenerateReportStoreInput{
		UserID: "user-1", ReportID: "report-1", JobID: "job-new", AuditEventID: "audit-new", Now: now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).WithArgs("report-regenerate|report-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)from async_jobs aj.*join feedback_reports fr.*status in \('queued', 'running'\).*for update of aj`).
		WithArgs(in.ReportID, in.UserID, string(sharedjobs.JobTypeReportGenerate)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`(?s)from feedback_reports.*where id = \$1 and user_id = \$2.*for update`).
		WithArgs(in.ReportID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "session_id", "target_job_id", "status", "error_code"}).
			AddRow(in.ReportID, "session-1", "target-1", "failed", sharederrors.CodeAiOutputInvalid))
	mock.ExpectExec(`(?s)update feedback_reports.*status = 'queued'.*summary = null.*provider = null.*error_code = null.*where id = \$2`).
		WithArgs(now, in.ReportID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)insert into async_jobs.*'feedback_report'.*'queued',0`).
		WithArgs(in.JobID, string(sharedjobs.JobTypeReportGenerate), in.ReportID, "session-1", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)insert into audit_events.*feedback_report\.regeneration_requested`).
		WithArgs(in.AuditEventID, in.UserID, in.ReportID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := NewRepository(db).RegenerateFeedbackReport(context.Background(), in)
	if err != nil {
		t.Fatalf("RegenerateFeedbackReport: %v", err)
	}
	if result.ReportID != in.ReportID || result.Job.ID != in.JobID || result.Job.ResourceID != in.ReportID || result.Job.Status != "queued" || !result.Job.CreatedAt.Equal(now) {
		t.Fatalf("result=%+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRegenerateFeedbackReportReturnsTypedRejectionsBeforeWrites(t *testing.T) {
	for _, tc := range []struct {
		name       string
		activeJob  bool
		reportRows *sqlmock.Rows
		want       error
	}{
		{name: "active job", activeJob: true, want: reviewdomain.ErrReportNotReady},
		{name: "hidden report", reportRows: sqlmock.NewRows([]string{"id", "session_id", "target_job_id", "status", "error_code"}), want: reviewdomain.ErrReportNotFound},
		{name: "ready report", reportRows: sqlmock.NewRows([]string{"id", "session_id", "target_job_id", "status", "error_code"}).AddRow("report-1", "session-1", "target-1", "ready", nil), want: reviewdomain.ErrReportInvalidStateTransition},
		{name: "oversized report", reportRows: sqlmock.NewRows([]string{"id", "session_id", "target_job_id", "status", "error_code"}).AddRow("report-1", "session-1", "target-1", "failed", sharederrors.CodeReportContextTooLarge), want: reviewdomain.ErrReportContextTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			in := reviewdomain.RegenerateReportStoreInput{UserID: "user-1", ReportID: "report-1", JobID: "job-new", AuditEventID: "audit-new", Now: time.Unix(1, 0).UTC()}
			mock.ExpectBegin()
			mock.ExpectExec(`select pg_advisory_xact_lock`).WithArgs("report-regenerate|report-1").WillReturnResult(sqlmock.NewResult(0, 1))
			activeRows := sqlmock.NewRows([]string{"id"})
			if tc.activeJob {
				activeRows.AddRow("job-old")
			}
			mock.ExpectQuery(`(?s)from async_jobs aj.*for update of aj`).WithArgs(in.ReportID, in.UserID, string(sharedjobs.JobTypeReportGenerate)).WillReturnRows(activeRows)
			if !tc.activeJob {
				mock.ExpectQuery(`(?s)from feedback_reports.*for update`).WithArgs(in.ReportID, in.UserID).WillReturnRows(tc.reportRows)
			}
			mock.ExpectRollback()

			result, err := NewRepository(db).RegenerateFeedbackReport(context.Background(), in)
			if !errors.Is(err, tc.want) || result.ReportID != "" || result.Job.ID != "" {
				t.Fatalf("result=%+v err=%v want=%v", result, err, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("rejection reached a write: %v", err)
			}
		})
	}
}

func TestRegenerateFeedbackReportRollsBackResetWhenFreshJobInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	in := reviewdomain.RegenerateReportStoreInput{UserID: "user-1", ReportID: "report-1", JobID: "job-new", AuditEventID: "audit-new", Now: time.Unix(1, 0).UTC()}
	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)from async_jobs aj.*for update of aj`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`(?s)from feedback_reports.*for update`).WillReturnRows(sqlmock.NewRows([]string{"id", "session_id", "target_job_id", "status", "error_code"}).AddRow("report-1", "session-1", "target-1", "failed", sharederrors.CodeAiOutputInvalid))
	mock.ExpectExec(`(?s)update feedback_reports.*status = 'queued'`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	if _, err := NewRepository(db).RegenerateFeedbackReport(context.Background(), in); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("error=%v want wrapped sql.ErrConnDone", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("rollback contract: %v", err)
	}
}

func TestRegenerateReportStoreSourceLocksJobBeforeReportAndHasAuditOnlySideEffects(t *testing.T) {
	raw, err := os.ReadFile("regenerate_report.go")
	if err != nil {
		t.Fatalf("read regenerate_report.go: %v", err)
	}
	source := strings.ToLower(string(raw))
	jobLock := strings.Index(source, "from async_jobs")
	reportLock := strings.Index(source, "from feedback_reports")
	if jobLock < 0 || reportLock < 0 || jobLock >= reportLock {
		t.Fatalf("regeneration transaction must lock active report job before owned feedback report: job=%d report=%d", jobLock, reportLock)
	}
	for _, required := range []string{
		"for update",
		"status in ('queued', 'running')",
		"update feedback_reports",
		"status = 'queued'",
		"summary = null",
		"preparedness_level = null",
		"dimension_assessments = '[]'::jsonb",
		"highlights = '[]'::jsonb",
		"issues = '[]'::jsonb",
		"next_actions = '[]'::jsonb",
		"retry_focus_dimension_codes = '{}'::text[]",
		"prompt_version = null",
		"rubric_version = null",
		"model_id = null",
		"provider = null",
		"feature_flag = 'none'",
		"data_source_version = 'not_applicable'",
		"error_code = null",
		"generated_at = null",
		"insert into async_jobs",
		"job_type = $3",
		"'feedback_report'",
		"insert into audit_events",
		"feedback_report.regeneration_requested",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("regeneration transaction missing %q", required)
		}
	}
	for _, forbidden := range []string{"insert into outbox_events", "delete from", "practice_messages", "raw_prompt", "raw_response"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("regeneration transaction must not contain %q", forbidden)
		}
	}
}

func TestRegenerateReportStoreDifferentKeysCreateAtMostOneActiveJobAndPreserveFrozenInput(t *testing.T) {
	fixture := newStoreReportEvidenceFixture(t, "0a61")
	fixture.seedBase(t)
	sessionID, reportID, oldJobID := fixture.seedGeneratingReport(t, 210, storeReportSnapshotOptions{})
	repository := NewRepository(fixture.db)
	method := reflect.ValueOf(repository).MethodByName("RegenerateFeedbackReport")
	if !method.IsValid() {
		t.Fatal("review Repository must expose RegenerateFeedbackReport")
	}
	if method.Type().NumIn() != 2 || method.Type().NumOut() != 2 {
		t.Fatalf("RegenerateFeedbackReport signature=%s, want (context.Context, RegenerateReportStoreInput) (RegenerateReportStoreResult, error)", method.Type())
	}

	fixture.exec(t, `
update feedback_reports set
  status='failed', summary='stale summary', preparedness_level='well_prepared',
  dimension_assessments='[{"code":"stale"}]'::jsonb,
  highlights='[{"evidence":"stale"}]'::jsonb,
  issues='[{"evidence":"stale"}]'::jsonb,
  next_actions='[{"label":"stale"}]'::jsonb,
  retry_focus_dimension_codes=array['stale'],
  prompt_version='stale-prompt', rubric_version='stale-rubric', model_id='stale-model', provider='stale-provider',
  feature_flag='stale-flag', data_source_version='stale-data', error_code=$2, generated_at=$3, updated_at=$3
where id=$1`, reportID, sharederrors.CodeAiOutputInvalid, fixture.now.Add(time.Minute))
	fixture.exec(t, `update async_jobs set status='failed', locked_at=null, completed_at=$2, error_code=$3, updated_at=$2 where id=$1`, oldJobID, fixture.now.Add(time.Minute), sharederrors.CodeAiOutputInvalid)

	var frozenBefore []byte
	var messageCountBefore int
	if err := fixture.db.QueryRowContext(fixture.ctx, `select generation_context from feedback_reports where id=$1`, reportID).Scan(&frozenBefore); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from practice_messages where session_id=$1`, sessionID).Scan(&messageCountBefore); err != nil {
		t.Fatal(err)
	}

	type callResult struct{ err error }
	results := make(chan callResult, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			input, inputErr := makeRegenerateStoreInput(method.Type().In(1), map[string]any{
				"UserID": fixture.userID, "ReportID": reportID,
				"JobID": fixture.id(230 + index), "AuditEventID": fixture.id(240 + index),
				"Now": fixture.now.Add(2 * time.Minute),
			})
			if inputErr != nil {
				results <- callResult{err: inputErr}
				return
			}
			outputs := method.Call([]reflect.Value{reflect.ValueOf(fixture.ctx), input})
			results <- callResult{err: reflectedError(outputs[1])}
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	successes := 0
	for result := range results {
		if result.err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent regeneration successes=%d, want exactly one", successes)
	}

	var (
		status, boundSessionID, boundTargetID string
		frozenAfter                           []byte
		readyFieldsEmpty                      bool
		messageCountAfter                     int
	)
	if err := fixture.db.QueryRowContext(fixture.ctx, `
select status, session_id::text, target_job_id::text, generation_context,
       summary is null and preparedness_level is null
       and dimension_assessments = '[]'::jsonb and highlights = '[]'::jsonb
       and issues = '[]'::jsonb and next_actions = '[]'::jsonb
       and retry_focus_dimension_codes = '{}'::text[]
       and prompt_version is null and rubric_version is null and model_id is null and provider is null
       and feature_flag = 'none' and data_source_version = 'not_applicable'
       and error_code is null and generated_at is null
from feedback_reports where id=$1`, reportID).Scan(&status, &boundSessionID, &boundTargetID, &frozenAfter, &readyFieldsEmpty); err != nil {
		t.Fatal(err)
	}
	if status != "queued" || boundSessionID != sessionID || boundTargetID != fixture.targetID || !jsonEqual(frozenBefore, frozenAfter) || !readyFieldsEmpty {
		t.Fatalf("reset drift: status=%s session=%s target=%s frozenEqual=%t readyEmpty=%t", status, boundSessionID, boundTargetID, jsonEqual(frozenBefore, frozenAfter), readyFieldsEmpty)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from practice_messages where session_id=$1`, sessionID).Scan(&messageCountAfter); err != nil {
		t.Fatal(err)
	}
	if messageCountAfter != messageCountBefore {
		t.Fatalf("transcript count=%d, want preserved %d", messageCountAfter, messageCountBefore)
	}

	var activeJobs, allJobs, auditCount, outboxCount int
	var newJobID, dedupeKey, jobStatus string
	var attempts int
	var payload, auditMetadata []byte
	if err := fixture.db.QueryRowContext(fixture.ctx, `
select count(*) filter (where status in ('queued','running')),
       count(*), max(id::text) filter (where id <> $2)
from async_jobs where resource_id=$1 and job_type='report_generate'`, reportID, oldJobID).Scan(&activeJobs, &allJobs, &newJobID); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select dedupe_key,status,attempts,payload from async_jobs where id=$1`, newJobID).Scan(&dedupeKey, &jobStatus, &attempts, &payload); err != nil {
		t.Fatal(err)
	}
	var jobPayload map[string]any
	if err := json.Unmarshal(payload, &jobPayload); err != nil {
		t.Fatal(err)
	}
	if activeJobs != 1 || allJobs != 2 || dedupeKey != sessionID || jobStatus != "queued" || attempts != 0 || jobPayload["reportId"] != reportID || jobPayload["sessionId"] != sessionID || jobPayload["targetJobId"] != fixture.targetID {
		t.Fatalf("job drift: active=%d total=%d dedupe=%s status=%s attempts=%d payload=%v", activeJobs, allJobs, dedupeKey, jobStatus, attempts, jobPayload)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from audit_events where resource_id=$1 and action='feedback_report.regeneration_requested'`, reportID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select metadata from audit_events where resource_id=$1 and action='feedback_report.regeneration_requested'`, reportID).Scan(&auditMetadata); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from outbox_events where aggregate_id=$1`, reportID).Scan(&outboxCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 1 || outboxCount != 0 {
		t.Fatalf("side effects audit=%d outbox=%d", auditCount, outboxCount)
	}
	var metadata map[string]any
	if err := json.Unmarshal(auditMetadata, &metadata); err != nil {
		t.Fatal(err)
	}
	if len(metadata) != 2 || metadata["jobId"] != newJobID || metadata["previousErrorCode"] != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("audit metadata=%v, want only fresh jobId and previousErrorCode", metadata)
	}
}

func TestRegenerateReportStoreRejectsInvalidStateOversizeActiveJobAndCrossUserWithoutWrites(t *testing.T) {
	for _, tc := range []struct {
		name        string
		prefix      string
		prepare     func(*testing.T, *storeReportEvidenceFixture, string, string)
		requestUser func(*storeReportEvidenceFixture) string
	}{
		{
			name: "ready report", prefix: "0a62",
			prepare: func(t *testing.T, f *storeReportEvidenceFixture, reportID, jobID string) {
				f.exec(t, `update feedback_reports set status='ready',error_code=null,generated_at=$2 where id=$1`, reportID, f.now.Add(time.Minute))
				f.exec(t, `update async_jobs set status='succeeded',locked_at=null,completed_at=$2 where id=$1`, jobID, f.now.Add(time.Minute))
			},
		},
		{
			name: "permanently oversized report", prefix: "0a63",
			prepare: func(t *testing.T, f *storeReportEvidenceFixture, reportID, jobID string) {
				f.exec(t, `update feedback_reports set status='failed',error_code=$2 where id=$1`, reportID, sharederrors.CodeReportContextTooLarge)
				f.exec(t, `update async_jobs set status='failed',locked_at=null,completed_at=$2 where id=$1`, jobID, f.now.Add(time.Minute))
			},
		},
		{
			name: "old generation job still active", prefix: "0a64",
			prepare: func(t *testing.T, f *storeReportEvidenceFixture, reportID, _ string) {
				f.exec(t, `update feedback_reports set status='failed',error_code=$2 where id=$1`, reportID, sharederrors.CodeAiOutputInvalid)
			},
		},
		{
			name: "cross user hidden report", prefix: "0a65",
			prepare: func(t *testing.T, f *storeReportEvidenceFixture, reportID, jobID string) {
				f.exec(t, `update feedback_reports set status='failed',error_code=$2 where id=$1`, reportID, sharederrors.CodeAiOutputInvalid)
				f.exec(t, `update async_jobs set status='failed',locked_at=null,completed_at=$2 where id=$1`, jobID, f.now.Add(time.Minute))
			},
			requestUser: func(f *storeReportEvidenceFixture) string { return f.id(900) },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newStoreReportEvidenceFixture(t, tc.prefix)
			fixture.seedBase(t)
			_, reportID, oldJobID := fixture.seedGeneratingReport(t, 310, storeReportSnapshotOptions{})
			tc.prepare(t, fixture, reportID, oldJobID)
			method := reflect.ValueOf(NewRepository(fixture.db)).MethodByName("RegenerateFeedbackReport")
			if !method.IsValid() {
				t.Fatal("review Repository must expose RegenerateFeedbackReport")
			}
			userID := fixture.userID
			if tc.requestUser != nil {
				userID = tc.requestUser(fixture)
			}
			assertRegenerationRejectedWithoutWrites(t, fixture, method, userID, reportID, fixture.id(330), fixture.id(331))
		})
	}
}

func assertRegenerationRejectedWithoutWrites(t *testing.T, fixture *storeReportEvidenceFixture, method reflect.Value, userID, reportID, jobID, auditID string) {
	t.Helper()
	var beforeJobs, beforeAudits, beforeOutbox int
	var beforeStatus, beforeErrorCode string
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from async_jobs where resource_id=$1`, reportID).Scan(&beforeJobs); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from audit_events where resource_id=$1`, reportID).Scan(&beforeAudits); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from outbox_events where aggregate_id=$1`, reportID).Scan(&beforeOutbox); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select status,coalesce(error_code,'') from feedback_reports where id=$1`, reportID).Scan(&beforeStatus, &beforeErrorCode); err != nil {
		t.Fatal(err)
	}
	input, err := makeRegenerateStoreInput(method.Type().In(1), map[string]any{
		"UserID": userID, "ReportID": reportID, "JobID": jobID, "AuditEventID": auditID,
		"Now": fixture.now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	outputs := method.Call([]reflect.Value{reflect.ValueOf(fixture.ctx), input})
	if callErr := reflectedError(outputs[1]); callErr == nil {
		t.Fatal("ineligible regeneration unexpectedly succeeded")
	}
	var afterJobs, afterAudits, afterOutbox, requestedEffects int
	var afterStatus, afterErrorCode string
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from async_jobs where resource_id=$1`, reportID).Scan(&afterJobs); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from audit_events where resource_id=$1`, reportID).Scan(&afterAudits); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from outbox_events where aggregate_id=$1`, reportID).Scan(&afterOutbox); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select status,coalesce(error_code,'') from feedback_reports where id=$1`, reportID).Scan(&afterStatus, &afterErrorCode); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `select (select count(*) from async_jobs where id=$1) + (select count(*) from audit_events where id=$2)`, jobID, auditID).Scan(&requestedEffects); err != nil {
		t.Fatal(err)
	}
	if beforeJobs != afterJobs || beforeAudits != afterAudits || beforeOutbox != afterOutbox || beforeStatus != afterStatus || beforeErrorCode != afterErrorCode || requestedEffects != 0 {
		t.Fatalf("rejection wrote state: jobs %d/%d audits %d/%d outbox %d/%d status %s/%s error %s/%s requested=%d", beforeJobs, afterJobs, beforeAudits, afterAudits, beforeOutbox, afterOutbox, beforeStatus, afterStatus, beforeErrorCode, afterErrorCode, requestedEffects)
	}
}

func makeRegenerateStoreInput(inputType reflect.Type, values map[string]any) (reflect.Value, error) {
	wantPointer := inputType.Kind() == reflect.Pointer
	structType := inputType
	if wantPointer {
		structType = inputType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("regeneration store input %s is not a struct", inputType)
	}
	input := reflect.New(structType).Elem()
	for name, value := range values {
		field := input.FieldByName(name)
		if !field.IsValid() || !field.CanSet() {
			return reflect.Value{}, fmt.Errorf("regeneration store input missing settable field %s", name)
		}
		candidate := reflect.ValueOf(value)
		if !candidate.Type().AssignableTo(field.Type()) {
			return reflect.Value{}, fmt.Errorf("regeneration store input field %s type=%s value=%s", name, field.Type(), candidate.Type())
		}
		field.Set(candidate)
	}
	if wantPointer {
		pointer := reflect.New(structType)
		pointer.Elem().Set(input)
		return pointer, nil
	}
	return input, nil
}

func reflectedError(value reflect.Value) error {
	if value.IsNil() {
		return nil
	}
	err, _ := value.Interface().(error)
	return err
}

func jsonEqual(left, right []byte) bool {
	var leftValue, rightValue any
	return json.Unmarshal(left, &leftValue) == nil && json.Unmarshal(right, &rightValue) == nil && reflect.DeepEqual(leftValue, rightValue)
}
