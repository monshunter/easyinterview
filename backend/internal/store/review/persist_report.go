package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
)

type PersistReportInput = reviewdomain.ReportResultPersistence

func (r *Repository) PersistReport(ctx context.Context, in PersistReportInput) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	if err := validateReportPersistenceProvenance(in); err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin persist report: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCurrentAsyncJobLease(ctx, tx, in.AsyncJobID, in.ClaimedAttempts); err != nil {
		return err
	}

	highlights, err := json.Marshal(in.Content.Highlights)
	if err != nil {
		return fmt.Errorf("marshal highlights: %w", err)
	}
	issues, err := json.Marshal(in.Content.Issues)
	if err != nil {
		return fmt.Errorf("marshal issues: %w", err)
	}
	nextActions, err := json.Marshal(in.Content.NextActions)
	if err != nil {
		return fmt.Errorf("marshal next_actions: %w", err)
	}
	retryFocusCodes := in.Content.RetryFocusDimensionCodes
	if retryFocusCodes == nil {
		retryFocusCodes = []string{}
	}
	retryFocus, err := json.Marshal(retryFocusCodes)
	if err != nil {
		return fmt.Errorf("marshal retry_focus_dimension_codes: %w", err)
	}
	dimensionAssessments, err := json.Marshal(in.Content.DimensionAssessments)
	if err != nil {
		return fmt.Errorf("marshal dimension_assessments: %w", err)
	}
	if err := assertNoReviewPersistencePII(map[string]any{
		"highlights":            json.RawMessage(highlights),
		"issues":                json.RawMessage(issues),
		"next_actions":          json.RawMessage(nextActions),
		"retry_focus":           json.RawMessage(retryFocus),
		"dimension_assessments": json.RawMessage(dimensionAssessments),
	}); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `
update feedback_reports
set status = 'ready',
    summary = $1,
    preparedness_level = $2,
    highlights = $3,
    issues = $4,
    next_actions = $5,
    prompt_version = $6,
    rubric_version = $7,
    model_id = $8,
    provider = $9,
    language = $10,
    feature_flag = $11,
    data_source_version = $12,
    retry_focus_dimension_codes = $13,
    dimension_assessments = $14,
    error_code = null,
	generated_at = $15,
	updated_at = $15
where id = $16 and status = 'generating'`,
		in.Content.Summary,
		string(in.Content.PreparednessLevel),
		highlights,
		issues,
		nextActions,
		in.PromptVersion,
		in.RubricVersion,
		in.ModelID,
		in.Provider,
		in.Language,
		in.FeatureFlag,
		in.DataSourceVersion,
		pq.Array(retryFocusCodes),
		dimensionAssessments,
		in.Now,
		in.ReportID,
	)
	if err != nil {
		return fmt.Errorf("update feedback_reports ready: %w", err)
	}
	if err := requireOneRow(res, "update feedback_reports ready"); err != nil {
		return err
	}
	payload, err := BuildReportGeneratedPayload(ReportGeneratedInput{
		ReportID:          in.ReportID,
		SessionID:         in.SessionID,
		TargetJobID:       in.TargetJobID,
		PreparednessLevel: in.Content.PreparednessLevel,
		PromptVersion:     in.PromptVersion,
		RubricVersion:     in.RubricVersion,
		ModelID:           in.ModelID,
	})
	if err != nil {
		return err
	}
	if err := insertReviewOutbox(ctx, tx, in.OutboxEventID, string(sharedevents.EventNameReportGenerated), in.ReportID, payload, in.Now); err != nil {
		return err
	}
	if err := insertReviewAudit(ctx, tx, in.AuditEventID, in.UserID, "feedback_report.generated", in.ReportID, "success", map[string]any{"status": "ready"}, in.Now); err != nil {
		return err
	}
	if err := updateAsyncJobSucceededTx(ctx, tx, in.AsyncJobID, in.ClaimedAttempts, in.Now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit persist report: %w", err)
	}
	return nil
}

func validateReportPersistenceProvenance(in PersistReportInput) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "prompt_version", value: in.PromptVersion},
		{name: "rubric_version", value: in.RubricVersion},
		{name: "model_id", value: in.ModelID},
		{name: "provider", value: in.Provider},
		{name: "language", value: in.Language},
		{name: "feature_flag", value: in.FeatureFlag},
		{name: "data_source_version", value: in.DataSourceVersion},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("review persistence provenance %s is required", field.name)
		}
	}
	return nil
}

func (r *Repository) PersistReportResult(ctx context.Context, in reviewdomain.ReportResultPersistence) error {
	return r.PersistReport(ctx, in)
}

func insertReviewOutbox(ctx context.Context, tx *sql.Tx, eventID, eventName, reportID string, payload any, now time.Time) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal review outbox payload: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, created_at, next_attempt_at
) values ($1,$2,1,'feedback_report',$3,$4,$5,$5)`,
		eventID,
		eventName,
		reportID,
		raw,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert outbox_events: %w", err)
	}
	return nil
}

func insertReviewAudit(ctx context.Context, tx *sql.Tx, auditID, userID, action, reportID, result string, metadata map[string]any, now time.Time) error {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal review audit metadata: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, action, resource_type, resource_id, result, metadata, created_at
) values ($1,$2,'system',$3,'feedback_report',$4,$5,$6,$7)`,
		auditID,
		userID,
		action,
		reportID,
		result,
		raw,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert audit_events: %w", err)
	}
	return nil
}

func updateAsyncJobSucceededTx(ctx context.Context, tx *sql.Tx, jobID string, claimedAttempts int32, now time.Time) error {
	res, err := tx.ExecContext(ctx, `
update async_jobs
set status = 'succeeded',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = null,
    error_message = null
where id = $2
  and status = 'running'
  and attempts = $3`, now, jobID, claimedAttempts)
	if err != nil {
		return fmt.Errorf("update async_jobs succeeded: %w", err)
	}
	return requireCurrentAsyncJobLeaseWrite(res, jobID, claimedAttempts)
}

func requireOneRow(res sql.Result, label string) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s rows affected: %w", label, err)
	}
	if rows != 1 {
		return fmt.Errorf("%s: expected 1 row, got %d", label, rows)
	}
	return nil
}

func assertNoReviewPersistencePII(payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	lower := strings.ToLower(string(raw))
	for _, forbidden := range []string{"question_text", "questiontext", "answer_text", "answertext", "hint_text", "hinttext", "prompt body", "prompt_body", "response body", "response_body", "provider_secret", "providersecret"} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("review persistence payload contains forbidden field %q", forbidden)
		}
	}
	return nil
}
