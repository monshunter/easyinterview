package practice

import (
	"context"
	stderrs "errors"
	"fmt"
	"strings"
	"time"
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

type PracticeMessageReservation struct {
	Session     SessionReservation
	UserMessage MessageRecord
	History     []MessageRecord
	Replay      *SendPracticeMessageResult
}

type CommitPracticeMessageInput struct {
	UserID             string
	SessionID          string
	UserMessageID      string
	AssistantMessageID string
	AssistantText      string
	Now                time.Time
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
	if stderrs.Is(err, ErrSessionConflict) || stderrs.Is(err, ErrClientEventMismatch) {
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
		return SendPracticeMessageResult{}, err
	}
	result, err := s.store.CommitPracticeMessage(ctx, CommitPracticeMessageInput{
		UserID:             in.UserID,
		SessionID:          in.SessionID,
		UserMessageID:      reservation.UserMessage.ID,
		AssistantMessageID: s.newID(),
		AssistantText:      assistantText,
		Now:                s.now().UTC(),
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return SendPracticeMessageResult{}, sessionNotFoundError()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return SendPracticeMessageResult{}, sessionConflictError()
	}
	return result, err
}
