package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/jobs"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGetJobReturnsOwnedAsyncJob(t *testing.T) {
	now := time.Date(2026, 5, 17, 10, 30, 0, 0, time.UTC)
	service := &fakeJobService{job: domain.JobRecord{
		ID:           "01918fa0-0000-7000-8000-00000000f301",
		JobType:      api.JobTypeDebriefGenerate,
		ResourceType: api.ResourceTypeDebrief,
		ResourceID:   "01918fa0-0000-7000-8000-00000000d010",
		Status:       sharedtypes.JobStatusQueued,
		CreatedAt:    now,
		UpdatedAt:    now,
	}}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.GetJob(rec, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/job-1", nil), " job-1 ")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out api.Job
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode Job: %v", err)
	}
	if service.userID != "user-1" || service.jobID != "job-1" {
		t.Fatalf("service input drifted: user=%q job=%q", service.userID, service.jobID)
	}
	if out.Id != service.job.ID || out.JobType != api.JobTypeDebriefGenerate || out.ResourceType != api.ResourceTypeDebrief {
		t.Fatalf("job response drifted: %+v", out)
	}
}

func TestGetJobNotFoundUsesOpaque404(t *testing.T) {
	service := &fakeJobService{err: domain.ErrJobNotFound}
	handler := NewHandler(HandlerOptions{Service: service, Session: staticSession("user-1")})
	rec := httptest.NewRecorder()

	handler.GetJob(rec, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/job-1", nil), "job-1")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode ApiErrorResponse: %v", err)
	}
	if out.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("error code = %s, want %s", out.Error.Code, sharederrors.CodeValidationFailed)
	}
}

func TestGetJobUnauthenticated(t *testing.T) {
	handler := NewHandler(HandlerOptions{Service: &fakeJobService{}})
	rec := httptest.NewRecorder()

	handler.GetJob(rec, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/job-1", nil), "job-1")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

type fakeJobService struct {
	userID string
	jobID  string
	job    domain.JobRecord
	err    error
}

func (s *fakeJobService) GetJob(ctx context.Context, userID, jobID string) (domain.JobRecord, error) {
	s.userID = userID
	s.jobID = jobID
	if s.err != nil {
		return domain.JobRecord{}, s.err
	}
	return s.job, nil
}

func staticSession(userID string) SessionResolver {
	return func(context.Context) (string, bool) { return userID, userID != "" }
}
