// Package jobs hosts the JD-Match background workers. P0 baseline
// registers exactly one job handler — jd_match_agent_scan — under
// spec D-12. jd_match_search exists in shared/jobs.yaml as a future-
// async reservation but is NOT registered to the drainer at P0.
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/generators"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

// AgentScanDeps bundles the dependencies the agent_scan job consumes.
type AgentScanDeps struct {
	AgentScans store.AgentScanRepository
	Generator  func(ctx context.Context, in generators.RunRecommendationGeneratorInput) (generators.RunRecommendationGeneratorResult, error)
	NewID      func() string
	Now        func() time.Time
	// NextScanInterval is the cadence used to compute next_scan_at on
	// successful runs; cmd/api wiring reads this from A4 config.
	NextScanInterval time.Duration
	// OutboxEmit emits the jd_match.recommendation.completed event on
	// success. cmd/api wires this to the B3 outbox writer.
	OutboxEmit func(ctx context.Context, event generators.RecommendationCompletedEvent) error
}

// AgentScanRepository is the slice of store.Repository the agent_scan
// job needs.
type AgentScanRepository = store.AgentScanRepository

// Run executes a single per-user agent scan. The state machine:
//
//   - create or update agent_scans row -> status='scanning'
//   - inline call RunRecommendationGenerator
//   - on success: update row -> status='idle' + last_scan_at + next_scan_at + recommendation_count;
//     emit jd_match.recommendation.completed event
//   - on failure: update row -> status='error' + error_message; no event
//
// The returned error is the cause; the caller (drainer) maps to retry
// or terminal failure based on its own policy. ai_task_runs typed
// columns are written by the AIClient observability decorator wired
// in cmd/api, so this job does not touch ai_task_runs directly.
func Run(ctx context.Context, userID string, deps AgentScanDeps) error {
	if deps.AgentScans == nil || deps.Generator == nil || deps.NewID == nil {
		return errors.New("jdmatch agent_scan: dependencies missing")
	}
	now := time.Now().UTC()
	if deps.Now != nil {
		now = deps.Now()
	}
	scanID := deps.NewID()
	started := now
	if _, err := deps.AgentScans.CreateAgentScan(ctx, store.CreateAgentScanInput{
		ID:        scanID,
		UserID:    userID,
		Status:    jdmatch.AgentScanStatusScanning,
		StartedAt: &started,
	}); err != nil {
		return fmt.Errorf("agent_scan: create row: %w", err)
	}
	res, genErr := deps.Generator(ctx, generators.RunRecommendationGeneratorInput{
		UserID:      userID,
		AgentScanID: scanID,
	})
	if genErr != nil {
		errMsg := redactGeneratorError(genErr)
		if _, err := deps.AgentScans.UpdateAgentScanStatus(ctx, store.UpdateAgentScanStatusInput{
			ID:           scanID,
			UserID:       userID,
			Status:       jdmatch.AgentScanStatusError,
			FinishedAt:   ptrTime(now),
			ErrorMessage: &errMsg,
		}); err != nil {
			return fmt.Errorf("agent_scan: mark error: %w (original: %w)", err, genErr)
		}
		return fmt.Errorf("agent_scan generator: %w", genErr)
	}
	finished := now
	last := finished
	next := finished.Add(deps.NextScanInterval)
	count := len(res.Recommendations)
	if _, err := deps.AgentScans.UpdateAgentScanStatus(ctx, store.UpdateAgentScanStatusInput{
		ID:                  scanID,
		UserID:              userID,
		Status:              jdmatch.AgentScanStatusIdle,
		FinishedAt:          &finished,
		LastScanAt:          &last,
		NextScanAt:          &next,
		RecommendationCount: &count,
	}); err != nil {
		return fmt.Errorf("agent_scan: mark idle: %w", err)
	}
	if deps.OutboxEmit != nil {
		// Stamp the agentScanId on the event so consumers can join
		// back to agent_scans without a separate lookup.
		event := res.CompletedEvent
		event.AgentScanID = scanID
		if err := deps.OutboxEmit(ctx, event); err != nil {
			return fmt.Errorf("agent_scan: outbox emit: %w", err)
		}
	}
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }

// redactGeneratorError summarises the generator error without leaking
// LLM prompt content or other PII. Known sentinels keep their short
// label; everything else collapses to the underlying type name.
func redactGeneratorError(err error) string {
	switch {
	case errors.Is(err, generators.ErrInvalidLLMOutput):
		return "invalid LLM output"
	case errors.Is(err, context.DeadlineExceeded):
		return "AI provider timeout"
	}
	if err == nil {
		return "unknown"
	}
	return fmt.Sprintf("%T", err)
}

// Ensure json is imported even if no JSON marshalling is used here —
// kept available for future event payload work the cmd/api wiring may
// inject.
var _ = json.Marshal
