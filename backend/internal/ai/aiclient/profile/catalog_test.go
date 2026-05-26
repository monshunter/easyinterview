package profile_test

import (
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
)

func TestTrackedCatalogCoversF3AndProductUICapabilityProfiles(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "config", "ai-profiles.yaml")
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader tracked catalog: %v", err)
	}
	defer loader.Close()

	required := map[string]struct {
		capability aiclient.Capability
		status     aiclient.ProfileStatus
	}{
		"target.import.default":             {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"practice.first_question.default":   {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"practice.followup.default":         {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"practice.turn_observe.default":     {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"report.generate.default":           {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"report.assessment.default":         {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"resume.parse.default":              {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"resume.tailor.default":             {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"debrief.generate.default":          {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"debrief.suggest_questions.default": {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"target.intel.default":              {aiclient.CapabilityChat, aiclient.ProfileStatusDisabled},
		"profile.update.default":            {aiclient.CapabilityChat, aiclient.ProfileStatusDisabled},
		"practice.dictation.stt.default":    {aiclient.CapabilitySTT, aiclient.ProfileStatusUnsupported},
		"practice.voice.realtime.default":   {aiclient.CapabilityRealtime, aiclient.ProfileStatusUnsupported},
		"debrief.voice.extract.default":     {aiclient.CapabilitySTT, aiclient.ProfileStatusUnsupported},
		"judge.default":                     {aiclient.CapabilityJudge, aiclient.ProfileStatusActive},
	}

	for name, want := range required {
		t.Run(name, func(t *testing.T) {
			got, err := loader.Resolve(name)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got.Capability != want.capability {
				t.Fatalf("expected capability=%q, got %q", want.capability, got.Capability)
			}
			if got.Status != want.status {
				t.Fatalf("expected status=%q, got %q", want.status, got.Status)
			}
			if got.Status != aiclient.ProfileStatusActive && got.UnsupportedReason == "" {
				t.Fatalf("non-active profile must explain unsupported_reason")
			}
		})
	}
}

func TestTrackedCatalogKeepsManualUATFullFunnelProfilesWithRealProviderBudget(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "config", "ai-profiles.yaml")
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader tracked catalog: %v", err)
	}
	defer loader.Close()

	requiredTimeoutMs := map[string]int{
		"resume.parse.default":            30000,
		"target.import.default":           30000,
		"practice.first_question.default": 30000,
		"practice.followup.default":       30000,
		"practice.turn_observe.default":   20000,
		"report.assessment.default":       30000,
		"report.generate.default":         60000,
	}
	for name, minTimeout := range requiredTimeoutMs {
		t.Run(name, func(t *testing.T) {
			got, err := loader.Resolve(name)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got.TimeoutMs < minTimeout {
				t.Fatalf("%s timeout_ms=%d, want at least %d for real provider manual UAT", name, got.TimeoutMs, minTimeout)
			}
		})
	}
}
