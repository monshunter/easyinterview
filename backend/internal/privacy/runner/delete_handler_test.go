package runner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/privacy/runner"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/upload/service"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestPrivacyDeleteHandlerDeletesUploadFilesForRequestUser(t *testing.T) {
	requests := &fakePrivacyRequestStore{userID: "user-1"}
	uploads := &fakeUploadFileDeleter{deleted: []store.DeletedFileObject{
		{ID: "file-1", ObjectKey: "user-1/resume/file-1.pdf", Purpose: store.PurposeResume},
	}}
	handler := runner.NewPrivacyDeleteHandler(runner.PrivacyDeleteHandlerOptions{
		Requests:    requests,
		UploadFiles: uploads,
		Now:         fixedPrivacyNow,
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID:        "job-1",
		JobType:      "privacy_delete",
		ResourceType: "privacy_request",
		ResourceID:   "privacy-request-1",
	})

	if !outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != "" {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if requests.lookupID != "privacy-request-1" || uploads.userID != "user-1" {
		t.Fatalf("lookupID=%q uploadUser=%q", requests.lookupID, uploads.userID)
	}
	if requests.completedID != "privacy-request-1" || requests.completedCount != 1 {
		t.Fatalf("completedID=%q completedCount=%d", requests.completedID, requests.completedCount)
	}
}

func TestPrivacyDeleteHandlerRetryableUploadFailureKeepsJobRetryable(t *testing.T) {
	requests := &fakePrivacyRequestStore{userID: "user-1"}
	uploads := &fakeUploadFileDeleter{err: service.ErrRetryableDelete}
	handler := runner.NewPrivacyDeleteHandler(runner.PrivacyDeleteHandlerOptions{
		Requests:    requests,
		UploadFiles: uploads,
		Now:         fixedPrivacyNow,
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID:        "job-1",
		JobType:      "privacy_delete",
		ResourceType: "privacy_request",
		ResourceID:   "privacy-request-1",
	})

	if outcome.Succeeded || !outcome.Retryable || outcome.ErrorCode != runner.ErrorCodePrivacyDeleteRetryable {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if requests.completedID != "" || requests.failedID != "" {
		t.Fatalf("retryable failure must not terminally complete request: completed=%q failed=%q", requests.completedID, requests.failedID)
	}
}

func TestPrivacyDeleteHandlerNonRetryableFailureMarksRequestFailed(t *testing.T) {
	requests := &fakePrivacyRequestStore{userID: "user-1"}
	uploads := &fakeUploadFileDeleter{err: errors.New("permission denied")}
	handler := runner.NewPrivacyDeleteHandler(runner.PrivacyDeleteHandlerOptions{
		Requests:    requests,
		UploadFiles: uploads,
		Now:         fixedPrivacyNow,
	})

	outcome := handler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID:        "job-1",
		JobType:      "privacy_delete",
		ResourceType: "privacy_request",
		ResourceID:   "privacy-request-1",
	})

	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != runner.ErrorCodePrivacyDeleteFailed {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if requests.failedID != "privacy-request-1" {
		t.Fatalf("failedID=%q", requests.failedID)
	}
}

type fakePrivacyRequestStore struct {
	userID         string
	lookupID       string
	processingID   string
	completedID    string
	completedCount int
	failedID       string
}

func (s *fakePrivacyRequestStore) LookupDeleteRequestUser(ctx context.Context, privacyRequestID string) (string, error) {
	s.lookupID = privacyRequestID
	return s.userID, nil
}

func (s *fakePrivacyRequestStore) MarkDeleteRequestProcessing(ctx context.Context, privacyRequestID string, now time.Time) error {
	s.processingID = privacyRequestID
	return nil
}

func (s *fakePrivacyRequestStore) MarkDeleteRequestCompleted(ctx context.Context, privacyRequestID string, deletedFileCount int, now time.Time) error {
	s.completedID = privacyRequestID
	s.completedCount = deletedFileCount
	return nil
}

func (s *fakePrivacyRequestStore) MarkDeleteRequestFailed(ctx context.Context, privacyRequestID string, errorCode string, errorMessage string, now time.Time) error {
	s.failedID = privacyRequestID
	return nil
}

func fixedPrivacyNow() time.Time {
	return time.Date(2026, 5, 12, 3, 0, 0, 0, time.UTC)
}
