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

const reportGenerateFeatureKey = string(featurekeys.ReportGenerate)

var ErrReviewAIOutputInvalid = errors.New("review: ai output invalid")

type SessionSnapshot struct {
	UserID, ReportID, SessionID, PlanID, TargetJobID, Language string
}

type PracticePlanSnapshot struct {
	ID, Goal, InterviewerPersona string
}

type MessageSnapshot struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	SeqNo   int    `json:"seqNo"`
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
	Summary                   string                  `json:"summary"`
	DimensionScores           []DimensionScoreDraft   `json:"dimension_scores"`
	Highlights                []ReportEvidenceDraft   `json:"highlights"`
	Issues                    []ReportEvidenceDraft   `json:"issues"`
	NextActions               []ReportNextActionDraft `json:"next_actions"`
	RetryFocusCompetencyCodes []string                `json:"retry_focus_competency_codes"`
}

type DimensionAssessmentDraft struct {
	Dimension  string
	Status     sharedtypes.DimensionStatus
	Confidence sharedtypes.Confidence
}

func (s *Service) generateReportContent(ctx context.Context, session SessionSnapshot, plan PracticePlanSnapshot, messages []MessageSnapshot, rubric registry.RubricSchema) (ReportContentDraft, error) {
	if s == nil || s.registry == nil || s.ai == nil {
		return ReportContentDraft{}, fmt.Errorf("review generation is not configured")
	}
	resolution, err := s.registry.ResolveActive(ctx, reportGenerateFeatureKey, fallbackLanguage(session.Language))
	if err != nil {
		return ReportContentDraft{}, fmt.Errorf("resolve report.generate: %w", err)
	}
	payload := reportCompletePayload(resolution, session, plan, messages, rubric)
	response, _, err := s.ai.Complete(ctx, resolution.ModelProfileName, payload)
	if err != nil {
		return ReportContentDraft{}, fmt.Errorf("complete report.generate: %w", err)
	}
	var draft ReportContentDraft
	if err := json.Unmarshal([]byte(response.Content), &draft); err != nil {
		return ReportContentDraft{}, fmt.Errorf("%w: parse report.generate response: %v", ErrReviewAIOutputInvalid, err)
	}
	draft.normalize()
	if draft.empty() {
		return ReportContentDraft{}, fmt.Errorf("%w: report.generate response is empty", ErrReviewAIOutputInvalid)
	}
	if err := validateReportContent(draft); err != nil {
		return ReportContentDraft{}, fmt.Errorf("%w: %v", ErrReviewAIOutputInvalid, err)
	}
	return draft, nil
}

