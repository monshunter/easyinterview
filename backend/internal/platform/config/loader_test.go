package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

type mapSecret map[string]string

func (m mapSecret) Get(name string) (string, error) {
	v, ok := m[name]
	if !ok {
		return "", config.ErrSecretMissing
	}
	return v, nil
}

func writeYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestLoaderFourLayerMerge(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	envFile := filepath.Join(dir, "dev.yaml")
	writeYAML(t, base, "log:\n  level: info\napp:\n  listenAddr: \":8080\"\nauth:\n  sessionCookieName: ei_session\n")
	writeYAML(t, envFile, "log:\n  level: debug\n")

	t.Setenv("APP_LISTEN_ADDR", ":9090")

	loader, err := config.Load(config.Options{
		AppEnv:         "dev",
		ConfigDir:      dir,
		EnvBindings:    map[string]string{"APP_LISTEN_ADDR": "app.listenAddr"},
		SecretBindings: map[string]string{"objectStorage.secretKey": "OBJECT_STORAGE_SECRET_KEY"},
		SecretSource:   mapSecret{"OBJECT_STORAGE_SECRET_KEY": "runtime-secret-value"},
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := loader.GetString("app.listenAddr"); got != ":9090" {
		t.Errorf("env override missed: %q", got)
	}
	if got := loader.GetString("log.level"); got != "debug" {
		t.Errorf("env-yaml override missed: %q", got)
	}
	if got := loader.GetString("auth.sessionCookieName"); got != "ei_session" {
		t.Errorf("default base missed: %q", got)
	}
	if got := loader.GetSecret("objectStorage.secretKey").Reveal(); got != "runtime-secret-value" {
		t.Errorf("runtime secret missed: %q", got)
	}
}

func TestLoaderRespectsAppEnvFile(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), "log:\n  level: info\n")
	writeYAML(t, filepath.Join(dir, "prod.yaml"), "log:\n  level: warn\n")

	loader, err := config.Load(config.Options{AppEnv: "prod", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loader.GetString("log.level"); got != "warn" {
		t.Errorf("prod override missed: %q", got)
	}
}
