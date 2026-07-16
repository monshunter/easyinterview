package review

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

var (
	ErrReportNotReady               = errors.New("review: report generation job is still active")
	ErrReportInvalidStateTransition = errors.New("review: report state transition is not allowed")
)

type RegenerateReportRequest struct {
	UserID   string
	ReportID string
}

type ReportJobRecord struct {
	ID           string
	JobType      string
	ResourceType string
	ResourceID   string
	Status       sharedtypes.JobStatus
	ErrorCode    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegenerateReportResult struct {
	ReportID string
	Job      ReportJobRecord
}

type RegenerateReportStoreInput struct {
	UserID       string
	ReportID     string
	JobID        string
	AuditEventID string
	Now          time.Time
}

type RegenerateReportStoreResult struct {
	ReportID string
	Job      ReportJobRecord
}

type reportRegenerationRepository interface {
	RegenerateFeedbackReport(ctx context.Context, in RegenerateReportStoreInput) (RegenerateReportStoreResult, error)
}

func (s *Service) RegenerateReport(ctx context.Context, in RegenerateReportRequest) (RegenerateReportResult, error) {
	userID := strings.TrimSpace(in.UserID)
	reportID := strings.TrimSpace(in.ReportID)
	if userID == "" || idx.RequireServerID(reportID) != nil {
		return RegenerateReportResult{}, ErrReportNotFound
	}
	if s == nil || s.repository == nil {
		return RegenerateReportResult{}, fmt.Errorf("review repository is not configured")
	}
	repository, ok := s.repository.(reportRegenerationRepository)
	if !ok {
		return RegenerateReportResult{}, fmt.Errorf("review repository does not implement report regeneration")
	}
	now := s.now().UTC()
	jobID := s.newID()
	result, err := repository.RegenerateFeedbackReport(ctx, RegenerateReportStoreInput{
		UserID:       userID,
		ReportID:     reportID,
		JobID:        jobID,
		AuditEventID: s.newID(),
		Now:          now,
	})
	if err != nil {
		return RegenerateReportResult{}, err
	}
	if err := validateRegenerateReportStoreResult(result, reportID, jobID); err != nil {
		return RegenerateReportResult{}, err
	}
	return RegenerateReportResult{ReportID: result.ReportID, Job: result.Job}, nil
}

func validateRegenerateReportStoreResult(result RegenerateReportStoreResult, reportID, jobID string) error {
	if result.ReportID != reportID || result.Job.ID != jobID ||
		result.Job.JobType != string(sharedjobs.JobTypeReportGenerate) ||
		result.Job.ResourceType != "feedback_report" || result.Job.ResourceID != reportID ||
		result.Job.Status != sharedtypes.JobStatusQueued || strings.TrimSpace(result.Job.ErrorCode) != "" ||
		result.Job.CreatedAt.IsZero() || result.Job.UpdatedAt.IsZero() {
		return fmt.Errorf("review repository returned invalid report regeneration result")
	}
	return nil
}
