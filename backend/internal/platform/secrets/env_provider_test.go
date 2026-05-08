package secrets_test

import (
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
)

func TestEnvSecretSourceGetReadsEnvironment(t *testing.T) {
	t.Setenv("EASYINTERVIEW_SECRET_TEST", "runtime-secret")

	got, err := (secrets.EnvSecretSource{}).Get("EASYINTERVIEW_SECRET_TEST")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "runtime-secret" {
		t.Fatalf("secret = %q, want runtime-secret", got)
	}
}

func TestEnvSecretSourceGetMissingReturnsConfigSentinel(t *testing.T) {
	t.Setenv("EASYINTERVIEW_SECRET_TEST_MISSING", "")

	_, err := (secrets.EnvSecretSource{}).Get("EASYINTERVIEW_SECRET_TEST_MISSING")
	if !errors.Is(err, config.ErrSecretMissing) {
		t.Fatalf("expected ErrSecretMissing, got %v", err)
	}
}
