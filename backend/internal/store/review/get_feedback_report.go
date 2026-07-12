package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *Repository) GetFeedbackReport(ctx context.Context, userID, reportID string) (reviewdomain.FeedbackReportRecord, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.FeedbackReportRecord{}, err
	}
	report, err := r.getFeedbackReport(ctx, userID, reportID)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.FeedbackReportRecord{}, reviewdomain.ErrReportNotFound
	}
	if err != nil {
		return reviewdomain.FeedbackReportRecord{}, err
	}
	return report, nil
}

func (r *Repository) ListTargetJobReports(ctx context.Context, in reviewdomain.ListTargetJobReportsInput) (reviewdomain.ListTargetJobReportsResult, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.ListTargetJobReportsResult{}, err
	}
	owned, err := r.targetJobOwnedByUser(ctx, in.TargetJobID, in.UserID)
	if err != nil {
		return reviewdomain.ListTargetJobReportsResult{}, err
	}
	if !owned {
		return reviewdomain.ListTargetJobReportsResult{}, reviewdomain.ErrReportNotFound
	}
	pageSize := reviewdomain.EffectiveReportPageSize(in.PageSize)
	args := []any{in.UserID, in.TargetJobID}
	query := `
select fr.id, fr.session_id, fr.target_job_id, fr.status, fr.preparedness_level,
       fr.highlights, fr.issues, fr.next_actions, fr.prompt_version, fr.rubric_version,
       fr.model_id, fr.provider, fr.language, fr.feature_flag, fr.data_source_version,
       fr.dimension_assessments, fr.retry_focus_competency_codes, fr.error_code, fr.created_at, fr.updated_at
from feedback_reports fr
where fr.user_id = $1 and fr.target_job_id = $2`
	if strings.TrimSpace(in.Cursor) != "" {
		if in.CursorCreatedAt.IsZero() || strings.TrimSpace(in.CursorID) == "" {
			createdAt, id, err := reviewdomain.DecodeCursor(in.Cursor)
			if err != nil {
				return reviewdomain.ListTargetJobReportsResult{}, reviewdomain.ErrInvalidCursor
			}
			in.CursorCreatedAt = createdAt
			in.CursorID = id
		}
		args = append(args, in.CursorCreatedAt, in.CursorID)
		query += ` and (fr.created_at, fr.id) < ($3, $4)`
	}
	args = append(args, pageSize+1)
	query += fmt.Sprintf(`
order by fr.created_at desc, fr.id desc
limit $%d`, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return reviewdomain.ListTargetJobReportsResult{}, fmt.Errorf("list feedback_reports: %w", err)
	}
	defer rows.Close()

	items := make([]reviewdomain.FeedbackReportRecord, 0, pageSize)
	for rows.Next() {
		item, err := scanFeedbackReport(rows)
		if err != nil {
			return reviewdomain.ListTargetJobReportsResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return reviewdomain.ListTargetJobReportsResult{}, fmt.Errorf("list feedback_reports rows: %w", err)
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	nextCursor := ""
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = reviewdomain.EncodeCursor(last.CreatedAt, last.ID)
	}
	return reviewdomain.ListTargetJobReportsResult{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) targetJobOwnedByUser(ctx context.Context, targetJobID, userID string) (bool, error) {
	var owned bool
	err := r.db.QueryRowContext(ctx, `
	select exists(
	  select 1
	  from target_jobs
	  where id = $1 and user_id = $2 and deleted_at is null
	)`, targetJobID, userID).Scan(&owned)
	if err != nil {
		return false, fmt.Errorf("check target_jobs ownership: %w", err)
	}
	return owned, nil
}

func (r *Repository) getFeedbackReport(ctx context.Context, userID, reportID string) (reviewdomain.FeedbackReportRecord, error) {
	row := r.db.QueryRowContext(ctx, `
select fr.id, fr.session_id, fr.target_job_id, fr.status, fr.preparedness_level,
       fr.highlights, fr.issues, fr.next_actions, fr.prompt_version, fr.rubric_version,
       fr.model_id, fr.provider, fr.language, fr.feature_flag, fr.data_source_version,
       fr.dimension_assessments, fr.retry_focus_competency_codes, fr.error_code, fr.created_at, fr.updated_at
from feedback_reports fr
where fr.id = $1 and fr.user_id = $2`, reportID, userID)
	return scanFeedbackReport(row)
}

type feedbackReportScanner interface {
	Scan(dest ...any) error
}

func scanFeedbackReport(row feedbackReportScanner) (reviewdomain.FeedbackReportRecord, error) {
	var (
		report                    reviewdomain.FeedbackReportRecord
		status                    string
		preparedness              sql.NullString
		highlightsRaw             []byte
		issuesRaw                 []byte
		nextActionsRaw            []byte
		promptVersion             sql.NullString
		rubricVersion             sql.NullString
		modelID                   sql.NullString
		provider                  sql.NullString
		language                  string
		featureFlag               string
		dataSourceVersion         string
		dimensionAssessmentsRaw   []byte
		retryFocusCompetencyCodes pq.StringArray
		errorCode                 sql.NullString
	)
	if err := row.Scan(
		&report.ID,
		&report.SessionID,
		&report.TargetJobID,
		&status,
		&preparedness,
		&highlightsRaw,
		&issuesRaw,
		&nextActionsRaw,
		&promptVersion,
		&rubricVersion,
		&modelID,
		&provider,
		&language,
		&featureFlag,
		&dataSourceVersion,
		&dimensionAssessmentsRaw,
		&retryFocusCompetencyCodes,
		&errorCode,
		&report.CreatedAt,
		&report.UpdatedAt,
	); err != nil {
		return reviewdomain.FeedbackReportRecord{}, err
	}
	report.Status = sharedtypes.ReportStatus(status)
	if preparedness.Valid && strings.TrimSpace(preparedness.String) != "" {
		tier := sharedtypes.ReadinessTier(preparedness.String)
		report.PreparednessLevel = &tier
	}
	highlights, err := decodeReportEvidence(highlightsRaw)
	if err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode highlights: %w", err)
	}
	report.Highlights = highlights
	issues, err := decodeReportEvidence(issuesRaw)
	if err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode issues: %w", err)
	}
	report.Issues = issues
	if len(nextActionsRaw) > 0 {
		if err := json.Unmarshal(nextActionsRaw, &report.NextActions); err != nil {
			return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode next_actions: %w", err)
		}
	}
	if len(dimensionAssessmentsRaw) > 0 {
		if err := json.Unmarshal(dimensionAssessmentsRaw, &report.DimensionAssessments); err != nil {
			return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode dimension_assessments: %w", err)
		}
	}
	report.RetryFocusCompetencyCodes = append([]string(nil), retryFocusCompetencyCodes...)
	if errorCode.Valid && strings.TrimSpace(errorCode.String) != "" {
		code := errorCode.String
		report.ErrorCode = &code
	}
	if promptVersion.Valid && rubricVersion.Valid && modelID.Valid {
		report.Provenance = &reviewdomain.GenerationProvenanceRecord{
			PromptVersion:     promptVersion.String,
			RubricVersion:     rubricVersion.String,
			ModelID:           modelID.String,
			Language:          language,
			FeatureFlag:       featureFlag,
			DataSourceVersion: dataSourceVersion,
		}
		_ = provider
	}
	return report, nil
}

func decodeReportEvidence(raw []byte) ([]reviewdomain.ReportEvidenceRecord, error) {
	if len(raw) == 0 {
		return []reviewdomain.ReportEvidenceRecord{}, nil
	}
	var payload []struct {
		Dimension  string `json:"dimension"`
		Evidence   string `json:"evidence"`
		Confidence any    `json:"confidence"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	out := make([]reviewdomain.ReportEvidenceRecord, 0, len(payload))
	for _, item := range payload {
		out = append(out, reviewdomain.ReportEvidenceRecord{
			Dimension:  item.Dimension,
			Evidence:   item.Evidence,
			Confidence: confidenceFromJSON(item.Confidence),
		})
	}
	return out, nil
}

func confidenceFromJSON(value any) sharedtypes.Confidence {
	switch v := value.(type) {
	case string:
		return sharedtypes.Confidence(v)
	case float64:
		return confidenceFromScore(v)
	default:
		return sharedtypes.ConfidenceLow
	}
}