func validateReportContent(content ReportContentDraft) error {
	if len(content.DimensionScores) == 0 {
		return fmt.Errorf("candidate dimension scores are required")
	}

	seen := make(map[string]struct{}, len(content.DimensionScores))
	for _, score := range content.DimensionScores {
		name := strings.TrimSpace(score.Name)
		if name == "" {
			return fmt.Errorf("dimension score name is required")
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("dimension score %q is duplicated", name)
		}
		if score.Score < 1 || score.Score > 5 {
			return fmt.Errorf("dimension score %q must be between 1.0 and 5.0", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func reportCompletePayload(resolution registry.PromptResolution, session SessionSnapshot, plan PracticePlanSnapshot, messages []MessageSnapshot, rubric registry.RubricSchema) aiclient.CompletePayload {
	ordered := append([]MessageSnapshot(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].SeqNo < ordered[j].SeqNo })
	replacements := map[string]string{
		"{{language}}": fallbackLanguage(session.Language),
		"{{session_metadata}}": mustJSONString(map[string]any{
			"reportId": session.ReportID, "sessionId": session.SessionID, "planId": session.PlanID,
			"targetJobId": session.TargetJobID, "goal": plan.Goal, "interviewerPersona": plan.InterviewerPersona,
		}),
		"{{conversation_messages}}": mustJSONString(ordered),
		"{{rubric_dimensions}}":     mustJSONString(rubricPromptPayload(rubric)),
	}
	user := strings.TrimSpace(resolution.UserMessageTemplate)
	for token, value := range replacements {
		user = strings.ReplaceAll(user, token, value)
	}
	messagesPayload := []aiclient.Message{}
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messagesPayload = append(messagesPayload, aiclient.Message{Role: "system", Content: system})
	}
	messagesPayload = append(messagesPayload, aiclient.Message{Role: "user", Content: user})
	metadata := aiclient.CallMetadata{
		FeatureKey: reportGenerateFeatureKey, PromptVersion: resolution.PromptVersion, RubricVersion: resolution.RubricVersion,
		Language: fallbackLanguage(session.Language), FeatureFlag: resolution.FeatureFlag, DataSourceVersion: resolution.DataSourceVersion,
		TaskRun: aiclient.AITaskRunContext{UserID: session.UserID, Capability: aiclient.AITaskRunTaskReportGenerate,
			ResourceType: aiclient.AITaskRunResourceFeedbackReport, ResourceID: session.ReportID},
	}
	if resolution.OutputSchema != nil {
		metadata.OutputSchema = *resolution.OutputSchema
	}
	return aiclient.CompletePayload{Messages: messagesPayload, Metadata: metadata}
}

func rubricPromptPayload(rubric registry.RubricSchema) []map[string]any {
	out := make([]map[string]any, 0, len(rubric.Dimensions))
	for _, dimension := range rubric.Dimensions {
		out = append(out, map[string]any{"name": dimension.Name, "weight": dimension.Weight, "description": dimension.Description})
	}
	return out
}

func dimensionAssessments(content ReportContentDraft) []DimensionAssessmentDraft {
	out := make([]DimensionAssessmentDraft, 0, len(content.DimensionScores))
	for _, score := range content.DimensionScores {
		status := sharedtypes.DimensionStatusNeedsWork
		if score.Score >= 4 {
			status = sharedtypes.DimensionStatusStrong
		} else if score.Score >= 3 {
			status = sharedtypes.DimensionStatusMeetsBar
		}
		confidence := sharedtypes.ConfidenceMedium
		if len(score.SupportingObservations) >= 2 {
			confidence = sharedtypes.ConfidenceHigh
		}
		out = append(out, DimensionAssessmentDraft{Dimension: score.Name, Status: status, Confidence: confidence})
	}
	return out
}

func readinessFromContent(content ReportContentDraft) sharedtypes.ReadinessTier {
	if len(content.DimensionScores) == 0 {
		return sharedtypes.ReadinessTierNotReady
	}
	total := 0.0
	for _, score := range content.DimensionScores {
		total += score.Score
	}
	average := total / float64(len(content.DimensionScores))
	switch {
	case average >= 4:
		return sharedtypes.ReadinessTierWellPrepared
	case average >= 3:
		return sharedtypes.ReadinessTierBasicallyReady
	case average >= 2:
		return sharedtypes.ReadinessTierNeedsPractice
	default:
		return sharedtypes.ReadinessTierNotReady
	}
}

func (d *ReportContentDraft) normalize() {
	d.Summary = strings.TrimSpace(d.Summary)
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
	if d.RetryFocusCompetencyCodes == nil {
		seen := map[string]struct{}{}
		for _, issue := range d.Issues {
			code := strings.TrimSpace(issue.Dimension)
			if code != "" {
				if _, ok := seen[code]; !ok {
					seen[code] = struct{}{}
					d.RetryFocusCompetencyCodes = append(d.RetryFocusCompetencyCodes, code)
				}
			}
		}
	}
}

func (d ReportContentDraft) empty() bool {
	return d.Summary == "" && len(d.DimensionScores) == 0 && len(d.Highlights) == 0 && len(d.Issues) == 0
}

func fallbackLanguage(language string) string {
	if strings.TrimSpace(language) == "" {
		return "en"
	}
	return strings.TrimSpace(language)
}
func mustJSONString(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
