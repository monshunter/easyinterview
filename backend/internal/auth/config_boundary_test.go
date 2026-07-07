package auth_test

import (
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

func TestEmailCodeConfigBoundaryConstants(t *testing.T) {
	if auth.SessionCookieName != "ei_session" {
		t.Fatalf("SessionCookieName = %q", auth.SessionCookieName)
	}
	if auth.ChallengeTTL != 5*time.Minute {
		t.Fatalf("ChallengeTTL = %s", auth.ChallengeTTL)
	}
	if auth.SessionTTL != 30*24*time.Hour {
		t.Fatalf("SessionTTL = %s", auth.SessionTTL)
	}
	if auth.RateLimitWindow != time.Minute {
		t.Fatalf("RateLimitWindow = %s", auth.RateLimitWindow)
	}
	if auth.RateLimitThreshold != 3 {
		t.Fatalf("RateLimitThreshold = %d", auth.RateLimitThreshold)
	}
	if auth.DevMailSinkName == "" {
		t.Fatal("DevMailSinkName must document the C1-owned dev sink default")
	}
}

func TestAuthConfigConsumesA4SecretsWithoutNewRuntimeKnobs(t *testing.T) {
	env := config.DefaultEnvBindings()
	for _, key := range []string{
		"SESSION_COOKIE_SECRET",
		"AUTH_CHALLENGE_TOKEN_PEPPER",
		"EMAIL_PROVIDER",
		"EMAIL_PROVIDER_API_KEY",
	} {
		if env[key] == "" {
			t.Fatalf("A4 env dictionary missing %s", key)
		}
	}
	for _, forbidden := range []string{
		"SESSION_COOKIE_NAME",
		"AUTH_CHALLENGE_TTL",
		"SESSION_TTL",
		"AUTH_RATE_LIMIT_WINDOW",
		"DEV_MAIL_SINK",
	} {
		if _, ok := env[forbidden]; ok {
			t.Fatalf("auth must not add runtime knob %s", forbidden)
		}
	}
}
