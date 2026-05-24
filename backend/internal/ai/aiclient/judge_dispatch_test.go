package aiclient_test

import (
	"context"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/stub"
)

// judgeDispatchResolver adds an active judge-capability profile alongside the
// chat profile so the F3 judge dispatch boundary can be exercised in isolation
// from the real config catalog (judge.default activation lands in Phase 3).
func judgeDispatchResolver() staticResolver {
	return staticResolver{
		"judge.default": {
			Name:       "judge.default",
			Capability: aiclient.CapabilityJudge,
			Status:     aiclient.ProfileStatusActive,
			Default: aiclient.ProviderConfig{
				ProviderRef: stub.Name,
				Model:       "stub-judge-1",
			},
			TimeoutMs: 5000,
			Route:     "judge.default",
			Version:   "0.1.0",
		},
		"practice.followup.default": {
			Name:       "practice.followup.default",
			Capability: aiclient.CapabilityChat,
			Status:     aiclient.ProfileStatusActive,
			Default: aiclient.ProviderConfig{
				ProviderRef: stub.Name,
				Model:       "stub-chat-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
	}
}

// TestCompleteRejectsJudgeProfile asserts the chat-only Complete entry point
// keeps fail-closing when handed a judge-capability profile so business code
// can never reach the judge dispatch (spec §4.2 / plan 004 §2.1).
func TestCompleteRejectsJudgeProfile(t *testing.T) {
	c := newClientWithProviders(t, judgeDispatchResolver(), stubProviderForJudge(t))
	_, meta, err := c.Complete(context.Background(), "judge.default", samplePayload())
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilityJudge)
}

// TestCompleteJudgeRoutesJudgeProfile asserts the dedicated judge dispatch
// accepts a CapabilityJudge profile and surfaces judge-tagged meta.
func TestCompleteJudgeRoutesJudgeProfile(t *testing.T) {
	c := newClientWithProviders(t, judgeDispatchResolver(), stubProviderForJudge(t))
	resp, meta, err := c.CompleteJudge(context.Background(), "judge.default", samplePayload())
	if err != nil {
		t.Fatalf("CompleteJudge: unexpected error: %v", err)
	}
	if resp.Content == "" {
		t.Fatalf("CompleteJudge: expected non-empty content")
	}
	if meta.Capability != aiclient.CapabilityJudge {
		t.Fatalf("CompleteJudge meta.Capability: want %q, got %q", aiclient.CapabilityJudge, meta.Capability)
	}
	if meta.ModelProfileName != "judge.default" {
		t.Fatalf("CompleteJudge meta.ModelProfileName: want judge.default, got %q", meta.ModelProfileName)
	}
}

// TestCompleteJudgeRejectsChatProfile asserts the judge dispatch fail-closes
// when handed a chat-capability profile (capability boundary is symmetric).
func TestCompleteJudgeRejectsChatProfile(t *testing.T) {
	c := newClientWithProviders(t, judgeDispatchResolver(), stubProviderForJudge(t))
	_, meta, err := c.CompleteJudge(context.Background(), "practice.followup.default", samplePayload())
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilityChat)
}

func stubProviderForJudge(t *testing.T) aiclient.Provider {
	t.Helper()
	p, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	return p
}
