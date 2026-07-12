package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type PersistReportInput = reviewdomain.ReportResultPersistence

func (r *Repository) PersistReport(ctx context.Context, in PersistReportInput) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin persist report: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	highlights, err := json.Marshal(wireReportEvidence(in.Content.Highlights))
	if err != nil {
		return fmt.Errorf("marshal highlights: %w", err)
	}
	issues, err := json.Marshal(wireReportEvidence(in.Content.Issues))
	if err != nil {
		return fmt.Errorf("marshal issues: %w", err)
	}
	nextActions, err := json.Marshal(in.Content.NextActions)
	if err != nil {
		return fmt.Errorf("marshal next_actions: %w", err)
	}
	retryFocus, err := json.Marshal(in.RetryFocusCompetencyCodes)
	if err != nil {
		return fmt.Errorf("marshal retry_focus_competency_codes: %w", err)
	}
	dimensionAssessments, err := json.Marshal(wireDimensionAssessments(in.DimensionAssessments))
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
    preparedness_level = $1,
    highlights = $2,
    issues = $3,
    next_actions = $4,
    prompt_version = $5,
    rubric_version = $6,
    model_id = $7,
    provider = $8,
    language = $9,
    feature_flag = $10,
    data_source_version = $11,
    retry_focus_competency_codes = $12,
    dimension_assessments = $13,
    error_code = null,
    generated_at = $14,
    updated_at = $14
where id = $15 and status = 'generating'`,
		string(in.PreparednessLevel),
		highlights,
		issues,
		nextActions,
		in.PromptVersion,
		in.RubricVersion,
		in.ModelID,
		nullableString(in.Provider),
		fallbackString(in.Language, "en"),
		fallbackString(in.FeatureFlag, "none"),
		fallbackString(in.DataSourceVersion, "not_applicable"),
		retryFocus,
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
		PreparednessLevel: in.PreparednessLevel,
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
	if err := updateAsyncJobSucceededTx(ctx, tx, in.AsyncJobID, in.Now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit persist report: %w", err)
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

func updateAsyncJobSucceededTx(ctx context.Context, tx *sql.Tx, jobID string, now time.Time) error {
	res, err := tx.ExecContext(ctx, `
update async_jobs
set status = 'succeeded',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = null,
    error_message = null
where id = $2`, now, jobID)
	if err != nil {
		return fmt.Errorf("update async_jobs succeeded: %w", err)
	}
	if err := requireOneRow(res, "update async_jobs succeeded"); err != nil {
		return err
	}
	return nil
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

func wireDimensionAssessments(in []reviewdomain.DimensionAssessmentDraft) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, value := range in {
		out = append(out, map[string]any{
			"dimension":  value.Dimension,
			"status":     string(value.Status),
			"confidence": string(value.Confidence),
		})
	}
	return out
}

func wireReportEvidence(in []reviewdomain.ReportEvidenceDraft) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, item := range in {
		out = append(out, map[string]any{
			"dimension":  item.Dimension,
			"evidence":   item.Evidence,
			"confidence": string(confidenceFromScore(item.Confidence)),
		})
	}
	return out
}

func confidenceFromScore(score float64) sharedtypes.Confidence {
	switch {
	case score >= 0.75:
		return sharedtypes.ConfidenceHigh
	case score >= 0.4:
		return sharedtypes.ConfidenceMedium
	default:
		return sharedtypes.ConfidenceLow
	}
}

func fallbackString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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
