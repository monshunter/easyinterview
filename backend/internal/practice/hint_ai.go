package practice

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const hintFeatureKey = "practice.turn.lightweight_observe"

type hintAIResponse struct {
	Cue           string `json:"cue"`
	AnswerSummary string `json:"answerSummary"`
}

type turnObservation struct {
	Hint          string
	AnswerSummary string
}

func (s *Service) applyHintAI(ctx context.Context, reservation SessionEventReservation, payload map[string]any, outcome *SessionEventOutcome) {
	if outcome == nil {
		return
	}
	outcome.AssistantAction.TurnID = reservation.LatestTurn.ID
	outcome.AssistantAction.SessionStatus = sharedtypes.SessionStatusRunning
	if s.registry == nil || s.ai == nil {
		s.degradeHint(ctx, reservation, outcome, sharederrors.CodeAiProviderConfigInvalid, registry.PromptResolution{})
		return
	}
	resolution, err := s.registry.ResolveActive(ctx, hintFeatureKey, reservation.Session.Language)
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if !(stderrs.Is(err, registry.ErrPromptUnsupported) || stderrs.Is(err, registry.ErrLanguageUnsupported)) {
			if mapped, ok := aiErrorCode(err); ok {
				code = mapped
			}
		}
		s.writeHintTaskRun(ctx, reservation, resolution, code)
		s.degradeHint(ctx, reservation, outcome, code, resolution)
		return
	}
	resp, meta, err := s.ai.Complete(ctx, resolution.ModelProfileName, hintPayload(resolution, reservation, payload))
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if mapped, ok := aiErrorCode(err); ok {
			code = mapped
		}
		if meta.ErrorCode == "" {
			meta.ErrorCode = code
		}
		s.degradeHint(ctx, reservation, outcome, code, resolution)
		return
	}
	hint, err := parseHint(resp.Content)
	if err != nil || strings.TrimSpace(hint) == "" {
		s.writeHintTaskRun(ctx, reservation, resolution, sharederrors.CodeAiOutputInvalid)
		s.degradeHint(ctx, reservation, outcome, sharederrors.CodeAiOutputInvalid, resolution)
		return
	}
	modelID := strings.TrimSpace(meta.ModelID)
	if modelID == "" {
		modelID = "model-profile:" + strings.TrimSpace(resolution.ModelProfileName)
	}
	outcome.AssistantAction.Hint = hint
	outcome.AssistantAction.Provenance = AssistantActionProvenance{
		PromptVersion:     fallbackString(resolution.PromptVersion, "not_applicable"),
		RubricVersion:     "not_applicable",
		ModelID:           fallbackString(modelID, "model-profile:unknown"),
		Language:          fallbackString(reservation.Session.Language, "en"),
		FeatureFlag:       fallbackString(resolution.FeatureFlag, "none"),
		DataSourceVersion: fallbackString(resolution.DataSourceVersion, "not_applicable"),
	}
	outcome.AssistantAction.RequiresAI = false
}

func (s *Service) applyAnswerObservationAI(ctx context.Context, reservation SessionEventReservation, payload map[string]any, outcome *SessionEventOutcome) {
	if outcome == nil || strings.TrimSpace(payloadString(payload, "answerText")) == "" {
		return
	}
	if s.registry == nil || s.ai == nil {
		outcome.AnswerSummary = fallbackAnswerSummary(payload)
		return
	}
	resolution, err := s.registry.ResolveActive(ctx, hintFeatureKey, reservation.Session.Language)
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if !(stderrs.Is(err, registry.ErrPromptUnsupported) || stderrs.Is(err, registry.ErrLanguageUnsupported)) {
			if mapped, ok := aiErrorCode(err); ok {
				code = mapped
			}
		}
		s.writeHintTaskRun(ctx, reservation, resolution, code)
		outcome.AnswerSummary = fallbackAnswerSummary(payload)
		markAnswerSummaryDegrade(outcome, code)
		return
	}
	resp, meta, err := s.ai.Complete(ctx, resolution.ModelProfileName, hintPayload(resolution, reservation, payload))
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if mapped, ok := aiErrorCode(err); ok {
			code = mapped
		}
		if meta.ErrorCode == "" {
			meta.ErrorCode = code
		}
		outcome.AnswerSummary = fallbackAnswerSummary(payload)
		markAnswerSummaryDegrade(outcome, code)
		return
	}
	observation, err := parseTurnObservation(resp.Content)
	if err != nil || strings.TrimSpace(observation.AnswerSummary) == "" {
		s.writeHintTaskRun(ctx, reservation, resolution, sharederrors.CodeAiOutputInvalid)
		outcome.AnswerSummary = fallbackAnswerSummary(payload)
		markAnswerSummaryDegrade(outcome, sharederrors.CodeAiOutputInvalid)
		return
	}
	outcome.AnswerSummary = strings.TrimSpace(observation.AnswerSummary)
}

func hintPayload(resolution registry.PromptResolution, reservation SessionEventReservation, eventPayload map[string]any) aiclient.CompletePayload {
	userContent := renderHintTemplate(resolution.UserMessageTemplate, reservation, eventPayload)
	if userContent == "" {
		userContent = "Generate one concise interview hint for the current turn."
	}
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: attachOutputSchema(aiclient.CallMetadata{
			FeatureKey:        hintFeatureKey,
			PromptVersion:     resolution.PromptVersion,
			RubricVersion:     "not_applicable",
			Language:          reservation.Session.Language,
			FeatureFlag:       resolution.FeatureFlag,
			DataSourceVersion: resolution.DataSourceVersion,
			TaskRun: aiclient.AITaskRunContext{
				UserID:       reservation.UserID,
				Capability:   aiclient.AITaskRunTaskHintGenerate,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   reservation.Plan.TargetJobID,
			},
		}, resolution),
	}
}

