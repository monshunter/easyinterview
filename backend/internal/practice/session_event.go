package practice

import (
	"context"
	"fmt"
	"strings"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	sessionEventKindAnswerSubmitted = "answer_submitted"
	sessionEventKindHintRequested   = "hint_requested"
	sessionEventKindTurnSkipped     = "turn_skipped"
	sessionEventKindSessionPaused   = "session_paused"
	sessionEventKindSessionResumed  = "session_resumed"

	assistantActionAskQuestion      = "ask_question"
	assistantActionAskFollowUp      = "ask_follow_up"
	assistantActionSessionWait      = "session_wait"
	assistantActionSessionCompleted = "session_completed"
)

type SessionEventInput struct {
	SessionID     string
	ClientEventID string
	Kind          string
	OccurredAt    time.Time
	Payload       map[string]any
}

type AssistantActionProvenance struct {
	PromptVersion     string `json:"promptVersion"`
	RubricVersion     string `json:"rubricVersion"`
	ModelID           string `json:"modelId"`
	Language          string `json:"language"`
	FeatureFlag       string `json:"featureFlag"`
	DataSourceVersion string `json:"dataSourceVersion"`
}

type AssistantActionRecord struct {
	Type          string
	TurnID        string
	QuestionText  string
	Hint          string
	SessionStatus sharedtypes.SessionStatus
	Provenance    AssistantActionProvenance
	RequiresAI    bool
}

type PracticeTurnCompletedRecord struct {
	SessionID        string
	TurnID           string
	FollowUpCount    int
	AnswerCharLength int
	CompletedAt      time.Time
}

type SessionEventOutcome struct {
	Acknowledged      bool
	AssistantAction   AssistantActionRecord
	NextSessionStatus sharedtypes.SessionStatus
	NextTurn          *TurnRecord
	OutboxRecord      *PracticeTurnCompletedRecord
	AuditMetadata     map[string]any
	Error             *ServiceError
}

type SessionEventService struct{}

func (s SessionEventService) Route(
	ctx context.Context,
	input SessionEventInput,
	session SessionRecord,
	latestTurn TurnRecord,
	plan PlanRecord,
) (SessionEventOutcome, error) {
	_ = ctx
	switch strings.TrimSpace(input.Kind) {
	case sessionEventKindAnswerSubmitted:
		return s.handleAnswerSubmitted(input, session, latestTurn, plan), nil
	case sessionEventKindHintRequested:
		return s.handleHintRequested(session, plan), nil
	case sessionEventKindTurnSkipped:
		return s.handleTurnSkipped(input, session, latestTurn, plan), nil
	case sessionEventKindSessionPaused:
		return s.handleSessionPaused(session), nil
	case sessionEventKindSessionResumed:
		return s.handleSessionResumed(session), nil
	default:
		return SessionEventOutcome{
			Acknowledged:      false,
			NextSessionStatus: session.Status,
			Error: validationError("event kind is invalid", map[string]any{
				"field": "kind",
				"kind":  strings.TrimSpace(input.Kind),
			}),
		}, nil
	}
}

func (s SessionEventService) handleAnswerSubmitted(
	input SessionEventInput,
	session SessionRecord,
	latestTurn TurnRecord,
	plan PlanRecord,
) SessionEventOutcome {
	followUpCount := latestTurn.FollowUpCount
	if followUpCount < 0 {
		followUpCount = 0
	}
	answerLength := len([]rune(payloadString(input.Payload, "answerText")))
	nextTurn := latestTurn
	nextTurn.Status = string(TurnStatusAssessed)
	nextTurn.FollowUpCount = followUpCount
	actionType := assistantActionAskQuestion
	nextStatus := sharedtypes.SessionStatusRunning
	requiresAI := false

	if session.TurnCount >= plan.QuestionBudget {
		actionType = assistantActionSessionCompleted
		nextStatus = sharedtypes.SessionStatusCompleted
		nextTurn.Status = string(TurnStatusAssessed)
	} else if followUpCount == 0 {
		actionType = assistantActionAskFollowUp
		nextTurn.Status = string(TurnStatusFollowUpRequested)
		nextTurn.FollowUpCount = 1
		requiresAI = true
	}

	var outbox *PracticeTurnCompletedRecord
	if nextTurn.Status == string(TurnStatusAssessed) {
		outbox = &PracticeTurnCompletedRecord{
			SessionID:        session.ID,
			TurnID:           latestTurn.ID,
			FollowUpCount:    nextTurn.FollowUpCount,
			AnswerCharLength: answerLength,
			CompletedAt:      input.OccurredAt.UTC(),
		}
	}
	return SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: nextStatus,
		NextTurn:          &nextTurn,
		AssistantAction:   s.assistantAction(actionType, latestTurn.ID, "", "", nextStatus, session.Language, requiresAI),
		OutboxRecord:      outbox,
		AuditMetadata: map[string]any{
			"event_kind":         sessionEventKindAnswerSubmitted,
			"answer_char_length": answerLength,
			"follow_up_count":    nextTurn.FollowUpCount,
		},
	}
}

