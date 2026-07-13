package practice

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type derivedReportDimension struct {
	Code       string `json:"code"`
	Label      string `json:"label"`
	Status     string `json:"status"`
	Confidence string `json:"confidence"`
}

type derivedReportIssue struct {
	DimensionCode      string  `json:"dimensionCode"`
	Evidence           string  `json:"evidence"`
	Confidence         string  `json:"confidence"`
	SourceMessageSeqNo []int32 `json:"sourceMessageSeqNos"`
}

type derivedReportSource struct {
	TargetJobID              string
	SessionID                string
	Context                  domain.ReportContextSnapshot
	Dimensions               []derivedReportDimension
	Issues                   []derivedReportIssue
	RetryFocusDimensionCodes []string
	SourcePlanID             string
	SourcePlanTargetJobID    string
	SourceResumeID           string
	SourceRoundID            string
	SourceRoundSequence      int32
	SourceInterviewerPersona string
	SourceDifficulty         string
	SourceLanguage           string
	SourceTimeBudgetMinutes  int32
}

func createDerivedPlan(ctx context.Context, tx *sql.Tx, in domain.CreatePlanStoreInput) (domain.PlanRecord, error) {
	source, err := loadDerivedReportSource(ctx, tx, in.SourceReportID, in.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
			return domain.PlanRecord{}, domain.ErrPlanPrerequisiteNotFound
		}
		return domain.PlanRecord{}, err
	}

	selectedRound, focusCodes, err := derivePlanProjection(in.Goal, source)
	if err != nil {
		return domain.PlanRecord{}, domain.ErrPlanPrerequisiteNotFound
	}

	var (
		plan           domain.PlanRecord
		sourceReportID sql.NullString
		persistedFocus pq.StringArray
	)
	err = tx.QueryRowContext(ctx, `
insert into practice_plans (
  id, user_id, target_job_id, source_report_id, goal, round_id, round_sequence,
  interviewer_persona, difficulty, language, time_budget_minutes,
  resume_id, focus_dimension_codes, status, created_at, updated_at
) values ($1,$2,$3,$4::uuid,$5,$6,$7,$8,$9,$10,$11,$12,$13,'ready',$14,$14)
returning id, target_job_id, source_report_id::text, goal, round_id, round_sequence,
          interviewer_persona, difficulty, language, time_budget_minutes,
          resume_id::text, focus_dimension_codes, status, created_at`,
		in.PlanID,
		in.UserID,
		source.Context.TargetJob.ID,
		in.SourceReportID,
		string(in.Goal),
		selectedRound.ID,
		selectedRound.Sequence,
		source.Context.Plan.InterviewerPersona,
		source.Context.Plan.Difficulty,
		source.Context.Plan.Language,
		selectedRound.DurationMinutes,
		source.Context.Resume.ID,
		pq.Array(focusCodes),
		in.Now,
	).Scan(
		&plan.ID,
		&plan.TargetJobID,
		&sourceReportID,
		&plan.Goal,
		&plan.RoundID,
		&plan.RoundSequence,
		&plan.InterviewerPersona,
		&plan.Difficulty,
		&plan.Language,
		&plan.TimeBudgetMinutes,
		&plan.ResumeID,
		&persistedFocus,
		&plan.Status,
		&plan.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanPrerequisiteNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("insert report-derived practice plan: %w", err)
	}
	plan.SourceReportID = sourceReportID.String
	plan.FocusDimensionCodes = append([]string{}, persistedFocus...)
	return plan, nil
}

