package practice

import (
	"bytes"
	"context"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const practiceChatFeatureKey = "practice.session.chat"

type StartSessionRequest struct {
	UserID             string
	PlanID             string
	IdempotencyKeyHash string
	RequestFingerprint string
}

type StartSessionReservationInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	PlanID              string
	IdempotencyKeyHash  string
	RequestFingerprint  string
	ExpiresAt           time.Time
	Now                 time.Time
}

type SessionReservation struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	PlanID              string
	TargetJobID         string
	Goal                sharedtypes.PracticeGoal
	InterviewerPersona  sharedtypes.InterviewerRole
	Language            string
	RoleTitle           string
	Seniority           string
	TopSkills           []string
	ResumeProfile       string
	FocusCompetencies   []string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ReplaySession       *SessionRecord
}

type CommitSessionStartInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	PlanID              string
	TargetJobID         string
	Goal                sharedtypes.PracticeGoal
	InterviewerPersona  sharedtypes.InterviewerRole
	Language            string
	MessageID           string
	SessionEventID      string
	OutboxEventID       string
	AuditEventID        string
	MessageText         string
	StartedAt           time.Time
}

type FailSessionStartInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	ErrorCode           string
	Retryable           bool
	FailedAt            time.Time
}

type MessageRecord struct {
	ID        string
	Role      string
	Content   string
	SeqNo     int32
	CreatedAt time.Time
}

type SessionRecord struct {
	ID          string
	PlanID      string
	TargetJobID string
	Status      sharedtypes.SessionStatus
	Language    string
	Messages    []MessageRecord
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Service) StartPracticeSession(ctx context.Context, in StartSessionRequest) (SessionRecord, error) {
	if s == nil || s.store == nil {
		return SessionRecord{}, fmt.Errorf("practice service is not initialised")
	}
	userID := strings.TrimSpace(in.UserID)
	planID := strings.TrimSpace(in.PlanID)
	if userID == "" {
		return SessionRecord{}, fmt.Errorf("userId is required")
	}
	if planID == "" {
		return SessionRecord{}, validationError("planId is required", map[string]any{"field": "planId"})
	}
	if strings.TrimSpace(in.IdempotencyKeyHash) == "" || strings.TrimSpace(in.RequestFingerprint) == "" {
		return SessionRecord{}, validationError("Idempotency-Key header is required", map[string]any{"field": "Idempotency-Key"})
	}

	now := s.now().UTC()
	reservation, err := s.store.ReserveSessionStart(ctx, StartSessionReservationInput{
		IdempotencyRecordID: s.newID(),
		SessionID:           s.newID(),
		UserID:              userID,
		PlanID:              planID,
		IdempotencyKeyHash:  strings.TrimSpace(in.IdempotencyKeyHash),
		RequestFingerprint:  strings.TrimSpace(in.RequestFingerprint),
		ExpiresAt:           now.Add(idempotency.DefaultTTL),
		Now:                 now,
	})
	if stderrs.Is(err, ErrPlanNotFound) {
		return SessionRecord{}, planNotFoundError()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return SessionRecord{}, sessionConflictError()
	}
	if err != nil {
		return SessionRecord{}, err
	}
	if reservation.ReplaySession != nil {
		return *reservation.ReplaySession, nil
	}

	message, err := s.generateChatMessage(ctx, reservation, nil)
	if err != nil {
		return SessionRecord{}, s.failReservedSessionStart(ctx, userID, reservation, err)
	}
	startedAt := s.now().UTC()
	return s.store.CommitSessionStart(ctx, CommitSessionStartInput{
		IdempotencyRecordID: reservation.IdempotencyRecordID,
		SessionID:           reservation.SessionID,
		UserID:              reservation.UserID,
		PlanID:              reservation.PlanID,
		TargetJobID:         reservation.TargetJobID,
		Goal:                reservation.Goal,
		InterviewerPersona:  reservation.InterviewerPersona,
		Language:            reservation.Language,
		MessageID:           s.newID(),
		SessionEventID:      s.newID(),
		OutboxEventID:       s.newID(),
		AuditEventID:        s.newID(),
		MessageText:         message,
		StartedAt:           startedAt,
	})
}

func (s *Service) generateChatMessage(ctx context.Context, reservation SessionReservation, history []MessageRecord) (string, error) {
	if s.registry == nil || s.ai == nil {
		return "", aiConfigError()
	}
	resolution, err := s.registry.ResolveActive(ctx, practiceChatFeatureKey, reservation.Language)
	if err != nil {
		return "", serviceErrorFromRegistry(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		payload := practiceChatPayload(resolution, reservation, history, attempt > 0)
		response, _, callErr := s.ai.Complete(ctx, resolution.ModelProfileName, payload)
		if callErr == nil {
			message, parseErr := parseChatMessage(response.Content)
			callErr = parseErr
			if callErr == nil {
				callErr = validateGeneratedMessageLanguage(message, reservation.Language)
			}
			if callErr == nil {
				return message, nil
			}
		}
		if attempt == 0 && isRepairableAIOutput(callErr) {
			continue
		}
		return "", serviceErrorFromAI(callErr)
	}
	return "", serviceErrorFromAI(sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice chat generation failed", false))
}

func practiceChatPayload(resolution registry.PromptResolution, reservation SessionReservation, history []MessageRecord, repair bool) aiclient.CompletePayload {
	messages := make([]aiclient.Message, 0, 3)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: system})
	}
	content := renderPracticeChatTemplate(resolution.UserMessageTemplate, reservation, history)
	if repair {
		content += "\nReturn only strict JSON with one non-empty messageText in the requested language."
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: content})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: attachOutputSchema(aiclient.CallMetadata{
			FeatureKey:        practiceChatFeatureKey,
			PromptVersion:     resolution.PromptVersion,
			RubricVersion:     resolution.RubricVersion,
			Language:          reservation.Language,
			FeatureFlag:       resolution.FeatureFlag,
			DataSourceVersion: resolution.DataSourceVersion,
			TaskRun: aiclient.AITaskRunContext{
				UserID:       reservation.UserID,
				Capability:   aiclient.AITaskRunTaskPracticeChat,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   reservation.TargetJobID,
			},
		}, resolution),
	}
}