func (s SessionEventService) handleHintRequested(session SessionRecord, plan PlanRecord) SessionEventOutcome {
	return SessionEventOutcome{
		Acknowledged:      false,
		NextSessionStatus: session.Status,
		Error: &ServiceError{
			Code:    sharederrors.CodePracticeSessionConflict,
			Message: "hints are disabled in this phase",
			Details: map[string]any{
				"policy": "hint_disabled_in_mode",
				"mode":   string(plan.Mode),
			},
		},
	}
}

func (s SessionEventService) handleTurnSkipped(
	input SessionEventInput,
	session SessionRecord,
	latestTurn TurnRecord,
	plan PlanRecord,
) SessionEventOutcome {
	nextTurn := latestTurn
	nextTurn.Status = string(TurnStatusSkipped)
	actionType := assistantActionAskQuestion
	nextStatus := sharedtypes.SessionStatusRunning
	if session.TurnCount >= plan.QuestionBudget {
		actionType = assistantActionSessionCompleted
		nextStatus = sharedtypes.SessionStatusCompleted
	}
	return SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: nextStatus,
		NextTurn:          &nextTurn,
		AssistantAction:   s.assistantAction(actionType, latestTurn.ID, "", "", nextStatus, session.Language, false),
		AuditMetadata: map[string]any{
			"event_kind":  sessionEventKindTurnSkipped,
			"turn_id":     latestTurn.ID,
			"occurred_at": input.OccurredAt.UTC().Format(time.RFC3339),
		},
	}
}

func (s SessionEventService) handleSessionPaused(session SessionRecord) SessionEventOutcome {
	return SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: sharedtypes.SessionStatusWaitingUserInput,
		AssistantAction: s.assistantAction(
			assistantActionSessionWait,
			"",
			"",
			"",
			sharedtypes.SessionStatusWaitingUserInput,
			session.Language,
			false,
		),
		AuditMetadata: map[string]any{"event_kind": sessionEventKindSessionPaused},
	}
}

func (s SessionEventService) handleSessionResumed(session SessionRecord) SessionEventOutcome {
	return SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: sharedtypes.SessionStatusRunning,
		AssistantAction: s.assistantAction(
			assistantActionSessionWait,
			"",
			"",
			"",
			sharedtypes.SessionStatusRunning,
			session.Language,
			false,
		),
		AuditMetadata: map[string]any{"event_kind": sessionEventKindSessionResumed},
	}
}

func (s SessionEventService) assistantAction(
	actionType string,
	turnID string,
	questionText string,
	hint string,
	sessionStatus sharedtypes.SessionStatus,
	language string,
	requiresAI bool,
) AssistantActionRecord {
	return AssistantActionRecord{
		Type:          actionType,
		TurnID:        turnID,
		QuestionText:  questionText,
		Hint:          hint,
		SessionStatus: sessionStatus,
		Provenance: AssistantActionProvenance{
			PromptVersion:     "not_applicable",
			RubricVersion:     "not_applicable",
			ModelID:           "model-profile:static",
			Language:          fallbackString(language, "en"),
			FeatureFlag:       "none",
			DataSourceVersion: "static",
		},
		RequiresAI: requiresAI,
	}
}

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
