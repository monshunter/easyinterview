package registry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRegistryClientLoadsAllBaselines(t *testing.T) {
	t.Parallel()
	prompts, rubrics := repoConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	// 13 feature_keys * canonical multi coordinate.
	if got := client.SnapshotSize(); got != 11 {
		t.Fatalf("SnapshotSize: want 11, got %d", got)
	}
}

func TestNewRegistryClientRequiresDirs(t *testing.T) {
	t.Parallel()
	if _, err := NewRegistryClient(RegistryOptions{RubricsDir: "/tmp"}); err == nil {
		t.Error("missing PromptsDir must error")
	}
	if _, err := NewRegistryClient(RegistryOptions{PromptsDir: "/tmp"}); err == nil {
		t.Error("missing RubricsDir must error")
	}
}

func TestNewRegistryClientRejectsOrphanFeatureKey(t *testing.T) {
	t.Parallel()
	prompts, rubrics := tempBaselineCopy(t)

	// Remove the rubric for one feature_key — startup must fail because the
	// prompt has no matching rubric.
	if err := os.RemoveAll(filepath.Join(rubrics, "target.import.parse")); err != nil {
		t.Fatalf("remove rubric: %v", err)
	}

	_, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err == nil {
		t.Fatalf("expected orphan-feature-key error, got nil")
	}
	if !errors.Is(err, err) || !strings.Contains(err.Error(), "no rubric") {
		t.Fatalf("expected 'no rubric' message, got: %v", err)
	}
}
