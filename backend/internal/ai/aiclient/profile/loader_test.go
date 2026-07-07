package profile_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
)

const sampleProfile = `name: practice.followup.default
capability: chat
status: active
default:
  provider_ref: unit-test-stub
  model: stub-chat-1
fallback:
  - provider_ref: unit-test-stub
    model: stub-chat-1
    when: [timeout]
timeout_ms: 5000
max_tokens: 1024
rate_limit:
  rps: 5
  tpm: 60000
route: practice.followup
version: 1.0.0
`

func catalog(profiles ...string) string {
	var b strings.Builder
	b.WriteString("profiles:\n")
	for _, p := range profiles {
		for i, line := range strings.Split(strings.TrimSuffix(p, "\n"), "\n") {
			if i == 0 {
				b.WriteString("  - ")
			} else {
				b.WriteString("    ")
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func writeProfileCatalog(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "ai-profiles.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
	return path
}

func TestLoaderResolvesParsedProfile(t *testing.T) {
	path := writeProfileCatalog(t, catalog(sampleProfile))
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	p, err := loader.Resolve("practice.followup.default")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.Capability != aiclient.CapabilityChat {
		t.Fatalf("expected capability=chat, got %q", p.Capability)
	}
	if p.Status != aiclient.ProfileStatusActive {
		t.Fatalf("expected status=active, got %q", p.Status)
	}
	if p.Default.ProviderRef != "unit-test-stub" || p.Default.Model != "stub-chat-1" {
		t.Fatalf("unexpected default: %+v", p.Default)
	}
	if len(p.Fallback) != 1 || len(p.Fallback[0].When) != 1 || p.Fallback[0].When[0] != "timeout" {
		t.Fatalf("fallback not parsed: %+v", p.Fallback)
	}
	if p.TimeoutMs != 5000 || p.MaxTokens != 1024 {
		t.Fatalf("limits not parsed: timeout=%d max=%d", p.TimeoutMs, p.MaxTokens)
	}
	if p.Route != "practice.followup" || p.Version != "1.0.0" {
		t.Fatalf("route/version not parsed: %+v", p)
	}
}

func TestLoaderAcceptsSTTAsReservedCapability(t *testing.T) {
	path := writeProfileCatalog(t, catalog(`name: voice.transcription.reserved
capability: stt
status: unsupported
unsupported_reason: "STT adapter is not active in this build"
default:
  provider_ref: unit-test-stub
  model: stub-stt-1
timeout_ms: 5000
version: 1.0.0
`))
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	p, err := loader.Resolve("voice.transcription.reserved")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.Capability != aiclient.CapabilitySTT {
		t.Fatalf("expected capability=stt, got %q", p.Capability)
	}
}

func TestLoaderRejectsNonCurrentProfileSchemaKeys(t *testing.T) {
	cases := map[string]string{
		"task-type": `name: sample
task_type: chat
status: active
default:
  provider_ref: unit-test-stub
  model: m
timeout_ms: 1
version: 1
`,
		"default-provider": `name: sample
capability: chat
status: active
default:
  provider: stub
  model: m
timeout_ms: 1
version: 1
`,
		"fallback-provider-trigger": `name: sample
capability: chat
status: active
default:
  provider_ref: unit-test-stub
  model: m
fallback:
  - provider: stub
    model: m
    trigger: timeout
timeout_ms: 1
version: 1
`,
	}
	for label, body := range cases {
		t.Run(label, func(t *testing.T) {
			path := writeProfileCatalog(t, catalog(body))
			_, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
			if err == nil {
				t.Fatalf("expected non-current schema key to be rejected")
			}
			if !strings.Contains(err.Error(), "non-current schema key") {
				t.Fatalf("expected non-current schema key error, got %v", err)
			}
		})
	}
}

func TestLoaderRejectsUnknownCatalogTopLevelField(t *testing.T) {
	path := writeProfileCatalog(t, "profile:\n  name: sample\n")
	_, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err == nil {
		t.Fatal("expected unknown top-level field to be rejected")
	}
	if !strings.Contains(err.Error(), "unsupported top-level field") {
		t.Fatalf("expected top-level field error, got %v", err)
	}
}

func TestLoaderResolveUnknownProfile(t *testing.T) {
	path := writeProfileCatalog(t, catalog(sampleProfile))
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()
	if _, err := loader.Resolve("does-not-exist"); err == nil {
		t.Fatalf("expected error for unknown profile")
	}
}

func TestLoaderMissingRequiredFields(t *testing.T) {
	cases := map[string]string{
		"missing-name":               "capability: chat\nstatus: active\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-capability":         "name: x\nstatus: active\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-provider-ref":       "name: x\ncapability: chat\nstatus: active\ndefault:\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-model":              "name: x\ncapability: chat\nstatus: active\ndefault:\n  provider_ref: unit-test-stub\ntimeout_ms: 1\nversion: 1\n",
		"bad-capability":             "name: x\ncapability: image\nstatus: active\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-status":             "name: x\ncapability: chat\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"bad-status":                 "name: x\ncapability: chat\nstatus: inactive\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"unsupported-without-reason": "name: x\ncapability: stt\nstatus: unsupported\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"disabled-without-reason":    "name: x\ncapability: stt\nstatus: disabled\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"non-positive-to":            "name: x\ncapability: chat\nstatus: active\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 0\nversion: 1\n",
		"missing-version":            "name: x\ncapability: chat\nstatus: active\ndefault:\n  provider_ref: unit-test-stub\n  model: m\ntimeout_ms: 1\n",
	}
	for label, body := range cases {
		t.Run(label, func(t *testing.T) {
			path := writeProfileCatalog(t, catalog(body))
			_, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
			if err == nil {
				t.Fatalf("expected error for %s", label)
			}
			if !strings.Contains(err.Error(), path) {
				t.Fatalf("expected file path in error, got %v", err)
			}
			if !strings.Contains(err.Error(), ":line ") {
				t.Fatalf("expected line number in error, got %v", err)
			}
		})
	}
}

func TestLoaderReloadPicksUpEdits(t *testing.T) {
	path := writeProfileCatalog(t, catalog(sampleProfile))
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	original, err := loader.Resolve("practice.followup.default")
	if err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	if original.TimeoutMs != 5000 {
		t.Fatalf("expected timeout=5000, got %d", original.TimeoutMs)
	}

	updated := `name: practice.followup.default
capability: chat
status: active
default:
  provider_ref: unit-test-stub
  model: stub-chat-2
timeout_ms: 9000
version: 1.1.0
`
	if err := os.WriteFile(path, []byte(catalog(updated)), 0o600); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	if err := loader.Reload(context.Background()); err != nil {
		t.Fatalf("Reload: %v", err)
	}

	got, err := loader.Resolve("practice.followup.default")
	if err != nil {
		t.Fatalf("post-reload Resolve: %v", err)
	}
	if got.Default.Model != "stub-chat-2" || got.TimeoutMs != 9000 || got.Version != "1.1.0" {
		t.Fatalf("reload did not propagate: %+v", got)
	}
	if original.TimeoutMs != 5000 {
		t.Fatalf("captured pre-reload pointer mutated: timeout=%d", original.TimeoutMs)
	}
}

func TestLoaderReloadConvergesUnderHotReloadSLA(t *testing.T) {
	path := writeProfileCatalog(t, catalog(sampleProfile))
	loader, err := profile.NewLoader(profile.Options{
		Path:         path,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	updated := `name: practice.followup.default
capability: chat
status: active
default:
  provider_ref: unit-test-stub
  model: stub-chat-2
timeout_ms: 8000
version: 1.2.0
`
	if err := os.WriteFile(path, []byte(catalog(updated)), 0o600); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, err := loader.Resolve("practice.followup.default")
		if err == nil && got.Version == "1.2.0" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("polling reload did not converge before deadline")
}

func TestLoaderPollLoopReportsReloadWarningAndKeepsOldSnapshot(t *testing.T) {
	path := writeProfileCatalog(t, catalog(sampleProfile))
	warnings := make(chan error, 1)
	loader, err := profile.NewLoader(profile.Options{
		Path:         path,
		PollInterval: 50 * time.Millisecond,
		OnWarn: func(err error) {
			warnings <- err
		},
	})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	if err := os.WriteFile(path, []byte(catalog(`name: invalid
capability: image
status: active
default:
  provider_ref: unit-test-stub
  model: m
timeout_ms: 1
version: 1
`)), 0o600); err != nil {
		t.Fatalf("rewrite invalid profile: %v", err)
	}

	select {
	case err := <-warnings:
		if err == nil || !strings.Contains(err.Error(), "unsupported capability") {
			t.Fatalf("unexpected warning error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected reload warning before deadline")
	}

	got, err := loader.Resolve("practice.followup.default")
	if err != nil {
		t.Fatalf("previous snapshot should remain resolvable: %v", err)
	}
	if got.Version != "1.0.0" {
		t.Fatalf("failed reload polluted previous snapshot: %+v", got)
	}
}

func TestLoaderRejectsDuplicateNames(t *testing.T) {
	dup := `name: practice.followup.default
capability: chat
status: active
default:
  provider_ref: unit-test-stub
  model: stub-chat-1
timeout_ms: 1
version: 1
`
	path := writeProfileCatalog(t, catalog(dup, dup))
	if _, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1}); err == nil {
		t.Fatal("expected duplicate-name error")
	}
}

func TestLoaderNamesIsSorted(t *testing.T) {
	path := writeProfileCatalog(t, catalog(sampleProfile, `name: review.report.default
capability: chat
status: active
default:
  provider_ref: unit-test-stub
  model: stub-chat-1
timeout_ms: 1
version: 1
`))
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	got := loader.Names()
	want := []string{"practice.followup.default", "review.report.default"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestLoaderRejectsNonCurrentDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "practice.followup.default.yaml"), []byte(sampleProfile), 0o600); err != nil {
		t.Fatalf("WriteFile non-current profile: %v", err)
	}
	_, err := profile.NewLoader(profile.Options{Path: dir, PollInterval: -1})
	if err == nil {
		t.Fatal("expected non-current profile directory to be rejected")
	}
	if !strings.Contains(err.Error(), "read") && !strings.Contains(err.Error(), "open") {
		t.Fatalf("expected file-path load error, got %v", err)
	}
}

func TestLoaderAcceptsTTSAsReservedCapability(t *testing.T) {
	path := writeProfileCatalog(t, catalog(`name: voice.tts.reserved
capability: tts
status: unsupported
unsupported_reason: "TTS adapter is not active in this build"
default:
  provider_ref: unit-test-stub
  model: stub-tts-1
timeout_ms: 5000
version: 1.0.0
`))
	loader, err := profile.NewLoader(profile.Options{Path: path, PollInterval: -1})
	if err != nil {
		t.Fatalf("expected tts capability to be accepted: %v", err)
	}
	defer loader.Close()

	p, err := loader.Resolve("voice.tts.reserved")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.Capability != aiclient.CapabilityTts {
		t.Fatalf("expected capability=tts, got %q", p.Capability)
	}
}
