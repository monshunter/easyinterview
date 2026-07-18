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

const (
	practiceSystemPolicyStart = "<system_policy>"
	practiceSystemPolicyEnd   = "</system_policy>"
)

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
	RoundID             string
	RoundSequence       int32
	RoundType           string
	RoundName           string
	RoundFocus          string
	Language            string
	RoleTitle           string
	Seniority           string
	TopSkills           []string
	ResumeContext       string
	SemanticFocus       []SemanticFocusDimension
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ReplaySession       *SessionRecord
	RecoverSession      *SessionRecord
}

type SemanticFocusDimension struct {
	Code   string   `json:"code"`
	Label  string   `json:"label"`
	Issues []string `json:"issues"`
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

type CommitSessionStartRecoveryInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	RecoveredAt         time.Time
}

type FailSessionStartInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	ErrorCode           string
	Retryable           bool
	FailedAt            time.Time
}

type PracticeReplyStatus string

const (
	PracticeReplyStatusPending         PracticeReplyStatus = "pending"
	PracticeReplyStatusRetryableFailed PracticeReplyStatus = "retryable_failed"
	PracticeReplyStatusTerminalFailed  PracticeReplyStatus = "terminal_failed"
	PracticeReplyStatusComplete        PracticeReplyStatus = "complete"
)

type MessageRecord struct {
	ID              string
	Role            string
	Content         string
	SeqNo           int32
	ClientMessageID string
	ReplyStatus     PracticeReplyStatus
	CreatedAt       time.Time
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
	if reservation.RecoverSession != nil {
		return s.recoverReservedSessionStart(ctx, userID, reservation)
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

func (s *Service) recoverReservedSessionStart(ctx context.Context, userID string, reservation SessionReservation) (SessionRecord, error) {
	session := *reservation.RecoverSession
	for session.Status == sharedtypes.SessionStatusQueued {
		current, err := s.store.GetSession(ctx, userID, session.ID, s.now().UTC())
		if err != nil {
			return SessionRecord{}, err
		}
		session = current
		if session.Status != sharedtypes.SessionStatusQueued {
			break
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return SessionRecord{}, ctx.Err()
		case <-timer.C:
		}
	}
	if session.Status != sharedtypes.SessionStatusRunning {
		return SessionRecord{}, sessionConflictError()
	}
	recovered, err := s.store.CommitSessionStartRecovery(ctx, CommitSessionStartRecoveryInput{
		IdempotencyRecordID: reservation.IdempotencyRecordID,
		SessionID:           session.ID,
		UserID:              userID,
		RecoveredAt:         s.now().UTC(),
	})
	if stderrs.Is(err, ErrSessionConflict) {
		return SessionRecord{}, sessionConflictError()
	}
	return recovered, err
}

func (s *Service) generateChatMessage(ctx context.Context, reservation SessionReservation, history []MessageRecord) (string, error) {
	if strings.TrimSpace(reservation.ResumeContext) == "" {
		return "", validationError("resume context is unavailable", map[string]any{"field": "resumeId"})
	}
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
			if strings.EqualFold(strings.TrimSpace(response.FinishReason), "length") {
				callErr = sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice chat response reached its output limit", false)
			}
		}
		var message string
		if callErr == nil {
			message, callErr = parseChatMessage(response.Content)
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
	templateSystem, contentTemplate := splitPracticeChatPrompt(resolution.UserMessageTemplate)
	templateSystem = strings.ReplaceAll(templateSystem, "{{language}}", safePromptLanguageTag(reservation.Language))
	content := renderPracticeChatTemplate(contentTemplate, reservation, history)
	systemParts := make([]string, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		systemParts = append(systemParts, system)
	}
	if templateSystem != "" {
		systemParts = append(systemParts, templateSystem)
	}
	if len(systemParts) > 0 {
		messages = append(messages, aiclient.Message{Role: "system", Content: strings.Join(systemParts, "\n\n")})
	}
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
	language := fallbackText(reservation.Language, "en")
	systemLanguage := safePromptLanguageTag(language)
	persona := fallbackText(string(reservation.InterviewerPersona), string(sharedtypes.InterviewerRoleGeneralist))
	targetJobContext := strings.TrimSpace(strings.Join([]string{reservation.RoleTitle, reservation.Seniority, fallbackList(reservation.TopSkills, "target job requirements")}, "; "))
	resumeContext := strings.TrimSpace(reservation.ResumeContext)
	interviewRound := formatPracticeRoundContext(reservation)
	practiceGoal := fallbackText(string(reservation.Goal), string(sharedtypes.PracticeGoalBaseline))
	semanticFocus := reservation.SemanticFocus
	if semanticFocus == nil {
		semanticFocus = []SemanticFocusDimension{}
	}
	conversationHistory := fallbackList(historyValues, "empty; open the conversation naturally")
	return strings.TrimSpace(strings.NewReplacer(
		"{{language}}", systemLanguage,
		"{{interviewer_persona}}", persona,
		"{{target_job_context}}", targetJobContext,
		"{{resume_context}}", resumeContext,
		"{{interview_round}}", interviewRound,
		"{{practice_goal}}", practiceGoal,
		"{{conversation_history}}", conversationHistory,
		"{{language_json}}", jsonTemplateString(language),
		"{{interviewer_persona_json}}", jsonTemplateString(persona),
		"{{target_job_context_json}}", jsonTemplateString(targetJobContext),
		"{{resume_context_json}}", jsonTemplateString(resumeContext),
		"{{interview_round_json}}", jsonTemplateString(interviewRound),
		"{{practice_goal_json}}", jsonTemplateString(practiceGoal),
		"{{semantic_focus_json}}", jsonTemplateValue(semanticFocus),
		"{{conversation_history_json}}", jsonTemplateString(conversationHistory),
	).Replace(template))
}

func safePromptLanguageTag(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "_", "-")
	if value == "" || len(value) > 35 || strings.HasPrefix(value, "-") || strings.HasSuffix(value, "-") || strings.Contains(value, "--") {
		return "en"
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return "en"
	}
	return value
}

func splitPracticeChatPrompt(rendered string) (string, string) {
	start := strings.Index(rendered, practiceSystemPolicyStart)
	end := strings.Index(rendered, practiceSystemPolicyEnd)
	if start < 0 || end < start {
		return "", strings.TrimSpace(rendered)
	}
	policyStart := start + len(practiceSystemPolicyStart)
	userContent := strings.TrimSpace(rendered[:start] + "\n" + rendered[end+len(practiceSystemPolicyEnd):])
	return strings.TrimSpace(rendered[policyStart:end]), userContent
}

func jsonTemplateString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func jsonTemplateValue(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func formatPracticeRoundContext(reservation SessionReservation) string {
	if strings.TrimSpace(reservation.RoundID) == "" || reservation.RoundSequence < 1 {
		return "round context unavailable"
	}
	return strings.Join([]string{
		"id=" + strings.TrimSpace(reservation.RoundID),
		fmt.Sprintf("sequence=%d", reservation.RoundSequence),
		"type=" + fallbackText(reservation.RoundType, "other"),
		"name=" + fallbackText(reservation.RoundName, "unnamed round"),
		"focus=" + fallbackText(reservation.RoundFocus, "not specified"),
	}, "; ")
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