func loadDerivedReportSource(ctx context.Context, tx *sql.Tx, reportID, userID string) (derivedReportSource, error) {
	var (
		source          derivedReportSource
		contextRaw      []byte
		dimensionsRaw   []byte
		issuesRaw       []byte
		retryFocusCodes pq.StringArray
	)
	err := tx.QueryRowContext(ctx, `
select fr.target_job_id::text, fr.session_id::text, fr.generation_context,
       fr.dimension_assessments, fr.issues, fr.retry_focus_dimension_codes,
       pp.id::text, pp.target_job_id::text, pp.resume_id::text, pp.round_id,
       pp.round_sequence, pp.interviewer_persona, pp.difficulty, pp.language,
       pp.time_budget_minutes
from feedback_reports fr
join practice_sessions ps
  on ps.id=fr.session_id and ps.user_id=fr.user_id and ps.target_job_id=fr.target_job_id
join practice_plans pp
  on pp.id=ps.plan_id and pp.user_id=ps.user_id and pp.target_job_id=ps.target_job_id
where fr.id=nullif($1,'')::uuid and fr.user_id=$2 and fr.status = 'ready'
  and exists (
    select 1 from practice_session_events pse
    where pse.session_id=ps.id and pse.event_type='session_completed'
  )
for update of fr, ps, pp`, reportID, userID).Scan(
		&source.TargetJobID,
		&source.SessionID,
		&contextRaw,
		&dimensionsRaw,
		&issuesRaw,
		&retryFocusCodes,
		&source.SourcePlanID,
		&source.SourcePlanTargetJobID,
		&source.SourceResumeID,
		&source.SourceRoundID,
		&source.SourceRoundSequence,
		&source.SourceInterviewerPersona,
		&source.SourceDifficulty,
		&source.SourceLanguage,
		&source.SourceTimeBudgetMinutes,
	)
	if err != nil {
		return derivedReportSource{}, err
	}
	if err := decodeStrictJSON(contextRaw, &source.Context); err != nil {
		return derivedReportSource{}, domain.ErrPlanPrerequisiteNotFound
	}
	if err := decodeStrictJSON(dimensionsRaw, &source.Dimensions); err != nil {
		return derivedReportSource{}, domain.ErrPlanPrerequisiteNotFound
	}
	if err := decodeStrictJSON(issuesRaw, &source.Issues); err != nil {
		return derivedReportSource{}, domain.ErrPlanPrerequisiteNotFound
	}
	source.RetryFocusDimensionCodes = append([]string(nil), retryFocusCodes...)
	return source, nil
}

func derivePlanProjection(goal sharedtypes.PracticeGoal, source derivedReportSource) (domain.ReportRoundSnapshot, []string, error) {
	if err := domain.ValidateReportContextSnapshot(source.Context); err != nil {
		return domain.ReportRoundSnapshot{}, nil, err
	}
	if err := validateDerivedSourceBinding(source); err != nil {
		return domain.ReportRoundSnapshot{}, nil, err
	}
	dimensions, issuesByCode, err := validateDerivedReportSemantics(source.Dimensions, source.Issues)
	if err != nil {
		return domain.ReportRoundSnapshot{}, nil, err
	}

	currentIndex := -1
	for i, round := range source.Context.CanonicalRounds {
		if round.ID == source.Context.Round.ID && round.Sequence == source.Context.Round.Sequence {
			currentIndex = i
			if round != source.Context.Round {
				return domain.ReportRoundSnapshot{}, nil, fmt.Errorf("frozen current round content mismatch")
			}
			break
		}
	}
	if currentIndex < 0 {
		return domain.ReportRoundSnapshot{}, nil, fmt.Errorf("frozen current round missing")
	}

	switch goal {
	case sharedtypes.PracticeGoalRetryCurrentRound:
		focus, err := validateRetryFocus(source.RetryFocusDimensionCodes, dimensions, issuesByCode)
		if err != nil {
			return domain.ReportRoundSnapshot{}, nil, err
		}
		return source.Context.CanonicalRounds[currentIndex], focus, nil
	case sharedtypes.PracticeGoalNextRound:
		if currentIndex+1 >= len(source.Context.CanonicalRounds) || !source.Context.HasNextRound {
			return domain.ReportRoundSnapshot{}, nil, fmt.Errorf("frozen report has no canonical successor")
		}
		return source.Context.CanonicalRounds[currentIndex+1], []string{}, nil
	default:
		return domain.ReportRoundSnapshot{}, nil, fmt.Errorf("goal is not report-derived")
	}
}

func validateDerivedSourceBinding(source derivedReportSource) error {
	context := source.Context
	if context.TargetJob.ID != source.TargetJobID || source.SourcePlanTargetJobID != source.TargetJobID ||
		context.Resume.ID != source.SourceResumeID || context.Plan.ResumeID != source.SourceResumeID ||
		context.Plan.ID != source.SourcePlanID || context.Conversation.SessionID != source.SessionID ||
		context.Plan.RoundID != source.SourceRoundID || context.Plan.RoundSequence != source.SourceRoundSequence ||
		context.Plan.InterviewerPersona != source.SourceInterviewerPersona || context.Plan.Difficulty != source.SourceDifficulty ||
		context.Plan.Language != source.SourceLanguage || context.Plan.TimeBudgetMinutes != source.SourceTimeBudgetMinutes {
		return fmt.Errorf("frozen report source binding mismatch")
	}
	return nil
}

