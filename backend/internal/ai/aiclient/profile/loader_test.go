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
task_type: chat
default:
  provider: stub
  model: stub-chat-1
fallback:
  - provider: stub
    model: stub-chat-1
    trigger: timeout
timeout_ms: 5000
max_tokens: 1024
rate_limit:
  rps: 5
  tpm: 60000
gateway_route: practice.followup
version: 1.0.0
`

func writeProfileDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("WriteFile %s: %v", path, err)
		}
	}
	return dir
}

func TestLoaderResolvesParsedProfile(t *testing.T) {
	dir := writeProfileDir(t, map[string]string{"practice.yaml": sampleProfile})
	loader, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	p, err := loader.Resolve("practice.followup.default")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.TaskType != aiclient.TaskTypeChat {
		t.Fatalf("expected task_type=chat, got %q", p.TaskType)
	}
	if p.Default.Provider != "stub" || p.Default.Model != "stub-chat-1" {
		t.Fatalf("unexpected default: %+v", p.Default)
	}
	if len(p.Fallback) != 1 || p.Fallback[0].Trigger != "timeout" {
		t.Fatalf("fallback not parsed: %+v", p.Fallback)
	}
	if p.TimeoutMs != 5000 || p.MaxTokens != 1024 {
		t.Fatalf("limits not parsed: timeout=%d max=%d", p.TimeoutMs, p.MaxTokens)
	}
	if p.GatewayRoute != "practice.followup" || p.Version != "1.0.0" {
		t.Fatalf("route/version not parsed: %+v", p)
	}
}

func TestLoaderAcceptsSTTAsReservedTaskType(t *testing.T) {
	dir := writeProfileDir(t, map[string]string{"stt.yaml": `name: voice.transcription.reserved
task_type: stt
default:
  provider: stub
  model: stub-stt-1
timeout_ms: 5000
version: 1.0.0
`})
	loader, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	p, err := loader.Resolve("voice.transcription.reserved")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.TaskType != aiclient.TaskTypeSTT {
		t.Fatalf("expected task_type=stt, got %q", p.TaskType)
	}
}

func TestLoaderResolveUnknownProfile(t *testing.T) {
	dir := writeProfileDir(t, map[string]string{"a.yaml": sampleProfile})
	loader, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
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
		"missing-name":     "task_type: chat\ndefault:\n  provider: stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-task":     "name: x\ndefault:\n  provider: stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-provider": "name: x\ntask_type: chat\ndefault:\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"missing-model":    "name: x\ntask_type: chat\ndefault:\n  provider: stub\ntimeout_ms: 1\nversion: 1\n",
		"bad-task-type":    "name: x\ntask_type: image\ndefault:\n  provider: stub\n  model: m\ntimeout_ms: 1\nversion: 1\n",
		"non-positive-to":  "name: x\ntask_type: chat\ndefault:\n  provider: stub\n  model: m\ntimeout_ms: 0\nversion: 1\n",
		"missing-version":  "name: x\ntask_type: chat\ndefault:\n  provider: stub\n  model: m\ntimeout_ms: 1\n",
	}
	for label, body := range cases {
		t.Run(label, func(t *testing.T) {
			dir := writeProfileDir(t, map[string]string{"p.yaml": body})
			_, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
			if err == nil {
				t.Fatalf("expected error for %s", label)
			}
			if !strings.Contains(err.Error(), filepath.Join(dir, "p.yaml")) {
				t.Fatalf("expected file path in error, got %v", err)
			}
			if !strings.Contains(err.Error(), ":line ") {
				t.Fatalf("expected line number in error, got %v", err)
			}
		})
	}
}

func TestLoaderReloadPicksUpEdits(t *testing.T) {
	dir := writeProfileDir(t, map[string]string{"p.yaml": sampleProfile})
	loader, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
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
task_type: chat
default:
  provider: stub
  model: stub-chat-2
timeout_ms: 9000
version: 1.1.0
`
	if err := os.WriteFile(filepath.Join(dir, "p.yaml"), []byte(updated), 0o600); err != nil {
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
	dir := writeProfileDir(t, map[string]string{"p.yaml": sampleProfile})
	loader, err := profile.NewLoader(profile.Options{
		Dir:          dir,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	updated := `name: practice.followup.default
task_type: chat
default:
  provider: stub
  model: stub-chat-2
timeout_ms: 8000
version: 1.2.0
`
	if err := os.WriteFile(filepath.Join(dir, "p.yaml"), []byte(updated), 0o600); err != nil {
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

func TestLoaderRejectsDuplicateNames(t *testing.T) {
	dup := `name: practice.followup.default
task_type: chat
default:
  provider: stub
  model: stub-chat-1
timeout_ms: 1
version: 1
`
	dir := writeProfileDir(t, map[string]string{
		"a.yaml": dup,
		"b.yaml": dup,
	})
	if _, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1}); err == nil {
		t.Fatal("expected duplicate-name error")
	}
}

func TestLoaderNamesIsSorted(t *testing.T) {
	dir := writeProfileDir(t, map[string]string{
		"a.yaml": sampleProfile,
		"b.yaml": `name: review.report.default
task_type: chat
default:
  provider: stub
  model: stub-chat-1
timeout_ms: 1
version: 1
`,
	})
	loader, err := profile.NewLoader(profile.Options{Dir: dir, PollInterval: -1})
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
