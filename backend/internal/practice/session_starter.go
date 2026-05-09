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
	PlanID              string
	TargetJobID         string
	Goal                sharedtypes.PracticeGoal
	Mode                sharedtypes.PracticeMode
	InterviewerPersona  sharedtypes.InterviewerRole
	Language            string
	HintsEnabled        bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CommitSessionStartInput struct {
	IdempotencyRecordID string
	SessionID           string
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
	QuestionText        string
	QuestionIntent      string
	StartedAt           time.Time
	CreatedAt           time.Time
}

type TurnRecord struct {
	ID             string
	TurnIndex      int32
	QuestionText   string
	QuestionIntent string
	Status         string
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
	if err != nil {
		return SessionRecord{}, err
	}

	resolution, err := s.registry.ResolveActive(ctx, firstQuestionFeatureKey, reservation.Language)
	if err != nil {
		return SessionRecord{}, err
	}
	resp, _, err := s.ai.Complete(ctx, resolution.ModelProfileName, firstQuestionPayload(resolution, reservation))
	if err != nil {
		return SessionRecord{}, err
	}
	question, err := parseFirstQuestion(resp.Content)
	if err != nil {
		return SessionRecord{}, err
	}

	return s.store.CommitSessionStart(ctx, CommitSessionStartInput{
		IdempotencyRecordID: reservation.IdempotencyRecordID,
		SessionID:           reservation.SessionID,
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
		QuestionText:        question.Text,
		QuestionIntent:      question.Intent,
		StartedAt:           s.now().UTC(),
		CreatedAt:           reservation.CreatedAt,
	})
}

func firstQuestionPayload(resolution registry.PromptResolution, reservation SessionReservation) aiclient.CompletePayload {
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	userContent := strings.TrimSpace(resolution.UserMessageTemplate)
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
		},
	}
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
	}
	if err := json.Unmarshal([]byte(content), &decoded); err == nil {
		if strings.TrimSpace(decoded.QuestionText) == "" {
			return firstQuestion{}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "first question response missing questionText", false)
		}
		return firstQuestion{Text: strings.TrimSpace(decoded.QuestionText), Intent: strings.TrimSpace(decoded.QuestionIntent)}, nil
	}
	return firstQuestion{Text: content, Intent: "general"}, nil
}
