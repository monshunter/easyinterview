package practice

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const firstQuestionFeatureKey = "practice.session.first_question"

type StartSessionRequest struct {
	UserID             string
	PlanID             string
	HintsEnabled       bool
	IdempotencyKeyHash string
	RequestFingerprint string
}

type StartSessionReservationInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	PlanID              string
	HintsEnabled        bool
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
	Mode                sharedtypes.PracticeMode
	InterviewerPersona  sharedtypes.InterviewerRole
	Language            string
	HintsEnabled        bool
	RoleTitle           string
	Seniority           string
	TopSkills           []string
	RubricDimensions    []string
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
	Mode                sharedtypes.PracticeMode
	InterviewerPersona  sharedtypes.InterviewerRole
	Language            string
	HintsEnabled        bool
	TurnID              string
	SessionEventID      string
	OutboxEventID       string
	AuditEventID        string
	QuestionText        string
	QuestionIntent      string
	StartedAt           time.Time
	CreatedAt           time.Time
}

type FailSessionStartInput struct {
	IdempotencyRecordID string
	SessionID           string
	UserID              string
	ErrorCode           string
	Retryable           bool
	FailedAt            time.Time
}

type TurnRecord struct {
	ID             string
	TurnIndex      int32
	QuestionText   string
	QuestionIntent string
	Status         string
	FollowUpCount  int
	AskedAt        time.Time
}

type SessionRecord struct {
	ID           string
	PlanID       string
	TargetJobID  string
	Status       sharedtypes.SessionStatus
	Language     string
	HintsEnabled bool
	TurnCount    int32
	CurrentTurn  *TurnRecord
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) StartPracticeSession(ctx context.Context, in StartSessionRequest) (SessionRecord, error) {
	if s == nil || s.store == nil {
		return SessionRecord{}, fmt.Errorf("practice service is not initialised")
	}
	if s.registry == nil {
		return SessionRecord{}, fmt.Errorf("practice prompt registry is not configured")
	}
	if s.ai == nil {
		return SessionRecord{}, fmt.Errorf("practice AI client is not configured")
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
		HintsEnabled:        in.HintsEnabled,
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

	resolution, err := s.registry.ResolveActive(ctx, firstQuestionFeatureKey, reservation.Language)
	if err != nil {
		return SessionRecord{}, s.failReservedSessionStart(ctx, userID, reservation, serviceErrorFromRegistry(err))
	}
	resp, _, err := s.ai.Complete(ctx, resolution.ModelProfileName, firstQuestionPayload(resolution, reservation))
	if err != nil {
		return SessionRecord{}, s.failReservedSessionStart(ctx, userID, reservation, serviceErrorFromAI(err))
	}
	question, err := parseFirstQuestion(resp.Content)
	if err != nil {
		return SessionRecord{}, s.failReservedSessionStart(ctx, userID, reservation, serviceErrorFromAI(err))
	}

	return s.store.CommitSessionStart(ctx, CommitSessionStartInput{
		IdempotencyRecordID: reservation.IdempotencyRecordID,
		SessionID:           reservation.SessionID,
		UserID:              reservation.UserID,
		PlanID:              reservation.PlanID,
		TargetJobID:         reservation.TargetJobID,
		Goal:                reservation.Goal,
		Mode:                reservation.Mode,
		InterviewerPersona:  reservation.InterviewerPersona,
		Language:            reservation.Language,
		HintsEnabled:        reservation.HintsEnabled,
		TurnID:              s.newID(),
		SessionEventID:      s.newID(),
		OutboxEventID:       s.newID(),
		AuditEventID:        s.newID(),
		QuestionText:        question.Text,
		QuestionIntent:      question.Intent,
		StartedAt:           s.now().UTC(),
		CreatedAt:           reservation.CreatedAt,
	})
}

func (s *Service) failReservedSessionStart(ctx context.Context, userID string, reservation SessionReservation, err error) error {
	var svcErr *ServiceError
	if !stderrs.As(err, &svcErr) || !isPracticeAIErrorCode(svcErr.Code) {
		return err
	}
	meta := sharederrors.CodeRegistry[svcErr.Code]
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

func firstQuestionPayload(resolution registry.PromptResolution, reservation SessionReservation) aiclient.CompletePayload {
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	userContent := renderFirstQuestionTemplate(resolution.UserMessageTemplate, reservation)
	if userContent == "" {
		userContent = "Generate the first interview question."
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
			FeatureKey:        firstQuestionFeatureKey,
			PromptVersion:     resolution.PromptVersion,
			RubricVersion:     resolution.RubricVersion,
			Language:          reservation.Language,
			FeatureFlag:       resolution.FeatureFlag,
			DataSourceVersion: resolution.DataSourceVersion,
			TaskRun: aiclient.AITaskRunContext{
				UserID:       reservation.UserID,
				Capability:   aiclient.AITaskRunTaskQuestionGenerate,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   reservation.TargetJobID,
			},
		},
	}
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

func renderFirstQuestionTemplate(template string, reservation SessionReservation) string {
	content := strings.TrimSpace(template)
	if content == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"{{language}}", fallbackText(reservation.Language, "en"),
		"{{role_title}}", fallbackText(reservation.RoleTitle, "target role"),
		"{{seniority}}", fallbackText(reservation.Seniority, "not specified"),
		"{{top_skills}}", fallbackList(reservation.TopSkills, "target job requirements"),
		"{{rubric_dimensions}}", fallbackList(reservation.RubricDimensions, "practice_depth, practice_dimension_coverage, language_consistency"),
		"{{practice_goal}}", fallbackText(string(reservation.Goal), string(sharedtypes.PracticeGoalBaseline)),
	)
	return strings.TrimSpace(replacer.Replace(content))
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func fallbackList(values []string, fallback string) string {
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		return fallback
	}
	return strings.Join(clean, ", ")
}

type firstQuestion struct {
	Text   string
	Intent string
}

func parseFirstQuestion(content string) (firstQuestion, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return firstQuestion{}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "first question response is empty", false)
	}
	var decoded struct {
		QuestionText   string `json:"questionText"`
		QuestionIntent string `json:"questionIntent"`
		Question       string `json:"question"`
		Intent         string `json:"intent"`
	}
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		return firstQuestion{}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "first question response must be strict JSON", false)
	}
	text := strings.TrimSpace(decoded.QuestionText)
	if text == "" {
		text = strings.TrimSpace(decoded.Question)
	}
	if text == "" {
		return firstQuestion{}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "first question response missing question", false)
	}
	intent := strings.TrimSpace(decoded.QuestionIntent)
	if intent == "" {
		intent = strings.TrimSpace(decoded.Intent)
	}
	return firstQuestion{Text: text, Intent: intent}, nil
}
