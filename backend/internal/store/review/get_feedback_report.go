package review

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
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

func (r *Repository) ListTargetJobReports(ctx context.Context, userID, targetJobID string) (reviewdomain.TargetJobReportsOverviewRecord, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, err
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("begin target job report overview snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var currentSummaryRaw []byte
	err = tx.QueryRowContext(ctx, `
select tj.summary
from target_jobs tj
where tj.id = $1 and tj.user_id = $2 and tj.deleted_at is null`, targetJobID, userID).Scan(&currentSummaryRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.TargetJobReportsOverviewRecord{}, reviewdomain.ErrReportNotFound
	}
	if err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("load target job report overview summary: %w", err)
	}
	currentRounds, err := practicedomain.ParseCanonicalReportRounds(currentSummaryRaw)
	if err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("%w: current target job canonical summary: %v", reviewdomain.ErrReportContextInvalid, err)
	}
	overview := reviewdomain.TargetJobReportsOverviewRecord{
		TargetJobID: targetJobID,
		Rounds:      make([]reviewdomain.TargetJobReportRoundOverviewRecord, len(currentRounds)),
	}
	roundIndexes := make(map[reportRoundPair]int, len(currentRounds))
	for i, round := range currentRounds {
		overview.Rounds[i].Round = reviewdomain.PracticeRoundRefRecord{RoundID: round.ID, RoundSequence: round.Sequence}
		roundIndexes[reportRoundPair{ID: round.ID, Sequence: round.Sequence}] = i
	}

	rows, err := tx.QueryContext(ctx, `
select fr.id::text, fr.user_id::text, fr.session_id::text, fr.target_job_id::text,
       ps.user_id::text, ps.target_job_id::text, fr.status, fr.error_code,
       fr.generation_context, fr.generated_at, fr.created_at
from feedback_reports fr
left join practice_sessions ps on ps.id = fr.session_id
where fr.target_job_id = $1 or ps.target_job_id = $1`, targetJobID)
	if err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("list target job report attempts: %w", err)
	}
	defer rows.Close()

	currentSelections := make([]targetJobCurrentSelection, len(currentRounds))
	for rows.Next() {
		attempt, err := scanTargetJobReportAttempt(rows)
		if err != nil {
			return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("%w: scan report overview attempt: %v", reviewdomain.ErrReportContextInvalid, err)
		}
		if err := validateTargetJobReportAttemptIdentity(attempt, userID, targetJobID); err != nil {
			return reviewdomain.TargetJobReportsOverviewRecord{}, err
		}
		if len(bytes.TrimSpace(attempt.GenerationContext)) == 0 {
			return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("%w: report overview attempt", reviewdomain.ErrReportContextMissing)
		}
		frozen, err := decodeFrozenReportContext(attempt.GenerationContext)
		if err != nil {
			return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("%w: decode frozen report context: %v", reviewdomain.ErrReportContextInvalid, err)
		}
		if attempt.SessionID != frozen.Conversation.SessionID || attempt.TargetJobID != frozen.TargetJob.ID {
			return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("%w: frozen report row identity mismatch", reviewdomain.ErrReportContextInvalid)
		}
		roundIndex, ok := roundIndexes[reportRoundPair{ID: frozen.Round.ID, Sequence: frozen.Round.Sequence}]
		if !ok {
			return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("%w: frozen report round pair is not current canonical round", reviewdomain.ErrReportContextInvalid)
		}
		if err := validateTargetJobReportAttemptState(attempt); err != nil {
			return reviewdomain.TargetJobReportsOverviewRecord{}, err
		}

		latest := reviewdomain.TargetJobReportAttemptSummaryRecord{
			ID: attempt.ID, Status: attempt.Status, CreatedAt: attempt.CreatedAt,
		}
		if attempt.Status == sharedtypes.ReportStatusFailed {
			code := strings.TrimSpace(attempt.ErrorCode.String)
			latest.ErrorCode = &code
		}
		if shouldReplaceLatestAttempt(latest, overview.Rounds[roundIndex].LatestAttempt) {
			overview.Rounds[roundIndex].LatestAttempt = &latest
		}
		if attempt.Status == sharedtypes.ReportStatusReady {
			candidate := targetJobCurrentSelection{
				Summary:   reviewdomain.TargetJobCurrentReportSummaryRecord{ID: attempt.ID, GeneratedAt: attempt.GeneratedAt.Time},
				CreatedAt: attempt.CreatedAt,
				Present:   true,
			}
			if shouldReplaceCurrentReport(candidate, currentSelections[roundIndex]) {
				currentSelections[roundIndex] = candidate
				current := candidate.Summary
				overview.Rounds[roundIndex].CurrentReport = &current
			}
		}
	}
	if err := rows.Err(); err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("iterate target job report attempts: %w", err)
	}
	if err := rows.Close(); err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("close target job report attempts: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return reviewdomain.TargetJobReportsOverviewRecord{}, fmt.Errorf("commit target job report overview snapshot: %w", err)
	}
	return overview, nil
}