func renderPracticeChatTemplate(template string, reservation SessionReservation, history []MessageRecord) string {
	historyValues := make([]string, 0, len(history))
	for _, message := range history {
		historyValues = append(historyValues, fmt.Sprintf("%s: %s", message.Role, message.Content))
	}
	return strings.TrimSpace(strings.NewReplacer(
		"{{language}}", fallbackText(reservation.Language, "en"),
		"{{target_job_context}}", strings.TrimSpace(strings.Join([]string{reservation.RoleTitle, reservation.Seniority, fallbackList(reservation.TopSkills, "target job requirements")}, "; ")),
		"{{resume_context}}", fallbackText(reservation.ResumeProfile, "resume context unavailable"),
		"{{interview_round}}", fallbackText(string(reservation.InterviewerPersona), "generalist"),
		"{{practice_goal}}", fallbackText(string(reservation.Goal), string(sharedtypes.PracticeGoalBaseline)),
		"{{focus_competencies}}", fallbackList(reservation.FocusCompetencies, "follow the strongest unresolved signal"),
		"{{conversation_history}}", fallbackList(historyValues, "empty; open the conversation naturally"),
	).Replace(template))
}

func parseChatMessage(content string) (string, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()
	var decoded struct {
		MessageText string `json:"messageText"`
	}
	if err := decoder.Decode(&decoded); err != nil {
		return "", sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice chat response must be strict JSON", false)
	}
	if err := decoder.Decode(&struct{}{}); !stderrs.Is(err, io.EOF) {
		return "", sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice chat response contains trailing data", false)
	}
	message := strings.TrimSpace(decoded.MessageText)
	if message == "" {
		return "", sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice chat response missing messageText", false)
	}
	return message, nil
}

func validateGeneratedMessageLanguage(text, language string) error {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(language), "_", "-"))
	hanCount, latinCount := 0, 0
	for _, r := range strings.TrimSpace(text) {
		if unicode.Is(unicode.Han, r) {
			hanCount++
		}
		if unicode.Is(unicode.Latin, r) {
			latinCount++
		}
	}
	switch {
	case normalized == "zh" || strings.HasPrefix(normalized, "zh-"):
		if hanCount > 0 && hanCount*5 >= (hanCount+latinCount)*3 {
			return nil
		}
	case normalized == "en" || strings.HasPrefix(normalized, "en-"):
		if latinCount > 0 && hanCount == 0 {
			return nil
		}
	default:
		return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice session language is unsupported", false)
	}
	return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "generated message language does not match the session", false)
}

func isRepairableAIOutput(err error) bool {
	code, ok := aiErrorCode(err)
	return ok && code == sharederrors.CodeAiOutputInvalid
}

func (s *Service) failReservedSessionStart(ctx context.Context, userID string, reservation SessionReservation, err error) error {
	var svcErr *ServiceError
	if !stderrs.As(err, &svcErr) {
		return err
	}
	meta, ok := sharederrors.CodeRegistry[svcErr.Code]
	if !ok {
		return err
	}
	if failErr := s.store.FailSessionStart(ctx, FailSessionStartInput{
		IdempotencyRecordID: reservation.IdempotencyRecordID,
		SessionID:           reservation.SessionID,
		UserID:              userID,
		ErrorCode:           svcErr.Code,
		Retryable:           meta.Retryable,
		FailedAt:            s.now().UTC(),
	}); failErr != nil {
		return failErr
	}
	return svcErr
}

func aiConfigError() *ServiceError {
	meta := sharederrors.CodeRegistry[sharederrors.CodeAiProviderConfigInvalid]
	return &ServiceError{Code: sharederrors.CodeAiProviderConfigInvalid, Message: meta.Message}
}

func attachOutputSchema(metadata aiclient.CallMetadata, resolution registry.PromptResolution) aiclient.CallMetadata {
	if resolution.OutputSchema != nil {
		metadata.OutputSchema = *resolution.OutputSchema
	}
	return metadata
}

func serviceErrorFromRegistry(err error) error {
	if err == nil {
		return nil
	}
	if code, ok := aiErrorCode(err); ok {
		meta := sharederrors.CodeRegistry[code]
		return &ServiceError{Code: code, Message: meta.Message}
	}
	meta := sharederrors.CodeRegistry[sharederrors.CodeAiProviderConfigInvalid]
	return &ServiceError{Code: sharederrors.CodeAiProviderConfigInvalid, Message: meta.Message}
}

func fallbackText(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func fallbackList(values []string, fallback string) string {
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			clean = append(clean, value)
		}
	}
	if len(clean) == 0 {
		return fallback
	}
	return strings.Join(clean, ", ")
}
