package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	reportGenerateFeatureKey           = string(featurekeys.ReportGenerate)
	reportQuestionAssessmentFeatureKey = string(featurekeys.ReportQuestionAssessment)
)

var ErrReviewAIOutputInvalid = errors.New("review: ai output invalid")

type SessionSnapshot struct {
	UserID      string
	ReportID    string
	SessionID   string
	PlanID      string
	TargetJobID string
	Language    string
}

type PracticePlanSnapshot struct {
	ID                 string
	Goal               string
	Mode               string
	InterviewerPersona string
}

type TurnSnapshot struct {
	ID              string
	TurnIndex       int
	QuestionIntent  string
	QuestionContext string
	AnswerSummary   string
}

type DimensionScoreDraft struct {
	Name                   string   `json:"name"`
	Score                  float64  `json:"score"`
	Reasoning              string   `json:"reasoning"`
	SupportingObservations []string `json:"supporting_observations"`
}

type ReportEvidenceDraft struct {
	Dimension  string  `json:"dimension"`
	Evidence   string  `json:"evidence"`
	Confidence float64 `json:"confidence"`
}

type ReportNextActionDraft struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type ReportContentDraft struct {
	Summary           string                  `json:"summary"`
	DimensionScores   []DimensionScoreDraft   `json:"dimension_scores"`
	Highlights        []ReportEvidenceDraft   `json:"highlights"`
	Issues            []ReportEvidenceDraft   `json:"issues"`
	NextActions       []ReportNextActionDraft `json:"next_actions"`
	RetryFocusTurnIDs []string                `json:"retry_focus_turn_ids"`
}

type DimensionResultDraft struct {
	Status     sharedtypes.DimensionStatus `json:"status"`
	Confidence float64                     `json:"confidence"`
	Score      float64                     `json:"score,omitempty"`
	ScoreLevel string                      `json:"score_level,omitempty"`
}

type QuestionAssessmentDraft struct {
	TurnID               string                           `json:"turn_id"`
	TurnIndex            int                              `json:"turn_index"`
	QuestionIntent       string                           `json:"question_intent"`
	DimensionResults     map[string]DimensionResultDraft  `json:"dimension_results"`
	OverallStatus        sharedtypes.DimensionStatus      `json:"overall_status"`
	Confidence           float64                          `json:"confidence"`
	Strengths            []string                         `json:"strengths"`
	Gaps                 []string                         `json:"gaps"`
	RecommendedFramework string                           `json:"recommended_framework"`
	ReviewStatus         sharedtypes.QuestionReviewStatus `json:"review_status"`
	IncludedInRetryPlan  bool                             `json:"included_in_retry_plan"`
}

func (s *Service) generateReportContent(ctx context.Context, session SessionSnapshot, plan PracticePlanSnapshot, turns []TurnSnapshot) (ReportContentDraft, error) {
	if s == nil || s.registry == nil {
		return ReportContentDraft{}, fmt.Errorf("review prompt registry is not configured")
	}
	if s.ai == nil {
		return ReportContentDraft{}, fmt.Errorf("review AI client is not configured")
	}
	language := fallbackLanguage(session.Language)
	resolution, err := s.registry.ResolveActive(ctx, reportGenerateFeatureKey, language)
	if err != nil {
		return ReportContentDraft{}, fmt.Errorf("resolve report.generate: %w", err)
	}
	payload := reportCompletePayload(resolution, session, plan, turns, reportGenerateFeatureKey, aiclient.AITaskRunTaskReportGenerate)
	resp, _, err := s.ai.Complete(ctx, resolution.ModelProfileName, payload)
	if err != nil {
		return ReportContentDraft{}, fmt.Errorf("complete report.generate: %w", err)
	}
	var draft ReportContentDraft
	if err := json.Unmarshal([]byte(resp.Content), &draft); err != nil {
		return ReportContentDraft{}, fmt.Errorf("%w: parse report.generate response: %v", ErrReviewAIOutputInvalid, err)
	}
	draft.normalize()
	if draft.empty() {
		return ReportContentDraft{}, fmt.Errorf("%w: report.generate response is empty", ErrReviewAIOutputInvalid)
	}
	if containsReviewForbiddenToken(draft) {
		return ReportContentDraft{}, fmt.Errorf("%w: report.generate response contains forbidden raw text token", ErrReviewAIOutputInvalid)
	}
	return draft, nil
}

