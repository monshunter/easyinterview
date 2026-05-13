package practice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const followUpFeatureKey = "practice.session.follow_up"

type AppendSessionEventRequest struct {
	UserID        string
	SessionID     string
	ClientEventID string
	Kind          string
	OccurredAt    time.Time
	Payload       map[string]any
}

type SessionEventReservationInput struct {
	EventID            string
	UserID             string
	SessionID          string
	ClientEventID      string
	Kind               string
	CurrentTurnID      string
	RequestFingerprint string
	Now                time.Time
}

type SessionEventReservation struct {
	UserID       string
	Session      SessionRecord
	Plan         PlanRecord
	LatestTurn   TurnRecord
	ReplayResult *AppendSessionEventResult
}

type AppendSessionEventStoreInput struct {
	EventID            string
	OutboxEventID      string
	UserID             string
	SessionID          string
	ClientEventID      string
	Kind               string
	OccurredAt         time.Time
	RequestFingerprint string
	RequestPayload     map[string]any
	Outcome            SessionEventOutcome
	NextQuestion       *TurnRecord
}

type AppendSessionEventResult struct {
	Acknowledged    bool
	Session         SessionRecord
	AssistantAction AssistantActionRecord
	Replay          bool
}

func (s *Service) AppendSessionEvent(ctx context.Context, in AppendSessionEventRequest) (AppendSessionEventResult, error) {
	if s == nil || s.store == nil {
		return AppendSessionEventResult{}, fmt.Errorf("practice service is not initialised")
	}
	userID := strings.TrimSpace(in.UserID)
	sessionID := strings.TrimSpace(in.SessionID)
	clientEventID := strings.TrimSpace(in.ClientEventID)
	kind := strings.TrimSpace(in.Kind)
	if userID == "" {
		return AppendSessionEventResult{}, fmt.Errorf("userId is required")
	}
	if sessionID == "" {
		return AppendSessionEventResult{}, sessionNotFoundError()
	}
	if clientEventID == "" {
		return AppendSessionEventResult{}, validationError("clientEventId is required", map[string]any{"field": "clientEventId"})
	}
	if kind == "" {
		return AppendSessionEventResult{}, validationError("kind is required", map[string]any{"field": "kind"})
	}
	occurredAt := in.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = s.now().UTC()
	}
	payload := clonePayload(in.Payload)
	fingerprint, err := sessionEventFingerprint(kind, occurredAt, payload)
	if err != nil {
		return AppendSessionEventResult{}, err
	}
	currentTurnID := ""
	if requiresCurrentTurn(kind) {
		currentTurnID = strings.TrimSpace(payloadString(payload, "turnId"))
		if currentTurnID == "" {
			return AppendSessionEventResult{}, validationError("turnId is required", map[string]any{"field": "payload.turnId"})
		}
	}
	if kind == sessionEventKindAnswerSubmitted && strings.TrimSpace(payloadString(payload, "answerText")) == "" {
		return AppendSessionEventResult{}, validationError("answerText is required", map[string]any{"field": "payload.answerText"})
	}
	eventID := s.newID()
	reservation, err := s.store.ReserveSessionEvent(ctx, SessionEventReservationInput{
		EventID:            eventID,
		UserID:             userID,
		SessionID:          sessionID,
		ClientEventID:      clientEventID,
		Kind:               kind,
		CurrentTurnID:      currentTurnID,
		RequestFingerprint: fingerprint,
		Now:                s.now().UTC(),
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return AppendSessionEventResult{}, sessionNotFoundError()
	}
	if stderrs.Is(err, ErrClientEventMismatch) {
		return AppendSessionEventResult{}, clientEventIdMismatchToConflict()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return AppendSessionEventResult{}, sessionConflictError()
	}
	if err != nil {
		return AppendSessionEventResult{}, err
	}
	if reservation.ReplayResult != nil {
		replay := *reservation.ReplayResult
		replay.Replay = true
		return replay, nil
	}
	if currentTurnID != "" {
		if currentTurnID != reservation.LatestTurn.ID {
			return AppendSessionEventResult{}, sessionConflictError()
		}
	}

	router := SessionEventService{}
	outcome, err := router.Route(ctx, SessionEventInput{
		SessionID:     sessionID,
		ClientEventID: clientEventID,
		Kind:          kind,
		OccurredAt:    occurredAt,
		Payload:       payload,
	}, reservation.Session, reservation.LatestTurn, reservation.Plan)
	if err != nil {
		return AppendSessionEventResult{}, err
	}
	if outcome.Error != nil {
		return AppendSessionEventResult{}, outcome.Error
	}
	if outcome.AssistantAction.RequiresAI {
		s.applyFollowUpAI(ctx, reservation, payload, &outcome)
	}
	nextQuestion := s.prepareNextQuestion(reservation, payload, &outcome)
	result, err := s.store.AppendSessionEvent(ctx, AppendSessionEventStoreInput{
		EventID:            eventID,
		OutboxEventID:      s.newID(),
		UserID:             userID,
		SessionID:          sessionID,
		ClientEventID:      clientEventID,
		Kind:               kind,
		OccurredAt:         occurredAt,
		RequestFingerprint: fingerprint,
		RequestPayload:     payload,
		Outcome:            outcome,
		NextQuestion:       nextQuestion,
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return AppendSessionEventResult{}, sessionNotFoundError()
	}
	if stderrs.Is(err, ErrClientEventMismatch) {
		return AppendSessionEventResult{}, clientEventIdMismatchToConflict()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return AppendSessionEventResult{}, sessionConflictError()
	}
	if err != nil {
		return AppendSessionEventResult{}, err
	}
	return result, nil
}

func requiresCurrentTurn(kind string) bool {
	switch strings.TrimSpace(kind) {
	case sessionEventKindAnswerSubmitted, sessionEventKindHintRequested, sessionEventKindTurnSkipped:
		return true
	default:
		return false
	}
}

func (s *Service) prepareNextQuestion(reservation SessionEventReservation, payload map[string]any, outcome *SessionEventOutcome) *TurnRecord {
	if outcome == nil || outcome.AssistantAction.Type != assistantActionAskQuestion || outcome.NextSessionStatus != sharedtypes.SessionStatusRunning {
		return nil
	}
	turnID := s.newID()
	questionText := payloadString(payload, "nextQuestionText")
	if strings.TrimSpace(questionText) == "" {
		questionText = "Please describe another example that is relevant to the target role."
	}
	questionIntent := payloadString(payload, "nextQuestionIntent")
	if strings.TrimSpace(questionIntent) == "" {
		questionIntent = "behavioral.depth"
	}
	next := &TurnRecord{
		ID:             turnID,
		TurnIndex:      reservation.LatestTurn.TurnIndex + 1,
		QuestionText:   questionText,
		QuestionIntent: questionIntent,
		Status:         string(TurnStatusAsked),
		AskedAt:        s.now().UTC(),
	}
	outcome.AssistantAction.TurnID = turnID
	outcome.AssistantAction.QuestionText = questionText
	return next
}

func (s *Service) applyFollowUpAI(ctx context.Context, reservation SessionEventReservation, payload map[string]any, outcome *SessionEventOutcome) {
	if outcome == nil {
		return
	}
	if s.registry == nil || s.ai == nil {
		outcome.AssistantAction = fallbackFollowUpAction(reservation, outcome.AssistantAction)
		return
	}
	resolution, err := s.registry.ResolveActive(ctx, followUpFeatureKey, reservation.Session.Language)
	if err != nil {
		outcome.AssistantAction = fallbackFollowUpAction(reservation, outcome.AssistantAction)
		return
	}
	resp, meta, err := s.ai.Complete(ctx, resolution.ModelProfileName, followUpPayload(resolution, reservation, payload))
	if err != nil {
		outcome.AssistantAction = fallbackFollowUpAction(reservation, outcome.AssistantAction)
		return
	}
	question, err := parseFirstQuestion(resp.Content)
	if err != nil {
		outcome.AssistantAction = fallbackFollowUpAction(reservation, outcome.AssistantAction)
		return
	}
	modelID := strings.TrimSpace(meta.ModelID)
	if modelID == "" {
		modelID = "model-profile:" + strings.TrimSpace(resolution.ModelProfileName)
	}
	outcome.AssistantAction.QuestionText = question.Text
	outcome.AssistantAction.Provenance = AssistantActionProvenance{
		PromptVersion:     fallbackString(resolution.PromptVersion, "not_applicable"),
		RubricVersion:     fallbackString(resolution.RubricVersion, "not_applicable"),
		ModelID:           fallbackString(modelID, "model-profile:unknown"),
		Language:          fallbackString(reservation.Session.Language, "en"),
		FeatureFlag:       fallbackString(resolution.FeatureFlag, "none"),
		DataSourceVersion: fallbackString(resolution.DataSourceVersion, "not_applicable"),
	}
	outcome.AssistantAction.RequiresAI = false
}

func fallbackFollowUpAction(reservation SessionEventReservation, action AssistantActionRecord) AssistantActionRecord {
	action.Type = assistantActionAskQuestion
	action.QuestionText = "Please continue with the next practice question."
	action.SessionStatus = sharedtypes.SessionStatusRunning
	action.RequiresAI = false
	action.Provenance = (SessionEventService{}).assistantAction(
		assistantActionAskQuestion,
		action.TurnID,
		action.QuestionText,
		"",
		sharedtypes.SessionStatusRunning,
		reservation.Session.Language,
		false,
	).Provenance
	return action
}

func followUpPayload(resolution registry.PromptResolution, reservation SessionEventReservation, eventPayload map[string]any) aiclient.CompletePayload {
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
		userContent = "Generate a concise follow-up question for the candidate's latest answer."
	}
	if answer := payloadString(eventPayload, "answerText"); strings.TrimSpace(answer) != "" {
		userContent += "\nAnswer summary source length: " + fmt.Sprintf("%d", len([]rune(answer)))
	}
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
			FeatureKey:        followUpFeatureKey,
			PromptVersion:     resolution.PromptVersion,
			RubricVersion:     resolution.RubricVersion,
			Language:          reservation.Session.Language,
			FeatureFlag:       resolution.FeatureFlag,
			DataSourceVersion: resolution.DataSourceVersion,
			TaskRun: aiclient.AITaskRunContext{
				UserID:       reservation.UserID,
				Capability:   aiclient.AITaskRunTaskFollowupGenerate,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   reservation.Plan.TargetJobID,
			},
		},
	}
}

func sessionEventFingerprint(kind string, occurredAt time.Time, payload map[string]any) (string, error) {
	raw, err := json.Marshal(struct {
		Kind       string         `json:"kind"`
		OccurredAt string         `json:"occurredAt"`
		Payload    map[string]any `json:"payload"`
	}{
		Kind:       strings.TrimSpace(kind),
		OccurredAt: occurredAt.UTC().Format(time.RFC3339Nano),
		Payload:    clonePayload(payload),
	})
	if err != nil {
		return "", fmt.Errorf("marshal session event fingerprint: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func clonePayload(payload map[string]any) map[string]any {
	if payload == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(payload))
	for key, value := range payload {
		out[key] = value
	}
	return out
}

func clientEventIdMismatchToConflict() *ServiceError {
	return &ServiceError{
		Code:    sharederrors.CodePracticeSessionConflict,
		Message: "clientEventId was already used with a different payload",
		Details: map[string]any{
			"policy": "client_event_payload_mismatch",
		},
	}
}
