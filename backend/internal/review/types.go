package review

import (
	"context"
	"errors"
	"time"

	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const ReportGenerateJobType = string(sharedjobs.JobTypeReportGenerate)

var ErrIllegalTransition = errors.New("review: illegal report status transition")

type AsyncJob struct {
	JobID        string
	JobType      string
	ResourceType string
	ResourceID   string
	Payload      []byte
	Attempts     int32
	MaxAttempts  int32
	AvailableAt  time.Time
	LockedAt     *time.Time
}

type ReportStatusUpdate struct {
	ReportID string
	From     sharedtypes.ReportStatus
	To       sharedtypes.ReportStatus
	Now      time.Time
}

type ReportOutcome struct {
	Succeeded         bool
	AsyncJobFinalized bool
	ErrorCode         string
	ErrorMessage      string
	Retryable         bool
}

type AsyncJobFailure struct {
	JobID       string
	Retryable   bool
	ErrorCode   string
	Error       string
	AvailableAt time.Time
	Now         time.Time
}

type ReportService interface {
	GenerateReport(ctx context.Context, job AsyncJob) ReportOutcome
}

func CanTransitionReportStatus(from, to sharedtypes.ReportStatus) bool {
	switch from {
	case sharedtypes.ReportStatusQueued:
		return to == sharedtypes.ReportStatusGenerating
	case sharedtypes.ReportStatusGenerating:
		return to == sharedtypes.ReportStatusReady || to == sharedtypes.ReportStatusFailed
	case sharedtypes.ReportStatusFailed:
		return to == sharedtypes.ReportStatusQueued
	default:
		return false
	}
}
