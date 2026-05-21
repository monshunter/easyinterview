package jobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/generators"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/jobs"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

type fakeAgentRepo struct {
	created   store.CreateAgentScanInput
	updated   store.UpdateAgentScanStatusInput
	createErr error
	updateErr error
}

func (f *fakeAgentRepo) CreateAgentScan(ctx context.Context, in store.CreateAgentScanInput) (jdmatch.AgentScanRecord, error) {
	f.created = in
	if f.createErr != nil {
		return jdmatch.AgentScanRecord{}, f.createErr
	}
	return jdmatch.AgentScanRecord{ID: in.ID, UserID: in.UserID, Status: in.Status}, nil
}

func (f *fakeAgentRepo) UpdateAgentScanStatus(ctx context.Context, in store.UpdateAgentScanStatusInput) (jdmatch.AgentScanRecord, error) {
	f.updated = in
	if f.updateErr != nil {
		return jdmatch.AgentScanRecord{}, f.updateErr
	}
	return jdmatch.AgentScanRecord{ID: in.ID, UserID: in.UserID, Status: in.Status}, nil
}

func fixedNow() time.Time { return time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC) }

func TestAgentScanRunSuccess(t *testing.T) {
	repo := &fakeAgentRepo{}
	emitted := false
	deps := jobs.AgentScanDeps{
		AgentScans: repo,
		Generator: func(ctx context.Context, _ generators.RunRecommendationGeneratorInput) (generators.RunRecommendationGeneratorResult, error) {
			return generators.RunRecommendationGeneratorResult{
				Recommendations: []jdmatch.RecommendationRecord{{ID: "rec-1"}, {ID: "rec-2"}},
				CompletedEvent:  generators.RecommendationCompletedEvent{UserID: "user-A", RecommendationCount: 2},
			}, nil
		},
		NewID:            func() string { return "scan-1" },
		Now:              fixedNow,
		NextScanInterval: 4 * time.Hour,
		OutboxEmit: func(ctx context.Context, event generators.RecommendationCompletedEvent) error {
			emitted = true
			if event.AgentScanID != "scan-1" || event.RecommendationCount != 2 {
				t.Fatalf("event = %#v", event)
			}
			return nil
		},
	}
	if err := jobs.Run(context.Background(), "user-A", deps); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if repo.created.Status != jdmatch.AgentScanStatusScanning {
		t.Fatalf("initial status = %s", repo.created.Status)
	}
	if repo.updated.Status != jdmatch.AgentScanStatusIdle {
		t.Fatalf("final status = %s", repo.updated.Status)
	}
	if !emitted {
		t.Fatalf("expected outbox event on success")
	}
	if repo.updated.RecommendationCount == nil || *repo.updated.RecommendationCount != 2 {
		t.Fatalf("recommendationCount = %v", repo.updated.RecommendationCount)
	}
}

func TestAgentScanRunFailureMarksError(t *testing.T) {
	repo := &fakeAgentRepo{}
	emitted := false
	deps := jobs.AgentScanDeps{
		AgentScans: repo,
		Generator: func(ctx context.Context, _ generators.RunRecommendationGeneratorInput) (generators.RunRecommendationGeneratorResult, error) {
			return generators.RunRecommendationGeneratorResult{}, generators.ErrInvalidLLMOutput
		},
		NewID:            func() string { return "scan-1" },
		Now:              fixedNow,
		NextScanInterval: 4 * time.Hour,
		OutboxEmit: func(ctx context.Context, event generators.RecommendationCompletedEvent) error {
			emitted = true
			return nil
		},
	}
	err := jobs.Run(context.Background(), "user-A", deps)
	if err == nil || !errors.Is(err, generators.ErrInvalidLLMOutput) {
		t.Fatalf("err = %v, want wrapped ErrInvalidLLMOutput", err)
	}
	if repo.updated.Status != jdmatch.AgentScanStatusError {
		t.Fatalf("final status = %s, want error", repo.updated.Status)
	}
	if repo.updated.ErrorMessage == nil || *repo.updated.ErrorMessage != "invalid LLM output" {
		t.Fatalf("error message = %v", repo.updated.ErrorMessage)
	}
	if emitted {
		t.Fatalf("must not emit outbox event on generator failure")
	}
}

func TestAgentScanRunRejectsMissingDeps(t *testing.T) {
	err := jobs.Run(context.Background(), "user-A", jobs.AgentScanDeps{})
	if err == nil {
		t.Fatalf("expected error for missing deps")
	}
}
