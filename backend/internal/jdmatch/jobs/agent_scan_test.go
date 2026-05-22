package jobs_test

import (
	"context"
	"encoding/json"
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
	var captured generators.RunRecommendationGeneratorInput
	deps := jobs.AgentScanDeps{
		AgentScans: repo,
		CandidateProfile: func(ctx context.Context, userID string) (json.RawMessage, error) {
			if userID != "user-A" {
				t.Fatalf("candidate profile user = %q", userID)
			}
			return json.RawMessage(`{"displayName":"Alice","skills":["Go"]}`), nil
		},
		JobsPool: func(ctx context.Context, userID string) (json.RawMessage, error) {
			if userID != "user-A" {
				t.Fatalf("jobs pool user = %q", userID)
			}
			return json.RawMessage(`[{"jobMatchId":"rec-1","title":"Backend"}]`), nil
		},
		Generator: func(ctx context.Context, in generators.RunRecommendationGeneratorInput) (generators.RunRecommendationGeneratorResult, error) {
			captured = in
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
	if captured.UserID != "user-A" || captured.AgentScanID != "scan-1" {
		t.Fatalf("generator input identity = %+v", captured)
	}
	if got := string(captured.CandidateProfileJSON); got != `{"displayName":"Alice","skills":["Go"]}` {
		t.Fatalf("candidate profile json = %s", got)
	}
	if got := string(captured.JobsPoolJSON); got != `[{"jobMatchId":"rec-1","title":"Backend"}]` {
		t.Fatalf("jobs pool json = %s", got)
	}
}

func TestAgentScanRunFailureMarksError(t *testing.T) {
	repo := &fakeAgentRepo{}
	emitted := false
	deps := jobs.AgentScanDeps{
		AgentScans: repo,
		CandidateProfile: func(ctx context.Context, userID string) (json.RawMessage, error) {
			return json.RawMessage(`{"displayName":"Alice"}`), nil
		},
		JobsPool: func(ctx context.Context, userID string) (json.RawMessage, error) {
			return json.RawMessage(`[{"jobMatchId":"rec-1"}]`), nil
		},
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
