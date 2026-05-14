package practice

import (
	"context"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCompletePracticeSessionCreatesReportJob(t *testing.T) {
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	store := &recordingPlanStore{}
	service := NewService(ServiceOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("report-1", "job-1", "event-1", "outbox-1", "audit-1"),
	})

	result, err := service.CompletePracticeSession(context.Background(), CompletePracticeSessionRequest{
		UserID:            "user-1",
		SessionID:         "session-1",
		ClientCompletedAt: now,
	})
	if err != nil {
		t.Fatalf("CompletePracticeSession returned error: %v", err)
	}
	if result.ReportID != "report-1" ||
		result.Job.ID != "job-1" ||
		result.Job.JobType != api.JobTypeReportGenerate ||
		result.Job.ResourceType != api.ResourceTypeFeedbackReport ||
		result.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected result: %+v", result)
	}
	if store.complete.ReportID != "report-1" ||
		store.complete.JobID != "job-1" ||
		store.complete.SessionEventID != "event-1" ||
		store.complete.OutboxEventID != "outbox-1" ||
		store.complete.AuditEventID != "audit-1" {
		t.Fatalf("complete ids not generated: %+v", store.complete)
	}
}

func TestCompletePracticeSessionReturnsRepositoryReplay(t *testing.T) {
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	store := &recordingPlanStore{
		completeResult: CompleteSessionResult{
			ReportID: "report-existing",
			Replay:   true,
			Job: JobRecord{
				ID:           "job-existing",
				JobType:      api.JobTypeReportGenerate,
				ResourceType: api.ResourceTypeFeedbackReport,
				ResourceID:   "report-existing",
				Status:       sharedtypes.JobStatusQueued,
				CreatedAt:    now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			},
		},
	}
	service := NewService(ServiceOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("report-new", "job-new", "event-new", "outbox-new", "audit-new"),
	})

	result, err := service.CompletePracticeSession(context.Background(), CompletePracticeSessionRequest{
		UserID:    "user-1",
		SessionID: "session-1",
	})
	if err != nil {
		t.Fatalf("CompletePracticeSession returned error: %v", err)
	}
	if !result.Replay || result.ReportID != "report-existing" || result.Job.ID != "job-existing" {
		t.Fatalf("expected repository replay result, got %+v", result)
	}
}
