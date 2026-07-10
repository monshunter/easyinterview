package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *SQLRepository) CompleteSession(ctx context.Context, in domain.CompleteSessionStoreInput) (domain.CompleteSessionResult, error) {
	if r == nil || r.db == nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("begin complete practice session: %w", err)
	}
	defer tx.Rollback()

	session, err := selectSessionForCompletion(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.CompleteSessionResult{}, err
	}
	if existing, hit, err := selectExistingReportWithJob(ctx, tx, in.UserID, in.SessionID); err != nil {
		return domain.CompleteSessionResult{}, err
	} else if hit {
		if err := tx.Commit(); err != nil {
			return domain.CompleteSessionResult{}, fmt.Errorf("commit complete practice session replay: %w", err)
		}
		existing.Replay = true
		return existing, nil
	}
	if !canCompletePracticeSessionStatus(session.Status) {
		return domain.CompleteSessionResult{}, domain.ErrSessionConflict
	}

	if _, err := tx.ExecContext(ctx, `
update practice_sessions
set status = $1,
    completed_at = coalesce(completed_at, $2),
    updated_at = $2
where id = $3
  and user_id = $4`,
		string(sharedtypes.SessionStatusCompleting),
		in.Now.UTC(),
		in.SessionID,
		in.UserID,
	); err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("mark practice session completing: %w", err)
	}
	seqNo, err := nextSessionEventSeq(ctx, tx, in.SessionID)
	if err != nil {
		return domain.CompleteSessionResult{}, err
	}
	eventPayload, err := json.Marshal(map[string]any{
		"sessionId":         in.SessionID,
		"clientCompletedAt": in.ClientCompletedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("marshal session completed event payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into practice_session_events (
  id, session_id, seq_no, event_type, client_event_id, payload, created_at
) values ($1,$2,$3,'session_completed',null,$4,$5)`,
		in.SessionEventID,
		in.SessionID,
		seqNo,
		eventPayload,
		in.Now.UTC(),
	); err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("insert practice session completed event: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into feedback_reports (
  id, user_id, session_id, target_job_id, status, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$6)`,
		in.ReportID,
		in.UserID,
		in.SessionID,
		session.TargetJobID,
		string(sharedtypes.ReportStatusQueued),
		in.Now.UTC(),
	); err != nil {
		if isUniqueViolation(err) {
			existing, hit, selectErr := selectExistingReportWithJob(ctx, tx, in.UserID, in.SessionID)
			if selectErr != nil {
				return domain.CompleteSessionResult{}, selectErr
			}
			if hit {
				if err := tx.Commit(); err != nil {
					return domain.CompleteSessionResult{}, fmt.Errorf("commit complete practice session unique replay: %w", err)
				}
				existing.Replay = true
				return existing, nil
			}
		}
		return domain.CompleteSessionResult{}, fmt.Errorf("insert queued feedback report: %w", err)
	}
	jobPayload, err := json.Marshal(map[string]any{
		"reportId":    in.ReportID,
		"sessionId":   in.SessionID,
		"targetJobId": session.TargetJobID,
	})
	if err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("marshal report job payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status,
  payload, available_at, created_at, updated_at
) values ($1,$2,'feedback_report',$3,$4,$5,$6,$7,$7,$7)`,
		in.JobID,
		string(sharedjobs.JobTypeReportGenerate),
		in.ReportID,
		in.SessionID,
		string(sharedtypes.JobStatusQueued),
		jobPayload,
		in.Now.UTC(),
	); err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("insert report_generate async job: %w", err)
	}
	outboxPayload, err := BuildPracticeSessionCompletedPayload(PracticeSessionCompletedInput{
		Language:    session.Language,
		PlanID:      session.PlanID,
		SessionID:   session.ID,
		TargetJobID: session.TargetJobID,
		TurnCount:   int(session.TurnCount),
	})
	if err != nil {
		return domain.CompleteSessionResult{}, err
	}
	outboxRaw, err := json.Marshal(outboxPayload)
	if err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("marshal practice session completed payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1,$2,1,'practice_session',$3,$4,'pending',$5)`,
		in.OutboxEventID,
		string(sharedevents.EventNamePracticeSessionCompleted),
		in.SessionID,
		outboxRaw,
		in.Now.UTC(),
	); err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("insert practice session completed outbox event: %w", err)
	}
	auditMetadata, err := json.Marshal(map[string]any{
		"session_id":    in.SessionID,
		"report_id":     in.ReportID,
		"job_id":        in.JobID,
		"target_job_id": session.TargetJobID,
		"turn_count":    session.TurnCount,
	})
	if err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("marshal practice session complete audit metadata: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type,
  resource_id, result, metadata, created_at
) values ($1,$2,'user',$3,'practice.session.complete','practice_session',$4,'success',$5,$6)`,
		in.AuditEventID,
		in.UserID,
		in.UserID,
		in.SessionID,
		auditMetadata,
		in.Now.UTC(),
	); err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("insert practice session complete audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.CompleteSessionResult{}, fmt.Errorf("commit complete practice session: %w", err)
	}
	return domain.CompleteSessionResult{
		ReportID: in.ReportID,
		Job: domain.JobRecord{
			ID:           in.JobID,
			JobType:      api.JobTypeReportGenerate,
			ResourceType: api.ResourceTypeFeedbackReport,
			ResourceID:   in.ReportID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    in.Now.UTC(),
			UpdatedAt:    in.Now.UTC(),
		},
	}, nil
}

func selectSessionForCompletion(ctx context.Context, tx *sql.Tx, userID, sessionID string) (domain.SessionRecord, error) {
	var session domain.SessionRecord
	err := tx.QueryRowContext(ctx, `
