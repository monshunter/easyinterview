package practice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	sessionEventKindAnswerSubmitted  = "answer_submitted"
	sessionEventKindHintRequested    = "hint_requested"
	sessionEventKindSessionPaused    = "session_paused"
	sessionEventKindSessionResumed   = "session_resumed"
	sessionEventKindTTSChunkStarted  = "tts_chunk_started"
	sessionEventKindTTSChunkPlayed   = "tts_chunk_played"
	sessionEventKindBargeInDetected  = "barge_in_detected"
	sessionEventKindContextCommitted = "assistant_context_committed"

	assistantActionAskQuestion      = "ask_question"
	assistantActionAskFollowUp      = "ask_follow_up"
	assistantActionShowHint         = "show_hint"
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
	Type           string
	TurnID         string
	QuestionText   string
	QuestionIntent string
	Hint           string
	SessionStatus  sharedtypes.SessionStatus
	Provenance     AssistantActionProvenance
	RequiresAI     bool
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
	AnswerSummary     string
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
	kind := strings.TrimSpace(input.Kind)
	if isClosedSessionStatus(session.Status) {
		return sessionEventConflictOutcome(session.Status, "session_event_closed", map[string]any{
			"kind": kind,
		}), nil
	}
	if requiresCurrentTurn(kind) && isClosedTurnStatus(latestTurn.Status) {
		return sessionEventConflictOutcome(session.Status, "turn_event_closed", map[string]any{
			"kind":       kind,
			"turnStatus": latestTurn.Status,
		}), nil
	}
	switch kind {
	case sessionEventKindAnswerSubmitted:
		return s.handleAnswerSubmitted(input, session, latestTurn, plan), nil
	case sessionEventKindHintRequested:
		return s.handleHintRequested(session, plan), nil
	case sessionEventKindSessionPaused:
		return s.handleSessionPaused(session), nil
	case sessionEventKindSessionResumed:
		return s.handleSessionResumed(session), nil
	case sessionEventKindTTSChunkStarted,
		sessionEventKindTTSChunkPlayed,
		sessionEventKindBargeInDetected,
		sessionEventKindContextCommitted:
		return s.handleVoicePlaybackEvent(input, session, latestTurn), nil
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

func sessionEventConflictOutcome(status sharedtypes.SessionStatus, policy string, details map[string]any) SessionEventOutcome {
	if details == nil {
		details = map[string]any{}
	}
	details["policy"] = policy
	details["sessionStatus"] = string(status)
	return SessionEventOutcome{
		Acknowledged:      false,
		NextSessionStatus: status,
		Error: &ServiceError{
			Code:    sharederrors.CodePracticeSessionConflict,
			Message: "practice session event is not allowed in the current state",
			Details: details,
		},
	}
}

func isClosedSessionStatus(status sharedtypes.SessionStatus) bool {
	switch status {
	case sharedtypes.SessionStatusCompleting,
		sharedtypes.SessionStatusCompleted,
		sharedtypes.SessionStatusFailed,
		sharedtypes.SessionStatusCancelled:
		return true
	default:
		return false
	}
}

func isClosedTurnStatus(status string) bool {
	switch TurnStatus(strings.TrimSpace(status)) {
	case TurnStatusAnswered, TurnStatusAssessed:
		return true
	default:
		return false
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
	requiresAI := true

	if session.TurnCount >= plan.QuestionBudget {
		actionType = assistantActionSessionCompleted
		nextStatus = sharedtypes.SessionStatusCompleted
		nextTurn.Status = string(TurnStatusAssessed)
		requiresAI = false
	} else if followUpCount == 0 {
		actionType = assistantActionAskFollowUp
		nextTurn.Status = string(TurnStatusFollowUpRequested)
		nextTurn.FollowUpCount = 1
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
	mode := string(plan.Mode)
	if strings.TrimSpace(mode) == "" {
		mode = string(sharedtypes.PracticeModeAssisted)
	}
	return SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: session.Status,
		AssistantAction: AssistantActionRecord{
			Type:          assistantActionShowHint,
			SessionStatus: session.Status,
			RequiresAI:    true,
		},
		AuditMetadata: map[string]any{
			"event_kind": sessionEventKindHintRequested,
			"mode":       mode,
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

func (s SessionEventService) handleVoicePlaybackEvent(input SessionEventInput, session SessionRecord, latestTurn TurnRecord) SessionEventOutcome {
	kind := strings.TrimSpace(input.Kind)
	if err := validateVoicePlaybackPayload(kind, input.Payload); err != nil {
		return SessionEventOutcome{
			Acknowledged:      false,
			NextSessionStatus: session.Status,
			Error:             err,
		}
	}
	audit := map[string]any{
		"event_kind":         kind,
		"voice_turn_id":      strings.TrimSpace(payloadString(input.Payload, "voiceTurnId")),
		"chunk_id":           strings.TrimSpace(payloadString(input.Payload, "chunkId")),
		"playback_offset_ms": payloadInt(input.Payload, "playbackOffsetMs"),
	}
	switch kind {
	case sessionEventKindTTSChunkPlayed:
		audit["played_text_hash"] = strings.TrimSpace(payloadString(input.Payload, "playedTextHash"))
		audit["played_text_length"] = payloadInt(input.Payload, "playedTextLength")
	case sessionEventKindBargeInDetected:
		audit["user_speech_started_at"] = strings.TrimSpace(payloadString(input.Payload, "userSpeechStartedAt"))
	case sessionEventKindContextCommitted:
		audit["committed_text_hash"] = strings.TrimSpace(payloadString(input.Payload, "committedTextHash"))
		audit["committed_text_length"] = payloadInt(input.Payload, "committedTextLength")
	}
	return SessionEventOutcome{
		Acknowledged:      true,
		NextSessionStatus: session.Status,
		AssistantAction: s.assistantAction(
			assistantActionSessionWait,
			latestTurn.ID,
			"",
			"",
			session.Status,
			session.Language,
			false,
		),
		AuditMetadata: audit,
	}
}

func validateVoicePlaybackPayload(kind string, payload map[string]any) *ServiceError {
	for _, field := range []string{"voiceTurnId", "chunkId"} {
		value := strings.TrimSpace(payloadString(payload, field))
		if value == "" {
			return validationError("voice playback event payload is missing required field", map[string]any{"field": "payload." + field})
		}
		if !validVoicePlaybackToken(value) {
			return validationError("voice playback event payload field is invalid", map[string]any{"field": "payload." + field})
		}
	}
	if value, ok := payloadIntOK(payload, "playbackOffsetMs"); !ok || value < 0 {
		return validationError("voice playback event payload field is invalid", map[string]any{"field": "payload.playbackOffsetMs"})
	}
	switch kind {
	case sessionEventKindTTSChunkPlayed:
		if !validSHA256Digest(payloadString(payload, "playedTextHash")) {
			return validationError("voice playback event payload field is invalid", map[string]any{"field": "payload.playedTextHash"})
		}
		if value, ok := payloadIntOK(payload, "playedTextLength"); !ok || value < 1 {
			return validationError("voice playback event payload is missing required field", map[string]any{"field": "payload.playedTextLength"})
		}
	case sessionEventKindBargeInDetected:
		startedAt := strings.TrimSpace(payloadString(payload, "userSpeechStartedAt"))
		if _, err := time.Parse(time.RFC3339, startedAt); err != nil {
			return validationError("voice playback event payload field is invalid", map[string]any{"field": "payload.userSpeechStartedAt"})
		}
	case sessionEventKindContextCommitted:
		if !validSHA256Digest(payloadString(payload, "committedTextHash")) {
			return validationError("voice playback event payload field is invalid", map[string]any{"field": "payload.committedTextHash"})
		}
		if value, ok := payloadIntOK(payload, "committedTextLength"); !ok || value < 1 {
			return validationError("voice playback event payload is missing required field", map[string]any{"field": "payload.committedTextLength"})
		}
	}
	return nil
}

func validVoicePlaybackToken(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 128 || !isASCIILetterOrDigit(value[0]) {
		return false
	}
	for i := 1; i < len(value); i++ {
		if !isASCIILetterOrDigit(value[i]) && value[i] != '-' && value[i] != '_' && value[i] != '.' {
			return false
		}
	}
	return true
}

func isASCIILetterOrDigit(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func validSHA256Digest(value string) bool {
	digest := strings.TrimSpace(value)
	digest = strings.TrimPrefix(digest, "sha256:")
	if len(digest) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil
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

func payloadInt(payload map[string]any, key string) int64 {
	value, _ := payloadIntOK(payload, key)
	return value
}

func payloadIntOK(payload map[string]any, key string) (int64, bool) {
	if payload == nil {
		return 0, false
	}
	value, ok := payload[key]
	if !ok || value == nil {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), typed == float64(int64(typed))
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