func renderHintTemplate(template string, reservation SessionEventReservation, eventPayload map[string]any) string {
	content := strings.TrimSpace(template)
	if content == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"{{language}}", fallbackString(reservation.Session.Language, "en"),
		"{{question}}", fallbackString(reservation.LatestTurn.QuestionText, "current question unavailable"),
		"{{partial_answer}}", fallbackString(payloadString(eventPayload, "answerText"), "not provided"),
		"{{elapsed_seconds}}", fallbackString(firstPayloadString(eventPayload, "elapsedSeconds", "elapsed_seconds"), "0"),
		"{{practice_goal}}", fallbackString(string(reservation.Plan.Goal), string(sharedtypes.PracticeGoalBaseline)),
		"{{practice_mode}}", fallbackString(string(reservation.Plan.Mode), string(sharedtypes.PracticeModeAssisted)),
	)
	return strings.TrimSpace(replacer.Replace(content))
}

func firstPayloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(payloadString(payload, key)); value != "" {
			return value
		}
	}
	return ""
}

func parseHint(content string) (string, error) {
	observation, err := parseTurnObservation(content)
	if err != nil {
		return "", err
	}
	if observation.Hint == "" {
		return "", fmt.Errorf("parse hint response: hint is empty")
	}
	return observation.Hint, nil
}

func parseTurnObservation(content string) (turnObservation, error) {
	var decoded hintAIResponse
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		return turnObservation{}, fmt.Errorf("parse turn observation response: %w", err)
	}
	hint := strings.TrimSpace(decoded.Cue)
	summary := strings.TrimSpace(decoded.AnswerSummary)
	return turnObservation{Hint: hint, AnswerSummary: summary}, nil
}

func fallbackAnswerSummary(payload map[string]any) string {
	answerLength := len([]rune(strings.TrimSpace(payloadString(payload, "answerText"))))
	if answerLength == 0 {
		return ""
	}
	return fmt.Sprintf("Candidate submitted an answer (%d characters); AI-generated answer summary was unavailable.", answerLength)
}

func markAnswerSummaryDegrade(outcome *SessionEventOutcome, code string) {
	if outcome == nil {
		return
	}
	if outcome.AuditMetadata == nil {
		outcome.AuditMetadata = map[string]any{}
	}
	outcome.AuditMetadata["answer_summary_degrade_reason"] = code
}

func (s *Service) degradeHint(ctx context.Context, reservation SessionEventReservation, outcome *SessionEventOutcome, code string, resolution registry.PromptResolution) {
	if code == "" {
		code = sharederrors.CodeAiProviderConfigInvalid
	}
	outcome.AssistantAction = (SessionEventService{}).assistantAction(
		assistantActionSessionWait,
		reservation.LatestTurn.ID,
		"",
		"",
		sharedtypes.SessionStatusRunning,
		reservation.Session.Language,
		false,
	)
	outcome.NextSessionStatus = sharedtypes.SessionStatusRunning
	if outcome.AuditMetadata == nil {
		outcome.AuditMetadata = map[string]any{}
	}
	outcome.AuditMetadata["event_kind"] = sessionEventKindHintRequested
	outcome.AuditMetadata["mode"] = string(sharedtypes.PracticeModeAssisted)
	outcome.AuditMetadata["hint_degrade_reason"] = code
	_ = ctx
	_ = resolution
}

func (s *Service) writeHintTaskRun(ctx context.Context, reservation SessionEventReservation, resolution registry.PromptResolution, code string) {
	if s == nil || s.aiTaskRuns == nil {
		return
	}
	now := s.now().UTC()
	meta := aiclient.AICallMeta{
		Provider:          "not_applicable",
		ModelFamily:       "not_applicable",
		ModelID:           fallbackString(resolution.ModelProfileName, "model-profile:unknown"),
		PromptVersion:     fallbackString(resolution.PromptVersion, "not_applicable"),
		RubricVersion:     "not_applicable",
		ModelProfileName:  fallbackString(resolution.ModelProfileName, "unknown"),
		FeatureKey:        hintFeatureKey,
		FeatureFlag:       fallbackString(resolution.FeatureFlag, "none"),
		DataSourceVersion: fallbackString(resolution.DataSourceVersion, "not_applicable"),
		Language:          fallbackString(reservation.Session.Language, "en"),
		ValidationStatus:  aiclient.ValidationStatusInvalid,
		ErrorCode:         code,
	}
	row, err := observability.AITaskRunRowFromMeta(meta, aiclient.AITaskRunContext{
		UserID:       reservation.UserID,
		Capability:   aiclient.AITaskRunTaskHintGenerate,
		ResourceType: aiclient.AITaskRunResourceTargetJob,
		ResourceID:   reservation.Plan.TargetJobID,
	}, aiclient.AuditMetadata{}, now, now, fmt.Errorf("%s", code))
	if err != nil {
		return
	}
	_ = s.aiTaskRuns.WriteAITaskRun(ctx, row)
}
