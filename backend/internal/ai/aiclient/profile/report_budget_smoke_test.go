//go:build smoke

package profile_test

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	openaicompatible "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible"
)

const reportProfileLiveTokenGateEnv = "RUN_REPORT_PROFILE_LIVE_TOKEN_GATE"

type environmentSecretSource struct{}

func (environmentSecretSource) Get(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", providerregistry.ErrSecretMissing
	}
	return value, nil
}

func TestReportProfileLiveTokenGate(t *testing.T) {
	if os.Getenv(reportProfileLiveTokenGateEnv) != "1" {
		t.Skipf("set %s=1 and source deploy/dev-stack/.env to run the opt-in live token gate", reportProfileLiveTokenGateEnv)
	}

	repoRoot := filepath.Clean(filepath.Join("..", "..", "..", "..", ".."))
	reportProfile := loadLiveReportProfile(t, repoRoot)
	adapter := loadLiveReportProvider(t, repoRoot, reportProfile)
	fixtureDir := filepath.Join(repoRoot, "backend", "internal", "review", "testdata", "report-boundary")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	framedMessages := buildLiveFramedInput(t, reportInputByteLimit)
	exactResponse, exactMeta, err := adapter.Complete(ctx, reportProfile, aiclient.CompletePayload{Messages: framedMessages})
	if err != nil {
		t.Fatal("exact framed live probe failed; provider error details redacted")
	}
	if err := validateReportProfileLiveProbeResult(reportProfile, exactMeta, exactResponse.FinishReason, 0); err != nil {
		t.Fatalf("exact framed live probe rejected: %v", err)
	}
	logLiveProbeEvidence(t, "framed-917504", reportProfile, exactMeta, exactResponse.FinishReason)

	for _, fixtureName := range []string{"output-worst-case-en.json", "output-worst-case-zh-CN.json"} {
		t.Run(fixtureName, func(t *testing.T) {
			fixture := mustReadLiveProbeFixture(t, filepath.Join(fixtureDir, fixtureName))
			probeProfile := *reportProfile
			probeProfile.Default = reportProfile.Default
			probeProfile.Default.Params = maps.Clone(reportProfile.Default.Params)
			probeProfile.Default.Params["temperature"] = 0.0
			response, meta, err := adapter.Complete(ctx, &probeProfile, aiclient.CompletePayload{Messages: []aiclient.Message{
				{Role: "system", Content: "Return exactly the ASCII word OK and nothing else."},
				{Role: "user", Content: string(fixture)},
			}})
			if err != nil {
				t.Fatal("output-fixture live token probe failed; provider error details redacted")
			}
			if err := validateReportProfileLiveProbeResult(reportProfile, meta, response.FinishReason, reportProfile.MaxTokens); err != nil {
				t.Fatalf("output-fixture live token probe rejected: %v", err)
			}
			logLiveProbeEvidence(t, fixtureName, reportProfile, meta, response.FinishReason)
		})
	}
}

func buildLiveFramedInput(t *testing.T, targetBytes int) []aiclient.Message {
	t.Helper()
	messages := []aiclient.Message{
		{Role: "system", Content: "Synthetic non-sensitive report capacity probe."},
		{Role: "user", Content: ""},
	}
	base, err := json.Marshal(messages)
	if err != nil {
		t.Fatalf("marshal live capacity probe base: %v", err)
	}
	padding := targetBytes - len(base)
	if padding < 0 {
		t.Fatalf("live capacity probe base=%d bytes exceeds target=%d", len(base), targetBytes)
	}
	messages[1].Content = strings.Repeat("x", padding)
	framed, err := json.Marshal(messages)
	if err != nil {
		t.Fatalf("marshal live capacity probe: %v", err)
	}
	if len(framed) != targetBytes {
		t.Fatalf("live capacity probe=%d bytes, want %d", len(framed), targetBytes)
	}
	var roundTrip []aiclient.Message
	if err := json.Unmarshal(framed, &roundTrip); err != nil {
		t.Fatalf("decode generated live capacity probe: %v", err)
	}
	reframed, err := json.Marshal(roundTrip)
	if err != nil || !bytes.Equal(reframed, framed) {
		t.Fatal("generated live capacity probe did not round-trip byte-identically")
	}
	return messages
}

func loadLiveReportProfile(t *testing.T, repoRoot string) *aiclient.ModelProfile {
	t.Helper()
	loader, err := profile.NewLoader(profile.Options{
		Path:         filepath.Join(repoRoot, "config", "ai-profiles.yaml"),
		PollInterval: -1,
	})
	if err != nil {
		t.Fatalf("load tracked profile catalog: %v", err)
	}
	t.Cleanup(loader.Close)
	reportProfile, err := loader.Resolve("report.generate.default")
	if err != nil {
		t.Fatalf("resolve report profile: %v", err)
	}
	return reportProfile
}

func loadLiveReportProvider(t *testing.T, repoRoot string, reportProfile *aiclient.ModelProfile) *openaicompatible.Adapter {
	t.Helper()
	registry, err := providerregistry.NewLoader(providerregistry.Options{
		Path:         filepath.Join(repoRoot, "config", "ai-providers.yaml"),
		PollInterval: -1,
	})
	if err != nil {
		t.Fatalf("load provider registry: %v", err)
	}
	t.Cleanup(registry.Close)
	entry, ok := registry.Provider(reportProfile.Default.ProviderRef)
	if !ok {
		t.Fatalf("provider ref %q is absent", reportProfile.Default.ProviderRef)
	}
	resolved, err := providerregistry.ResolveProviderEntry(entry, "dev", environmentSecretSource{})
	if err != nil {
		t.Fatal("live provider secrets are unavailable; details redacted")
	}
	adapter, err := openaicompatible.New(openaicompatible.Options{Provider: resolved})
	if err != nil {
		t.Fatal("construct live provider adapter failed; details redacted")
	}
	return adapter
}

func mustReadLiveProbeFixture(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read live probe fixture %s: %v", filepath.Base(path), err)
	}
	return body
}

func logLiveProbeEvidence(t *testing.T, name string, reportProfile *aiclient.ModelProfile, meta aiclient.AICallMeta, finishReason string) {
	t.Helper()
	reserved := meta.InputTokens + reportProviderFrameReserve + reportProfile.MaxTokens
	t.Logf("REPORT_PROFILE_LIVE_TOKEN_PASS probe=%s input_tokens=%d output_tokens=%d finish_reason=%s reserved_tokens=%d context_window=%d", name, meta.InputTokens, meta.OutputTokens, finishReason, reserved, reportProfile.ContextWindowTokens)
}
