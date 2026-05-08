package secrets_test

import (
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
)

func TestEnvSecretSourceGet(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECRET", "session-secret")

	got, err := (secrets.EnvSecretSource{}).Get("SESSION_COOKIE_SECRET")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "session-secret" {
		t.Fatalf("Get = %q, want session-secret", got)
	}
}

func TestEnvSecretSourceMissing(t *testing.T) {
	_, err := (secrets.EnvSecretSource{}).Get("MISSING_SECRET")
	if !errors.Is(err, config.ErrSecretMissing) {
		t.Fatalf("Get error = %v, want ErrSecretMissing", err)
	}
}
