package practice

import (
	"context"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type CompletePracticeSessionRequest struct {
	UserID            string
	SessionID         string
	ClientCompletedAt time.Time
}

type CompleteSessionStoreInput struct {
	UserID            string
	SessionID         string
	ReportID          string
	JobID             string
	SessionEventID    string
	OutboxEventID     string
	AuditEventID      string
	ClientCompletedAt time.Time
	Now               time.Time
}

type CompleteSessionResult struct {
	ReportID string
	Job      JobRecord
	Replay   bool
}

type JobRecord struct {
	ID           string
	JobType      api.JobType
	ResourceType api.ResourceType
	ResourceID   string
	Status       sharedtypes.JobStatus
	ErrorCode    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CompletePracticeSession(ctx context.Context, in CompletePracticeSessionRequest) (CompleteSessionResult, error) {
	if s == nil || s.store == nil {
		return CompleteSessionResult{}, fmt.Errorf("practice service is not initialised")
	}
	userID := strings.TrimSpace(in.UserID)
	sessionID := strings.TrimSpace(in.SessionID)
	if userID == "" {
		return CompleteSessionResult{}, fmt.Errorf("userId is required")
	}
	if sessionID == "" {
		return CompleteSessionResult{}, sessionNotFoundError()
	}
	now := s.now().UTC()
	clientCompletedAt := in.ClientCompletedAt.UTC()
	if clientCompletedAt.IsZero() {
		clientCompletedAt = now
	}
	result, err := s.store.CompleteSession(ctx, CompleteSessionStoreInput{
		UserID:            userID,
		SessionID:         sessionID,
		ReportID:          s.newID(),
		JobID:             s.newID(),
		SessionEventID:    s.newID(),
		OutboxEventID:     s.newID(),
		AuditEventID:      s.newID(),
		ClientCompletedAt: clientCompletedAt,
		Now:               now,
	})
	if stderrs.Is(err, ErrSessionNotFound) {
		return CompleteSessionResult{}, sessionNotFoundError()
	}
	if stderrs.Is(err, ErrSessionConflict) {
		return CompleteSessionResult{}, sessionConflictError()
	}
	if err != nil {
		return CompleteSessionResult{}, err
	}
	return result, nil
}
