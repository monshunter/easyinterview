package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	assessments, err := r.loadQuestionAssessments(ctx, report.ID)
	if err != nil {
		return reviewdomain.FeedbackReportRecord{}, err
	}
	report.QuestionAssessments = assessments
	return report, nil
}

func (r *Repository) ListTargetJobReports(ctx context.Context, in reviewdomain.ListTargetJobReportsInput) (reviewdomain.ListTargetJobReportsResult, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.ListTargetJobReportsResult{}, err
	}
	pageSize := reviewdomain.EffectiveReportPageSize(in.PageSize)
	args := []any{in.UserID, in.TargetJobID}
	query := `
select fr.id, fr.session_id, fr.target_job_id, fr.status, fr.preparedness_level,
       fr.highlights, fr.issues, fr.next_actions, fr.prompt_version, fr.rubric_version,
       fr.model_id, fr.provider, fr.language, fr.feature_flag, fr.data_source_version,
       fr.retry_focus_turn_ids, fr.error_code, fr.created_at, fr.updated_at
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
	for i := range items {
		assessments, err := r.loadQuestionAssessments(ctx, items[i].ID)
		if err != nil {
			return reviewdomain.ListTargetJobReportsResult{}, err
		}
		items[i].QuestionAssessments = assessments
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

func (r *Repository) getFeedbackReport(ctx context.Context, userID, reportID string) (reviewdomain.FeedbackReportRecord, error) {
	row := r.db.QueryRowContext(ctx, `
select fr.id, fr.session_id, fr.target_job_id, fr.status, fr.preparedness_level,
       fr.highlights, fr.issues, fr.next_actions, fr.prompt_version, fr.rubric_version,
       fr.model_id, fr.provider, fr.language, fr.feature_flag, fr.data_source_version,
       fr.retry_focus_turn_ids, fr.error_code, fr.created_at, fr.updated_at
from feedback_reports fr
where fr.id = $1 and fr.user_id = $2`, reportID, userID)
	return scanFeedbackReport(row)
}

type feedbackReportScanner interface {
	Scan(dest ...any) error
}

func scanFeedbackReport(row feedbackReportScanner) (reviewdomain.FeedbackReportRecord, error) {
	var (
		report               reviewdomain.FeedbackReportRecord
		status               string
		preparedness         sql.NullString
		highlightsRaw        []byte
		issuesRaw            []byte
		nextActionsRaw       []byte
		promptVersion        sql.NullString
		rubricVersion        sql.NullString
		modelID              sql.NullString
		provider             sql.NullString
		language             string
		featureFlag          string
		dataSourceVersion    string
		retryFocusTurnIDsRaw []byte
		errorCode            sql.NullString
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
		&retryFocusTurnIDsRaw,
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
	if len(retryFocusTurnIDsRaw) > 0 {
		if err := json.Unmarshal(retryFocusTurnIDsRaw, &report.RetryFocusTurnIDs); err != nil {
			return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode retry_focus_turn_ids: %w", err)
		}
	}
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

func (r *Repository) loadQuestionAssessments(ctx context.Context, reportID string) ([]reviewdomain.QuestionAssessmentRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
select qa.turn_id, coalesce(qa.question_intent, ''), qa.dimension_results,
       qa.review_status, qa.included_in_retry_plan
from question_assessments qa
join practice_turns pt on pt.id = qa.turn_id
where qa.report_id = $1
order by pt.turn_index asc, qa.created_at asc`, reportID)
	if err != nil {
		return nil, fmt.Errorf("load question_assessments: %w", err)
	}
	defer rows.Close()
	out := []reviewdomain.QuestionAssessmentRecord{}
	for rows.Next() {
		var rec reviewdomain.QuestionAssessmentRecord
		var dimensionsRaw []byte
		var reviewStatus string
		if err := rows.Scan(&rec.TurnID, &rec.QuestionIntent, &dimensionsRaw, &reviewStatus, &rec.IncludedInRetryPlan); err != nil {
			return nil, fmt.Errorf("scan question_assessment: %w", err)
		}
		rec.ReviewStatus = sharedtypes.QuestionReviewStatus(reviewStatus)
		dimensions, err := decodeDimensionResults(dimensionsRaw)
		if err != nil {
			return nil, fmt.Errorf("decode dimension_results: %w", err)
		}
		rec.DimensionResults = dimensions
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("question_assessments rows: %w", err)
	}
	return out, nil
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

func decodeDimensionResults(raw []byte) (map[string]reviewdomain.DimensionResultRecord, error) {
	if len(raw) == 0 {
		return map[string]reviewdomain.DimensionResultRecord{}, nil
	}
	var payload map[string]struct {
		Status     sharedtypes.DimensionStatus `json:"status"`
		Confidence any                         `json:"confidence"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	out := make(map[string]reviewdomain.DimensionResultRecord, len(payload))
	for key, item := range payload {
		out[key] = reviewdomain.DimensionResultRecord{
			Status:     item.Status,
			Confidence: confidenceFromJSON(item.Confidence),
		}
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
