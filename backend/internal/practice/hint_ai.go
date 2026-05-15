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
	Hint string `json:"hint"`
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

func hintPayload(resolution registry.PromptResolution, reservation SessionEventReservation, eventPayload map[string]any) aiclient.CompletePayload {
	userContent := renderFirstQuestionTemplate(resolution.UserMessageTemplate, SessionReservation{
		UserID:             reservation.Session.ID,
		SessionID:          reservation.Session.ID,
		PlanID:             reservation.Plan.ID,
		TargetJobID:        reservation.Plan.TargetJobID,
		Goal:               reservation.Plan.Goal,
		Mode:               reservation.Plan.Mode,
		InterviewerPersona: reservation.Plan.InterviewerPersona,
		Language:           reservation.Session.Language,
	})
	if userContent == "" {
		userContent = "Generate one concise interview hint for the current turn."
	}
	if question := strings.TrimSpace(reservation.LatestTurn.QuestionText); question != "" {
		userContent += "\nCurrent question length: " + fmt.Sprintf("%d", len([]rune(question)))
	}
	if answer := payloadString(eventPayload, "answerText"); strings.TrimSpace(answer) != "" {
		userContent += "\nCurrent answer length: " + fmt.Sprintf("%d", len([]rune(answer)))
	}
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
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
		},
	}
}

func parseHint(content string) (string, error) {
	var decoded hintAIResponse
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		return "", fmt.Errorf("parse hint response: %w", err)
	}
	hint := strings.TrimSpace(decoded.Hint)
	if hint == "" {
		return "", fmt.Errorf("parse hint response: hint is empty")
	}
	return hint, nil
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