type reportRoundPair struct {
	ID       string
	Sequence int32
}

type targetJobReportAttempt struct {
	ID                 string
	UserID             string
	SessionID          string
	TargetJobID        string
	SessionUserID      sql.NullString
	SessionTargetJobID sql.NullString
	Status             sharedtypes.ReportStatus
	ErrorCode          sql.NullString
	GenerationContext  []byte
	GeneratedAt        sql.NullTime
	CreatedAt          time.Time
}

func scanTargetJobReportAttempt(row feedbackReportScanner) (targetJobReportAttempt, error) {
	var attempt targetJobReportAttempt
	var status string
	if err := row.Scan(
		&attempt.ID,
		&attempt.UserID,
		&attempt.SessionID,
		&attempt.TargetJobID,
		&attempt.SessionUserID,
		&attempt.SessionTargetJobID,
		&status,
		&attempt.ErrorCode,
		&attempt.GenerationContext,
		&attempt.GeneratedAt,
		&attempt.CreatedAt,
	); err != nil {
		return targetJobReportAttempt{}, err
	}
	attempt.Status = sharedtypes.ReportStatus(status)
	return attempt, nil
}

func validateTargetJobReportAttemptIdentity(attempt targetJobReportAttempt, userID, targetJobID string) error {
	if attempt.UserID != userID || attempt.TargetJobID != targetJobID ||
		!attempt.SessionUserID.Valid || attempt.SessionUserID.String != userID ||
		!attempt.SessionTargetJobID.Valid || attempt.SessionTargetJobID.String != targetJobID ||
		strings.TrimSpace(attempt.SessionID) == "" {
		return fmt.Errorf("%w: report attempt identity mismatch", reviewdomain.ErrReportContextInvalid)
	}
	return nil
}

func validateTargetJobReportAttemptState(attempt targetJobReportAttempt) error {
	switch attempt.Status {
	case sharedtypes.ReportStatusQueued, sharedtypes.ReportStatusGenerating:
		return nil
	case sharedtypes.ReportStatusReady:
		if !attempt.GeneratedAt.Valid {
			return fmt.Errorf("%w: ready report generated_at is null", reviewdomain.ErrReportContextInvalid)
		}
		return nil
	case sharedtypes.ReportStatusFailed:
		code := strings.TrimSpace(attempt.ErrorCode.String)
		if !attempt.ErrorCode.Valid || code == "" {
			return fmt.Errorf("%w: failed report error code is missing", reviewdomain.ErrReportContextInvalid)
		}
		if !isKnownAPIErrorCode(code) {
			return fmt.Errorf("%w: failed report error code is invalid", reviewdomain.ErrReportContextInvalid)
		}
		return nil
	default:
		return fmt.Errorf("%w: report status is invalid", reviewdomain.ErrReportContextInvalid)
	}
}

func isKnownAPIErrorCode(code string) bool {
	for _, candidate := range api.AllApiErrorCodes {
		if string(candidate) == code {
			return true
		}
	}
	return false
}

type targetJobCurrentSelection struct {
	Summary   reviewdomain.TargetJobCurrentReportSummaryRecord
	CreatedAt time.Time
	Present   bool
}

func shouldReplaceCurrentReport(candidate, current targetJobCurrentSelection) bool {
	if !current.Present {
		return true
	}
	if !candidate.Summary.GeneratedAt.Equal(current.Summary.GeneratedAt) {
		return candidate.Summary.GeneratedAt.After(current.Summary.GeneratedAt)
	}
	if !candidate.CreatedAt.Equal(current.CreatedAt) {
		return candidate.CreatedAt.After(current.CreatedAt)
	}
	return candidate.Summary.ID > current.Summary.ID
}

func shouldReplaceLatestAttempt(candidate reviewdomain.TargetJobReportAttemptSummaryRecord, current *reviewdomain.TargetJobReportAttemptSummaryRecord) bool {
	if current == nil {
		return true
	}
	if !candidate.CreatedAt.Equal(current.CreatedAt) {
		return candidate.CreatedAt.After(current.CreatedAt)
	}
	return candidate.ID > current.ID
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
