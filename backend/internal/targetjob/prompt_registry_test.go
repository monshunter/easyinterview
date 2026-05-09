package targetjob_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// repoConfigRoots walks upward from the test wd to locate the backend
// go.mod, then returns the absolute paths to the in-repo config/prompts
// and config/rubrics roots.
func repoConfigRoots(t *testing.T) (string, string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("could not locate backend go.mod from %s", wd)
			return "", ""
		}
		dir = parent
	}
	repoRoot := filepath.Dir(dir)
	return filepath.Join(repoRoot, "config", "prompts"),
		filepath.Join(repoRoot, "config", "rubrics")
}

func TestRegistryAdapterMapsAllSevenFields(t *testing.T) {
	t.Parallel()

	prompts, rubrics := repoConfigRoots(t)
	client, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}

	adapter := targetjob.NewRegistryAdapter(client)
	got, err := adapter.Resolve(context.Background(), targetjob.FeatureKeyTargetImportParse, "en")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	// Spec D-4 + plan §3.1 lock the seven targetjob fields. List each one
	// explicitly so future drift fails the test rather than slipping past
	// a struct-level reflect comparison.
	if got.PromptVersion != "v0.1.0" {
		t.Errorf("PromptVersion: want v0.1.0, got %q", got.PromptVersion)
	}
	if got.RubricVersion != "v0.1.0" {
		t.Errorf("RubricVersion: want v0.1.0, got %q", got.RubricVersion)
	}
	if got.ModelProfileName != "target.import.default" {
		t.Errorf("ModelProfileName: want target.import.default, got %q", got.ModelProfileName)
	}
	if got.DataSourceVersion == "" {
		t.Errorf("DataSourceVersion must be populated; got empty string")
	}
	if got.FeatureFlag != "none" {
		t.Errorf("FeatureFlag: want 'none', got %q", got.FeatureFlag)
	}
	if got.UserMessageTemplate == "" {
		t.Errorf("UserMessageTemplate must be populated for plan 001 baseline")
	}
	// SystemMessage may legitimately be empty in plan 001 (the body lives
	// entirely in UserMessageTemplate). Just freeze the field shape.
	_ = got.SystemMessage
}

func TestRegistryAdapterRejectsNilClient(t *testing.T) {
	t.Parallel()
	adapter := targetjob.NewRegistryAdapter(nil)
	if adapter != nil {
		t.Fatalf("NewRegistryAdapter(nil) must return nil, got %+v", adapter)
	}
	// A zero-value adapter still satisfies PromptRegistryClient and must
	// fail-closed on Resolve so wiring bugs surface at call time.
	var zero *targetjob.RegistryAdapter
	if _, err := zero.Resolve(context.Background(), targetjob.FeatureKeyTargetImportParse, "en"); !errors.Is(err, targetjob.ErrPromptUnsupported) {
		t.Fatalf("nil adapter Resolve: want ErrPromptUnsupported, got %v", err)
	}
}

func TestRegistryAdapterRejectsUnknownFeatureKey(t *testing.T) {
	t.Parallel()
	prompts, rubrics := repoConfigRoots(t)
	client, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	adapter := targetjob.NewRegistryAdapter(client)

	if _, err := adapter.Resolve(context.Background(), "no.such.feature", "en"); !errors.Is(err, targetjob.ErrPromptUnsupported) {
		t.Fatalf("unknown feature_key: want ErrPromptUnsupported, got %v", err)
	}
	if _, err := adapter.Resolve(context.Background(), "", "en"); !errors.Is(err, targetjob.ErrPromptUnsupported) {
		t.Fatalf("empty feature_key: want ErrPromptUnsupported, got %v", err)
	}
	if _, err := adapter.Resolve(context.Background(), targetjob.FeatureKeyTargetImportParse, ""); !errors.Is(err, targetjob.ErrPromptUnsupported) {
		t.Fatalf("empty language: want ErrPromptUnsupported, got %v", err)
	}
}
