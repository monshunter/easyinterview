package practice

import (
	"context"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type SendPracticeMessageRequest struct {
	UserID          string
	SessionID       string
	ClientMessageID string
	Text            string
}

type ReservePracticeMessageInput struct {
	UserMessageID   string
	UserID          string
	SessionID       string
	ClientMessageID string
	Text            string
	Now             time.Time
}

const PracticeReplyLeaseDuration = 90 * time.Second

type PracticeMessageReservation struct {
	Session         SessionReservation
	UserMessage     MessageRecord
	History         []MessageRecord
	Replay          *SendPracticeMessageResult
	ReplyGeneration int64
}

type CommitPracticeMessageInput struct {
	UserID                  string
	SessionID               string
	UserMessageID           string
	ExpectedReplyGeneration int64
	AssistantMessageID      string
	AssistantText           string
	Now                     time.Time
}

type FailPracticeMessageInput struct {
	UserID                  string
	SessionID               string
	UserMessageID           string
	ExpectedReplyGeneration int64
	ReplyStatus             PracticeReplyStatus
}

type SendPracticeMessageResult struct {
	Acknowledged     bool
	UserMessage      MessageRecord
	AssistantMessage MessageRecord
	Session          SessionRecord
}

func (s *Service) SendPracticeMessage(ctx context.Context, in SendPracticeMessageRequest) (SendPracticeMessageResult, error) {
	if s == nil || s.store == nil {
		return SendPracticeMessageResult{}, fmt.Errorf("practice service is not initialised")
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.SessionID = strings.TrimSpace(in.SessionID)
	in.ClientMessageID = strings.TrimSpace(in.ClientMessageID)
	in.Text = strings.TrimSpace(in.Text)
	if in.UserID == "" {
		return SendPracticeMessageResult{}, fmt.Errorf("userId is required")
	}
	if in.SessionID == "" {
		return SendPracticeMessageResult{}, sessionNotFoundError()
	}
	if in.ClientMessageID == "" {
		return SendPracticeMessageResult{}, validationError("clientMessageId is required", map[string]any{"field": "clientMessageId"})
	}
	if in.Text == "" {
		return SendPracticeMessageResult{}, validationError("text is required", map[string]any{"field": "text"})
	}
	if len([]rune(in.Text)) > 8000 {
		return SendPracticeMessageResult{}, validationError("text is too long", map[string]any{"field": "text"})
	}

	reservation, err := s.store.ReservePracticeMessage(ctx, ReservePracticeMessageInput{
		UserMessageID:   s.newID(),
		UserID:          in.UserID,
		SessionID:       in.SessionID,
		ClientMessageID: in.ClientMessageID,
		Text:            in.Text,
		Now:             s.now().UTC(),
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return SendPracticeMessageResult{}, sessionNotFoundError()
	}
	if stderrs.Is(err, ErrClientEventMismatch) {
		return SendPracticeMessageResult{}, idempotencyKeyMismatchError()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return SendPracticeMessageResult{}, sessionConflictError()
	}
	if err != nil {
		return SendPracticeMessageResult{}, err
	}
	if reservation.Replay != nil {
		return *reservation.Replay, nil
	}
	history := append(append([]MessageRecord(nil), reservation.History...), reservation.UserMessage)
	assistantText, err := s.generateChatMessage(ctx, reservation.Session, history)
	if err != nil {
		return SendPracticeMessageResult{}, s.failReservedPracticeMessage(ctx, reservation, err)
	}
	result, err := s.store.CommitPracticeMessage(ctx, CommitPracticeMessageInput{
		UserID:                  in.UserID,
		SessionID:               in.SessionID,
		UserMessageID:           reservation.UserMessage.ID,
		ExpectedReplyGeneration: reservation.ReplyGeneration,
		AssistantMessageID:      s.newID(),
		AssistantText:           assistantText,
		Now:                     s.now().UTC(),
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return SendPracticeMessageResult{}, sessionNotFoundError()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return SendPracticeMessageResult{}, sessionConflictError()
	}
	if err != nil {
		return SendPracticeMessageResult{}, s.finalizeReservedPracticeMessageFailure(
			ctx, reservation, PracticeReplyStatusRetryableFailed, err,
		)
	}
	return result, nil
}

const practiceMessageFailureFinalizeTimeout = 5 * time.Second

func (s *Service) failReservedPracticeMessage(ctx context.Context, reservation PracticeMessageReservation, err error) error {
	status := PracticeReplyStatusTerminalFailed
	var serviceErr *ServiceError
	if stderrs.As(err, &serviceErr) {
		if meta, ok := sharederrors.CodeRegistry[serviceErr.Code]; ok && meta.Retryable {
			status = PracticeReplyStatusRetryableFailed
		}
	}
	return s.finalizeReservedPracticeMessageFailure(ctx, reservation, status, err)
}

func (s *Service) finalizeReservedPracticeMessageFailure(
	ctx context.Context,
	reservation PracticeMessageReservation,
	status PracticeReplyStatus,
	err error,
) error {
	failureCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), practiceMessageFailureFinalizeTimeout)
	defer cancel()
	if failErr := s.store.FailPracticeMessage(failureCtx, FailPracticeMessageInput{
		UserID:                  reservation.Session.UserID,
		SessionID:               reservation.Session.SessionID,
		UserMessageID:           reservation.UserMessage.ID,
		ExpectedReplyGeneration: reservation.ReplyGeneration,
		ReplyStatus:             status,
	}); failErr != nil {
		return failErr
	}
	return err
}
