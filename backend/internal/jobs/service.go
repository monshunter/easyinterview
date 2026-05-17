package jobs

import (
	"context"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

var ErrJobNotFound = stderrs.New("async job not found")

type Store interface {
	GetJob(ctx context.Context, userID, jobID string) (JobRecord, error)
}

type ServiceOptions struct {
	Store Store
}

type Service struct {
	store Store
}

func NewService(opts ServiceOptions) *Service {
	return &Service{store: opts.Store}
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

func (s *Service) GetJob(ctx context.Context, userID, jobID string) (JobRecord, error) {
	if s == nil || s.store == nil {
		return JobRecord{}, fmt.Errorf("job service store is not configured")
	}
	userID = strings.TrimSpace(userID)
	jobID = strings.TrimSpace(jobID)
	if userID == "" || jobID == "" {
		return JobRecord{}, ErrJobNotFound
	}
	return s.store.GetJob(ctx, userID, jobID)
}