select id, plan_id, target_job_id, status, language, hints_enabled,
       turn_count, created_at, updated_at
from practice_sessions
where user_id = $1
  and id = $2
for update`,
		userID,
		sessionID,
	).Scan(
		&session.ID,
		&session.PlanID,
		&session.TargetJobID,
		&session.Status,
		&session.Language,
		&session.HintsEnabled,
		&session.TurnCount,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SessionRecord{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("select practice session for completion: %w", err)
	}
	return session, nil
}

func canCompletePracticeSessionStatus(status sharedtypes.SessionStatus) bool {
	switch status {
	case sharedtypes.SessionStatusRunning, sharedtypes.SessionStatusWaitingUserInput, sharedtypes.SessionStatusCompleted:
		return true
	default:
		return false
	}
}

func selectExistingReportWithJob(ctx context.Context, tx *sql.Tx, userID, sessionID string) (domain.CompleteSessionResult, bool, error) {
	var result domain.CompleteSessionResult
	var errorCode sql.NullString
	err := tx.QueryRowContext(ctx, `
select fr.id,
       j.id, j.job_type, j.resource_type, j.resource_id, j.status, j.error_code,
       j.created_at, j.updated_at
from feedback_reports fr
join async_jobs j on j.resource_id = fr.id
where fr.user_id = $1
  and fr.session_id = $2
  and j.job_type = $3
  and j.resource_type = 'feedback_report'
  and j.dedupe_key = $4
order by j.created_at asc
limit 1`,
		userID,
		sessionID,
		string(sharedjobs.JobTypeReportGenerate),
		sessionID,
	).Scan(
		&result.ReportID,
		&result.Job.ID,
		&result.Job.JobType,
		&result.Job.ResourceType,
		&result.Job.ResourceID,
		&result.Job.Status,
		&errorCode,
		&result.Job.CreatedAt,
		&result.Job.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.CompleteSessionResult{}, false, nil
	}
	if err != nil {
		return domain.CompleteSessionResult{}, false, fmt.Errorf("select existing practice report job: %w", err)
	}
	result.Job.ErrorCode = errorCode.String
	return result, true, nil
}
