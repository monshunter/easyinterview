package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

type e2eP0MissingSecretSource struct{}

func (e2eP0MissingSecretSource) Get(string) (string, error) {
	return "", config.ErrSecretMissing
}

func TestE2EP0ConfigPreflightLoadsFunnelProfilesWithoutProviderSecrets(t *testing.T) {
	t.Setenv("AI_PROVIDER_BASE_URL", "")
	t.Setenv("AI_PROVIDER_API_KEY", "")

	loader := testLoader(t)
	if loader.AppEnv() != aiclient.AppEnvTest {
		t.Fatalf("loader AppEnv=%q, want test", loader.AppEnv())
	}
	root := scenarioRepoRoot(t)
	registryPath := repoRelativePath(root, loader.GetString("ai.providerRegistryPath"))
	profilePath := repoRelativePath(root, loader.GetString("ai.modelProfilePath"))

	registryLoader, err := providerregistry.NewLoader(providerregistry.Options{
		Path:         registryPath,
		PollInterval: -1,
	})
	if err != nil {
		t.Fatalf("providerregistry.NewLoader: %v", err)
	}
	defer registryLoader.Close()

	profileLoader, err := profile.NewLoader(profile.Options{
		Path:         profilePath,
		PollInterval: -1,
	})
	if err != nil {
		t.Fatalf("profile.NewLoader: %v", err)
	}
	defer profileLoader.Close()

	required := map[string]aiclient.Capability{
		"resume.parse.default":            aiclient.CapabilityChat,
		"target.import.default":           aiclient.CapabilityChat,
		"practice.first_question.default": aiclient.CapabilityChat,
		"practice.followup.default":       aiclient.CapabilityChat,
		"practice.turn_observe.default":   aiclient.CapabilityChat,
		"report.generate.default":         aiclient.CapabilityChat,
	}
	for name, capability := range required {
		t.Run(name, func(t *testing.T) {
			resolved, err := profileLoader.Resolve(name)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if resolved.Status != aiclient.ProfileStatusActive {
				t.Fatalf("profile status=%q, want active", resolved.Status)
			}
			if resolved.Capability != capability {
				t.Fatalf("profile capability=%q, want %q", resolved.Capability, capability)
			}
			if _, err := registryLoader.ResolveSelectedProviders(resolved, loader.AppEnv(), e2eP0MissingSecretSource{}); err != nil {
				t.Fatalf("ResolveSelectedProviders without AI_PROVIDER_* secrets: %v", err)
			}
		})
	}
}

func scenarioRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("abs cwd: %v", err)
	}
	root := wd
	for i := 0; i < 6; i++ {
		if e2eP0FileExists(filepath.Join(root, "shared", "conventions.yaml")) {
			return root
		}
		root = filepath.Dir(root)
	}
	t.Fatalf("repo root not found from %s", wd)
	return ""
}

func repoRelativePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func e2eP0FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
