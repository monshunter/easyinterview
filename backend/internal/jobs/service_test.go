package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestServiceGetJobRequiresUserOwnedJob(t *testing.T) {
	now := time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC)
	store := &recordingStore{
		job: JobRecord{
			ID:           "job-1",
			JobType:      api.JobTypeReportGenerate,
			ResourceType: api.ResourceTypeFeedbackReport,
			ResourceID:   "report-1",
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	service := NewService(ServiceOptions{Store: store})

	got, err := service.GetJob(context.Background(), " user-1 ", " job-1 ")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if store.userID != "user-1" || store.jobID != "job-1" {
		t.Fatalf("store input was not normalized: user=%q job=%q", store.userID, store.jobID)
	}
	if got.ID != "job-1" || got.ResourceType != api.ResourceTypeFeedbackReport || got.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("job drifted: %+v", got)
	}
}

func TestServiceGetJobRejectsMissingIdentityAsNotFound(t *testing.T) {
	service := NewService(ServiceOptions{Store: &recordingStore{}})

	if _, err := service.GetJob(context.Background(), "", "job-1"); !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("missing user err = %v, want ErrJobNotFound", err)
	}
	if _, err := service.GetJob(context.Background(), "user-1", ""); !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("missing job err = %v, want ErrJobNotFound", err)
	}
}

type recordingStore struct {
	userID string
	jobID  string
	job    JobRecord
	err    error
}

func (s *recordingStore) GetJob(ctx context.Context, userID, jobID string) (JobRecord, error) {
	s.userID = userID
	s.jobID = jobID
	if s.err != nil {
		return JobRecord{}, s.err
	}
	return s.job, nil
}
