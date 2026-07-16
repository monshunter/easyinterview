package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *Repository) RegenerateFeedbackReport(ctx context.Context, in reviewdomain.RegenerateReportStoreInput) (reviewdomain.RegenerateReportStoreResult, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, err
	}
	if strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.ReportID) == "" || strings.TrimSpace(in.JobID) == "" || strings.TrimSpace(in.AuditEventID) == "" || in.Now.IsZero() {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("regenerate feedback report requires userId, reportId, jobId, auditEventId and now")
	}
	now := in.Now.UTC()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("begin regenerate feedback report: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1))`, "report-regenerate|"+in.ReportID); err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("lock feedback report regeneration: %w", err)
	}
	var activeJobID string
	err = tx.QueryRowContext(ctx, `
select aj.id::text
from async_jobs aj
join feedback_reports fr on fr.id = aj.resource_id and fr.user_id = $2
where aj.resource_id = $1
  and aj.job_type = $3
  and aj.resource_type = 'feedback_report'
  and aj.status in ('queued', 'running')
order by aj.created_at desc, aj.id desc
limit 1
for update of aj`, in.ReportID, in.UserID, string(sharedjobs.JobTypeReportGenerate)).Scan(&activeJobID)
	if err == nil {
		return reviewdomain.RegenerateReportStoreResult{}, reviewdomain.ErrReportNotReady
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("lock active report generation job: %w", err)
	}

	var (
		reportID, sessionID, targetJobID, status string
		previousErrorCode                        sql.NullString
	)
	err = tx.QueryRowContext(ctx, `
select id::text, session_id::text, target_job_id::text, status, error_code
from feedback_reports
where id = $1 and user_id = $2
for update`, in.ReportID, in.UserID).Scan(&reportID, &sessionID, &targetJobID, &status, &previousErrorCode)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.RegenerateReportStoreResult{}, reviewdomain.ErrReportNotFound
	}
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("lock owned feedback report: %w", err)
	}
	if status != string(sharedtypes.ReportStatusFailed) {
		return reviewdomain.RegenerateReportStoreResult{}, reviewdomain.ErrReportInvalidStateTransition
	}
	if previousErrorCode.Valid && previousErrorCode.String == sharederrors.CodeReportContextTooLarge {
		return reviewdomain.RegenerateReportStoreResult{}, reviewdomain.ErrReportContextTooLarge
	}

	res, err := tx.ExecContext(ctx, `
update feedback_reports
set status = 'queued',
    summary = null,
    preparedness_level = null,
    dimension_assessments = '[]'::jsonb,
    highlights = '[]'::jsonb,
    issues = '[]'::jsonb,
    next_actions = '[]'::jsonb,
    retry_focus_dimension_codes = '{}'::text[],
    prompt_version = null,
    rubric_version = null,
    model_id = null,
    provider = null,
    feature_flag = 'none',
    data_source_version = 'not_applicable',
    error_code = null,
    generated_at = null,
    updated_at = $1
where id = $2 and user_id = $3 and status = 'failed'`, now, reportID, in.UserID)
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("reset failed feedback report: %w", err)
	}
	if err := requireOneRow(res, "reset failed feedback report"); err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, err
	}

	jobPayload, err := json.Marshal(map[string]string{
		"reportId": reportID, "sessionId": sessionID, "targetJobId": targetJobID,
	})
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("marshal regenerated report job payload: %w", err)
	}
	res, err = tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, attempts,
  payload, available_at, created_at, updated_at
) values ($1,$2,'feedback_report',$3,$4,'queued',0,$5,$6,$6,$6)`,
		in.JobID, string(sharedjobs.JobTypeReportGenerate), reportID, sessionID, jobPayload, now)
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("insert regenerated report job: %w", err)
	}
	if err := requireOneRow(res, "insert regenerated report job"); err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, err
	}

	auditMetadata, err := json.Marshal(map[string]string{
		"jobId": in.JobID, "previousErrorCode": previousErrorCode.String,
	})
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("marshal report regeneration audit metadata: %w", err)
	}
	res, err = tx.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type, resource_id, result, metadata, created_at
) values ($1,$2,'user',$2,'feedback_report.regeneration_requested','feedback_report',$3,'success',$4,$5)`,
		in.AuditEventID, in.UserID, reportID, auditMetadata, now)
	if err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("insert report regeneration audit event: %w", err)
	}
	if err := requireOneRow(res, "insert report regeneration audit event"); err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return reviewdomain.RegenerateReportStoreResult{}, fmt.Errorf("commit regenerate feedback report: %w", err)
	}

	return reviewdomain.RegenerateReportStoreResult{
		ReportID: reportID,
		Job: reviewdomain.ReportJobRecord{
			ID: in.JobID, JobType: string(sharedjobs.JobTypeReportGenerate), ResourceType: "feedback_report",
			ResourceID: reportID, Status: sharedtypes.JobStatusQueued, CreatedAt: now, UpdatedAt: now,
		},
	}, nil
}