func (s *Service) assessQuestionsForAllTurns(ctx context.Context, session SessionSnapshot, plan PracticePlanSnapshot, turns []TurnSnapshot) ([]QuestionAssessmentDraft, error) {
	if s == nil || s.registry == nil {
		return nil, fmt.Errorf("review prompt registry is not configured")
	}
	if s.ai == nil {
		return nil, fmt.Errorf("review AI client is not configured")
	}
	ordered := append([]TurnSnapshot(nil), turns...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].TurnIndex < ordered[j].TurnIndex
	})
	out := make([]QuestionAssessmentDraft, 0, len(ordered))
	language := fallbackLanguage(session.Language)
	for _, turn := range ordered {
		resolution, err := s.registry.ResolveActive(ctx, reportQuestionAssessmentFeatureKey, language)
		if err != nil {
			return nil, fmt.Errorf("resolve report.question_assessment: %w", err)
		}
		payload := questionAssessmentPayload(resolution, session, plan, ordered, turn)
		resp, _, err := s.ai.Complete(ctx, resolution.ModelProfileName, payload)
		if err != nil {
			return nil, fmt.Errorf("complete report.question_assessment: %w", err)
		}
		var draft QuestionAssessmentDraft
		if err := json.Unmarshal([]byte(resp.Content), &draft); err != nil {
			return nil, fmt.Errorf("%w: parse report.question_assessment response: %v", ErrReviewAIOutputInvalid, err)
		}
		draft.TurnID = turn.ID
		draft.TurnIndex = turn.TurnIndex
		draft.QuestionIntent = turn.QuestionIntent
		draft.normalize()
		if draft.empty() {
			return nil, fmt.Errorf("%w: report.question_assessment response is empty", ErrReviewAIOutputInvalid)
		}
		if containsReviewForbiddenToken(draft) {
			return nil, fmt.Errorf("%w: report.question_assessment response contains forbidden raw text token", ErrReviewAIOutputInvalid)
		}
		out = append(out, draft)
	}
	return out, nil
}

func reportCompletePayload(resolution registry.PromptResolution, session SessionSnapshot, plan PracticePlanSnapshot, turns []TurnSnapshot, featureKey string, capability aiclient.AITaskRunCapability) aiclient.CompletePayload {
	messages := reportMessages(resolution, map[string]string{
		"{{language}}":          fallbackLanguage(session.Language),
		"{{session_metadata}}":  mustJSONString(sessionMetadata(session, plan)),
		"{{turn_summaries}}":    mustJSONString(turnSummaryPayload(turns)),
		"{{rubric_dimensions}}": "[]",
	})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: reportCallMetadata(resolution, session, featureKey, capability),
	}
}

func questionAssessmentPayload(resolution registry.PromptResolution, session SessionSnapshot, plan PracticePlanSnapshot, turns []TurnSnapshot, turn TurnSnapshot) aiclient.CompletePayload {
	messages := reportMessages(resolution, map[string]string{
		"{{language}}":         fallbackLanguage(session.Language),
		"{{session_metadata}}": mustJSONString(sessionMetadata(session, plan)),
		"{{turn_summaries}}":   mustJSONString(turnSummaryPayload(turns)),
		"{{question_context}}": sanitizePromptSegment(turn.QuestionContext),
		"{{answer_summary}}":   sanitizePromptSegment(turn.AnswerSummary),
		"{{rubric}}":           "[]",
	})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: reportCallMetadata(resolution, session, reportQuestionAssessmentFeatureKey, aiclient.AITaskRunTaskReportAssessment),
	}
}

func reportMessages(resolution registry.PromptResolution, replacements map[string]string) []aiclient.Message {
	messages := make([]aiclient.Message, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: redactPromptForbiddenLiterals(system)})
	}
	user := strings.TrimSpace(resolution.UserMessageTemplate)
	for token, value := range replacements {
		user = strings.ReplaceAll(user, token, value)
	}
	user = strings.TrimSpace(redactPromptForbiddenLiterals(user))
	if user == "" {
		user = "Return strict JSON for the interview report."
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: user})
	return messages
}

