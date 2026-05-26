package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
)

const (
	ErrorCodePrivacyDeleteFailed    = "PRIVACY_DELETE_FAILED"
	ErrorCodePrivacyDeleteRetryable = "PRIVACY_DELETE_RETRYABLE"
)

type PrivacyRequestStore interface {
	LookupDeleteRequestUser(ctx context.Context, privacyRequestID string) (string, error)
	MarkDeleteRequestProcessing(ctx context.Context, privacyRequestID string, now time.Time) error
	MarkDeleteRequestCompleted(ctx context.Context, privacyRequestID string, userID string, deletedFileCount int, now time.Time) error
	MarkDeleteRequestFailed(ctx context.Context, privacyRequestID string, errorCode string, errorMessage string, now time.Time) error
}

type PrivacyDeleteHandlerOptions struct {
	Requests    PrivacyRequestStore
	UploadFiles UploadFileDeleter
	ProfileData func(ctx context.Context, userID string, jobID string) error
	JDMatchData func(ctx context.Context, userID string) error
	Now         func() time.Time
}

type PrivacyDeleteHandler struct {
	requests    PrivacyRequestStore
	uploadFiles UploadFileDeleter
	profileData func(ctx context.Context, userID string, jobID string) error
	jdMatchData func(ctx context.Context, userID string) error
	now         func() time.Time
}

func NewPrivacyDeleteHandler(opts PrivacyDeleteHandlerOptions) *PrivacyDeleteHandler {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &PrivacyDeleteHandler{
		requests:    opts.Requests,
		uploadFiles: opts.UploadFiles,
		profileData: opts.ProfileData,
		jdMatchData: opts.JDMatchData,
		now:         now,
	}
}

func (h *PrivacyDeleteHandler) Handle(ctx context.Context, job targetjob.ClaimedJob) targetjob.JobOutcome {
	if h == nil || h.requests == nil || h.uploadFiles == nil {
		return failedOutcome(ErrorCodePrivacyDeleteFailed, "privacy delete handler is not configured", false)
	}
	if strings.TrimSpace(job.ResourceID) == "" || job.ResourceType != "privacy_request" {
		return failedOutcome(ErrorCodePrivacyDeleteFailed, "privacy delete job resource is invalid", false)
	}
	now := h.now().UTC()
	userID, err := h.requests.LookupDeleteRequestUser(ctx, job.ResourceID)
	if err != nil {
		return failedOutcome(ErrorCodePrivacyDeleteRetryable, fmt.Sprintf("lookup privacy request: %v", err), true)
	}
	if strings.TrimSpace(userID) == "" {
		return failedOutcome(ErrorCodePrivacyDeleteFailed, "privacy delete request has no user", false)
	}
	if err := h.requests.MarkDeleteRequestProcessing(ctx, job.ResourceID, now); err != nil {
		return failedOutcome(ErrorCodePrivacyDeleteRetryable, fmt.Sprintf("mark privacy request processing: %v", err), true)
	}
	deleted, err := Runner{UploadFiles: h.uploadFiles}.DeleteUploadFilesForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, uploadservice.ErrRetryableDelete) {
			return failedOutcome(ErrorCodePrivacyDeleteRetryable, err.Error(), true)
		}
		_ = h.requests.MarkDeleteRequestFailed(ctx, job.ResourceID, ErrorCodePrivacyDeleteFailed, err.Error(), now)
		return failedOutcome(ErrorCodePrivacyDeleteFailed, err.Error(), false)
	}
	if h.profileData != nil {
		if err := h.profileData(ctx, userID, job.JobID); err != nil {
			_ = h.requests.MarkDeleteRequestFailed(ctx, job.ResourceID, ErrorCodePrivacyDeleteFailed, err.Error(), now)
			return failedOutcome(ErrorCodePrivacyDeleteFailed, err.Error(), false)
		}
	}
	if h.jdMatchData != nil {
		if err := h.jdMatchData(ctx, userID); err != nil {
			_ = h.requests.MarkDeleteRequestFailed(ctx, job.ResourceID, ErrorCodePrivacyDeleteFailed, err.Error(), now)
			return failedOutcome(ErrorCodePrivacyDeleteFailed, err.Error(), false)
		}
	}
	if err := h.requests.MarkDeleteRequestCompleted(ctx, job.ResourceID, userID, len(deleted), now); err != nil {
		return failedOutcome(ErrorCodePrivacyDeleteRetryable, fmt.Sprintf("mark privacy request completed: %v", err), true)
	}
	return targetjob.JobOutcome{Succeeded: true}
}

func failedOutcome(code string, message string, retryable bool) targetjob.JobOutcome {
	return targetjob.JobOutcome{ErrorCode: code, ErrorMessage: message, Retryable: retryable}
}