func validateDerivedReportSemantics(dimensions []derivedReportDimension, issues []derivedReportIssue) (map[string]derivedReportDimension, map[string][]derivedReportIssue, error) {
	if len(dimensions) < 1 || len(dimensions) > 6 || len(issues) > 4 {
		return nil, nil, fmt.Errorf("report dimension or issue cardinality is invalid")
	}
	dimensionByCode := make(map[string]derivedReportDimension, len(dimensions))
	for _, dimension := range dimensions {
		dimension.Code = strings.TrimSpace(dimension.Code)
		dimension.Label = strings.TrimSpace(dimension.Label)
		if !validDimensionCode(dimension.Code) || dimension.Label == "" || len([]rune(dimension.Label)) > 48 ||
			!oneOf(dimension.Status, "strong", "meets_bar", "needs_work") || !oneOf(dimension.Confidence, "high", "medium", "low") {
			return nil, nil, fmt.Errorf("report dimension is invalid")
		}
		if _, duplicate := dimensionByCode[dimension.Code]; duplicate {
			return nil, nil, fmt.Errorf("report dimension code is duplicated")
		}
		dimensionByCode[dimension.Code] = dimension
	}
	issuesByCode := make(map[string][]derivedReportIssue)
	for _, issue := range issues {
		issue.DimensionCode = strings.TrimSpace(issue.DimensionCode)
		issue.Evidence = strings.TrimSpace(issue.Evidence)
		if _, ok := dimensionByCode[issue.DimensionCode]; !ok || issue.Evidence == "" || len([]rune(issue.Evidence)) > 240 ||
			!oneOf(issue.Confidence, "high", "medium", "low") || !strictPositiveSequence(issue.SourceMessageSeqNo) {
			return nil, nil, fmt.Errorf("report issue is invalid")
		}
		issuesByCode[issue.DimensionCode] = append(issuesByCode[issue.DimensionCode], issue)
	}
	return dimensionByCode, issuesByCode, nil
}

func validateRetryFocus(codes []string, dimensions map[string]derivedReportDimension, issues map[string][]derivedReportIssue) ([]string, error) {
	if len(codes) > 6 {
		return nil, fmt.Errorf("retry focus cardinality is invalid")
	}
	seen := make(map[string]struct{}, len(codes))
	out := make([]string, 0, len(codes))
	for _, rawCode := range codes {
		code := strings.TrimSpace(rawCode)
		dimension, ok := dimensions[code]
		if code == "" || !ok || dimension.Status != "needs_work" || len(issues[code]) == 0 {
			return nil, fmt.Errorf("retry focus is not issue-backed needs-work dimension")
		}
		if _, duplicate := seen[code]; duplicate {
			return nil, fmt.Errorf("retry focus code is duplicated")
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out, nil
}

func resolveDerivedSemanticFocus(codes []string, dimensionsRaw, issuesRaw []byte) ([]domain.SemanticFocusDimension, error) {
	if len(codes) == 0 {
		return []domain.SemanticFocusDimension{}, nil
	}
	var dimensions []derivedReportDimension
	if err := decodeStrictJSON(dimensionsRaw, &dimensions); err != nil {
		return nil, fmt.Errorf("decode report dimensions: %w", err)
	}
	var issues []derivedReportIssue
	if err := decodeStrictJSON(issuesRaw, &issues); err != nil {
		return nil, fmt.Errorf("decode report issues: %w", err)
	}
	dimensionByCode, issuesByCode, err := validateDerivedReportSemantics(dimensions, issues)
	if err != nil {
		return nil, err
	}
	validatedCodes, err := validateRetryFocus(codes, dimensionByCode, issuesByCode)
	if err != nil {
		return nil, err
	}
	focus := make([]domain.SemanticFocusDimension, 0, len(validatedCodes))
	for _, code := range validatedCodes {
		issueEvidence := make([]string, 0, len(issuesByCode[code]))
		for _, issue := range issuesByCode[code] {
			issueEvidence = append(issueEvidence, issue.Evidence)
		}
		focus = append(focus, domain.SemanticFocusDimension{
			Code: code, Label: dimensionByCode[code].Label, Issues: issueEvidence,
		})
	}
	return focus, nil
}

func decodeStrictJSON(raw []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("JSON contains trailing data")
	}
	return nil
}

func validDimensionCode(value string) bool {
	if len(value) < 2 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for i := 1; i < len(value); i++ {
		char := value[i]
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			continue
		}
		return false
	}
	return true
}

func strictPositiveSequence(values []int32) bool {
	if len(values) == 0 {
		return false
	}
	var previous int32
	for _, value := range values {
		if value <= previous {
			return false
		}
		previous = value
	}
	return true
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