func reportCallMetadata(resolution registry.PromptResolution, session SessionSnapshot, featureKey string, capability aiclient.AITaskRunCapability) aiclient.CallMetadata {
	return aiclient.CallMetadata{
		FeatureKey:        featureKey,
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          fallbackLanguage(session.Language),
		FeatureFlag:       resolution.FeatureFlag,
		DataSourceVersion: resolution.DataSourceVersion,
		TaskRun: aiclient.AITaskRunContext{
			UserID:       session.UserID,
			Capability:   capability,
			ResourceType: aiclient.AITaskRunResourceFeedbackReport,
			ResourceID:   session.ReportID,
		},
	}
}

func sessionMetadata(session SessionSnapshot, plan PracticePlanSnapshot) map[string]any {
	return map[string]any{
		"userId":             session.UserID,
		"reportId":           session.ReportID,
		"sessionId":          session.SessionID,
		"planId":             session.PlanID,
		"targetJobId":        session.TargetJobID,
		"language":           fallbackLanguage(session.Language),
		"goal":               sanitizePromptSegment(plan.Goal),
		"mode":               sanitizePromptSegment(plan.Mode),
		"interviewerPersona": sanitizePromptSegment(plan.InterviewerPersona),
	}
}

func turnSummaryPayload(turns []TurnSnapshot) []map[string]any {
	ordered := append([]TurnSnapshot(nil), turns...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].TurnIndex < ordered[j].TurnIndex
	})
	out := make([]map[string]any, 0, len(ordered))
	for _, turn := range ordered {
		out = append(out, map[string]any{
			"turnId":          turn.ID,
			"turnIndex":       turn.TurnIndex,
			"questionIntent":  sanitizePromptSegment(turn.QuestionIntent),
			"questionContext": sanitizePromptSegment(turn.QuestionContext),
			"answerSummary":   sanitizePromptSegment(turn.AnswerSummary),
		})
	}
	return out
}

func fallbackLanguage(language string) string {
	if strings.TrimSpace(language) == "" {
		return "en"
	}
	return strings.TrimSpace(language)
}

func mustJSONString(v any) string {
	body, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(body)
}

func sanitizePromptSegment(value string) string {
	return redactPromptForbiddenLiterals(strings.TrimSpace(value))
}

func redactPromptForbiddenLiterals(value string) string {
	replacer := strings.NewReplacer(
		"question_text", "question detail",
		"answer_text", "answer summary",
		"hint_text", "hint summary",
		"prompt body", "prompt metadata",
		"response body", "response metadata",
	)
	return replacer.Replace(value)
}

func (d *ReportContentDraft) normalize() {
	if d.DimensionScores == nil {
		d.DimensionScores = []DimensionScoreDraft{}
	}
	if d.Highlights == nil {
		d.Highlights = []ReportEvidenceDraft{}
	}
	if d.Issues == nil {
		d.Issues = []ReportEvidenceDraft{}
	}
	if d.NextActions == nil {
		d.NextActions = []ReportNextActionDraft{}
	}
	if d.RetryFocusTurnIDs == nil {
		d.RetryFocusTurnIDs = []string{}
	}
}

func (d ReportContentDraft) empty() bool {
	return strings.TrimSpace(d.Summary) == "" &&
		len(d.DimensionScores) == 0 &&
		len(d.Highlights) == 0 &&
		len(d.Issues) == 0 &&
		len(d.NextActions) == 0
}

func (d *QuestionAssessmentDraft) normalize() {
	if d.DimensionResults == nil {
		d.DimensionResults = map[string]DimensionResultDraft{}
	}
	for key, value := range d.DimensionResults {
		if value.Status == "" && strings.TrimSpace(value.ScoreLevel) != "" {
			value.Status = dimensionStatusFromScoreLevel(value.ScoreLevel)
			d.DimensionResults[key] = value
		}
	}
	if d.Strengths == nil {
		d.Strengths = []string{}
	}
	if d.Gaps == nil {
		d.Gaps = []string{}
	}
	if d.ReviewStatus == "" {
		d.ReviewStatus = sharedtypes.QuestionReviewStatusOpen
	}
}

func (d QuestionAssessmentDraft) empty() bool {
	return len(d.DimensionResults) == 0 || d.OverallStatus == ""
}

func containsReviewForbiddenToken(value any) bool {
	raw := strings.ToLower(mustJSONString(value))
	for _, forbidden := range []string{"question_text", "questiontext", "answer_text", "answertext", "hint_text", "hinttext", "prompt body", "prompt_body", "response body", "response_body", "provider_secret", "providersecret"} {
		if strings.Contains(raw, forbidden) {
			return true
		}
	}
	return false
}
