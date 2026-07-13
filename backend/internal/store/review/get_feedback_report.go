package review

import (
	"bytes"
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
	select fr.id::text, fr.session_id::text, fr.target_job_id::text, fr.status, fr.error_code, fr.summary, fr.generation_context,
	       fr.preparedness_level, fr.dimension_assessments, fr.highlights, fr.issues, fr.next_actions,
	       fr.retry_focus_dimension_codes, fr.prompt_version, fr.rubric_version, fr.model_id,
	       fr.language, fr.feature_flag, fr.data_source_version, fr.created_at, fr.updated_at
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
	select fr.id::text, fr.session_id::text, fr.target_job_id::text, fr.status, fr.error_code, fr.summary, fr.generation_context,
	       fr.preparedness_level, fr.dimension_assessments, fr.highlights, fr.issues, fr.next_actions,
	       fr.retry_focus_dimension_codes, fr.prompt_version, fr.rubric_version, fr.model_id,
	       fr.language, fr.feature_flag, fr.data_source_version, fr.created_at, fr.updated_at
	from feedback_reports fr
where fr.id = $1 and fr.user_id = $2`, reportID, userID)
	return scanFeedbackReport(row)
}

type feedbackReportScanner interface {
	Scan(dest ...any) error
}

func scanFeedbackReport(row feedbackReportScanner) (reviewdomain.FeedbackReportRecord, error) {
	var (
		report                   reviewdomain.FeedbackReportRecord
		status                   string
		errorCode                sql.NullString
		summary                  sql.NullString
		generationContextRaw     []byte
		preparedness             sql.NullString
		dimensionAssessmentsRaw  []byte
		highlightsRaw            []byte
		issuesRaw                []byte
		nextActionsRaw           []byte
		retryFocusDimensionCodes pq.StringArray
		promptVersion            sql.NullString
		rubricVersion            sql.NullString
		modelID                  sql.NullString
		language                 string
		featureFlag              string
		dataSourceVersion        string
	)
	if err := row.Scan(
		&report.ID,
		&report.SessionID,
		&report.TargetJobID,
		&status,
		&errorCode,
		&summary,
		&generationContextRaw,
		&preparedness,
		&dimensionAssessmentsRaw,
		&highlightsRaw,
		&issuesRaw,
		&nextActionsRaw,
		&retryFocusDimensionCodes,
		&promptVersion,
		&rubricVersion,
		&modelID,
		&language,
		&featureFlag,
		&dataSourceVersion,
		&report.CreatedAt,
		&report.UpdatedAt,
	); err != nil {
		return reviewdomain.FeedbackReportRecord{}, err
	}
	report.Status = sharedtypes.ReportStatus(status)
	frozen, err := decodeFrozenReportContext(generationContextRaw)
	if err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode generation_context: %w", err)
	}
	if report.SessionID != frozen.Conversation.SessionID || report.TargetJobID != frozen.TargetJob.ID {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("frozen report projection row identity mismatch")
	}
	report.Context = reviewdomain.ProjectFrozenReportContext(frozen)
	report.Highlights = []reviewdomain.ReportEvidenceRecord{}
	report.Issues = []reviewdomain.ReportEvidenceRecord{}
	report.NextActions = []reviewdomain.ReportNextActionRecord{}
	report.DimensionAssessments = []reviewdomain.DimensionAssessmentRecord{}
	report.RetryFocusDimensionCodes = []string{}
	if errorCode.Valid && strings.TrimSpace(errorCode.String) != "" {
		code := errorCode.String
		report.ErrorCode = &code
	}
	if report.Status != sharedtypes.ReportStatusReady {
		return report, nil
	}
	if !summary.Valid || strings.TrimSpace(summary.String) == "" || !preparedness.Valid || strings.TrimSpace(preparedness.String) == "" {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("ready report summary/preparedness is incomplete")
	}
	summaryValue := summary.String
	report.Summary = &summaryValue
	tier := sharedtypes.ReadinessTier(preparedness.String)
	report.PreparednessLevel = &tier
	if err := decodeClosedJSON(dimensionAssessmentsRaw, &report.DimensionAssessments); err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode dimension_assessments: %w", err)
	}
	if err := decodeClosedJSON(highlightsRaw, &report.Highlights); err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode highlights: %w", err)
	}
	if err := decodeClosedJSON(issuesRaw, &report.Issues); err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode issues: %w", err)
	}
	if err := decodeClosedJSON(nextActionsRaw, &report.NextActions); err != nil {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("decode next_actions: %w", err)
	}
	report.RetryFocusDimensionCodes = append([]string(nil), retryFocusDimensionCodes...)
	if !promptVersion.Valid || !rubricVersion.Valid || !modelID.Valid {
		return reviewdomain.FeedbackReportRecord{}, fmt.Errorf("ready report provenance is incomplete")
	}
	report.Provenance = &reviewdomain.GenerationProvenanceRecord{
		PromptVersion:     promptVersion.String,
		RubricVersion:     rubricVersion.String,
		ModelID:           modelID.String,
		Language:          language,
		FeatureFlag:       featureFlag,
		DataSourceVersion: dataSourceVersion,
	}
	return report, nil
}

func decodeClosedJSON(raw []byte, destination any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = []byte(`[]`)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return nil
}
