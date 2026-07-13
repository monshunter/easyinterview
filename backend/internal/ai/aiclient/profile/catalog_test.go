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
		"target.import.default":           {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"practice.chat.default":           {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"report.generate.default":         {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"resume.parse.default":            {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"resume.tailor.default":           {aiclient.CapabilityChat, aiclient.ProfileStatusActive},
		"target.intel.default":            {aiclient.CapabilityChat, aiclient.ProfileStatusDisabled},
		"practice.voice.stt.default":      {aiclient.CapabilitySTT, aiclient.ProfileStatusDisabled},
		"practice.voice.tts.default":      {aiclient.CapabilityTts, aiclient.ProfileStatusDisabled},
		"practice.voice.realtime.default": {aiclient.CapabilityRealtime, aiclient.ProfileStatusUnsupported},
		"judge.default":                   {aiclient.CapabilityJudge, aiclient.ProfileStatusActive},
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
		"resume.parse.default":    30000,
		"target.import.default":   30000,
		"practice.chat.default":   30000,
		"report.generate.default": 60000,
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

func TestCatalogKeepsResumeParseOutputBudget(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "config", "ai-profiles.yaml")
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader tracked catalog: %v", err)
	}
	defer loader.Close()

	got, err := loader.Resolve("resume.parse.default")
	if err != nil {
		t.Fatalf("Resolve resume.parse.default: %v", err)
	}
	if got.MaxTokens < 8192 {
		t.Fatalf("resume.parse.default max_tokens=%d, want at least 8192 for complete long-resume JSON", got.MaxTokens)
	}
	if got.Version == "1.1.0" {
		t.Fatalf("resume.parse.default version=%q, want a new version for the output budget change", got.Version)
	}
}

func TestTrackedCatalogKeepsExactReportGenerationBudget(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "config", "ai-profiles.yaml")
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader tracked catalog: %v", err)
	}
	defer loader.Close()

	got, err := loader.Resolve("report.generate.default")
	if err != nil {
		t.Fatalf("Resolve report.generate.default: %v", err)
	}
	if got.ContextWindowTokens != 1_000_000 || got.MaxTokens != 6144 || got.TimeoutMs != 60_000 || got.RateLimit.TPM != 60_000 || got.Version != "1.2.0" {
		t.Fatalf("report generation budget drift: context=%d max=%d timeout=%d tpm=%d version=%q", got.ContextWindowTokens, got.MaxTokens, got.TimeoutMs, got.RateLimit.TPM, got.Version)
	}
	if got.Capability != aiclient.CapabilityChat || got.Default.ProviderRef != "deepseek" || got.Default.Model != "deepseek-v4-pro" || got.Route != "report.generate" || len(got.Fallback) != 0 {
		t.Fatalf("report generation route drift: %+v", got)
	}
	if got.Default.Params["thinking"] != "disabled" {
		t.Fatalf("report generation must disable provider thinking mode: %+v", got.Default.Params)
	}
	if _, ok := got.Default.Params["response_format"]; ok {
		t.Fatalf("report response_format must remain output-schema driven: %+v", got.Default.Params)
	}
}

func TestTrackedCatalogKeepsContextAwareJudgeBudgetAndNonThinkingJSON(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "config", "ai-profiles.yaml")
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader tracked catalog: %v", err)
	}
	defer loader.Close()

	got, err := loader.Resolve("judge.default")
	if err != nil {
		t.Fatalf("Resolve judge.default: %v", err)
	}
	if got.TimeoutMs != 60_000 || got.MaxTokens != 6144 || got.RateLimit.TPM != 60_000 || got.Version != "1.2.0" {
		t.Fatalf("context-aware judge budget drift: timeout=%d max=%d tpm=%d version=%q", got.TimeoutMs, got.MaxTokens, got.RateLimit.TPM, got.Version)
	}
	if got.Default.Params["thinking"] != "disabled" || got.Default.Params["response_format"] != "json_object" {
		t.Fatalf("context-aware judge must use non-thinking JSON mode: %+v", got.Default.Params)
	}
	if got.Capability != aiclient.CapabilityJudge || got.Default.ProviderRef != "judge-deepseek" || got.Default.Model != "deepseek-v4-pro" || got.Route != "judge.default" || len(got.Fallback) != 0 {
		t.Fatalf("context-aware judge route drift: %+v", got)
	}
}
