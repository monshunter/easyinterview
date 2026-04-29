package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkerConfigProdReadsCanonicalSecrets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "config.yaml"), `
app:
  env: dev
worker:
  listenAddr: ":8081"
auth:
  sessionCookieName: ei_session
  sessionCookieSecret: ""
  challengeTokenPepper: ""
ai:
  gatewayBaseURL: ""
  gatewayApiKey: ""
email:
  provider: ""
  providerApiKey: ""
featureFlag:
  source: file
  posthogSelfHosted: false
  posthogProjectApiKey: ""
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
`)
	writeFile(t, filepath.Join(dir, "prod.yaml"), `
featureFlag:
  source: posthog
  posthogSelfHosted: true
`)

	for key, value := range map[string]string{
		"SESSION_COOKIE_SECRET":       "session-secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_GATEWAY_BASE_URL":         "https://gateway.example",
		"AI_GATEWAY_API_KEY":          "ai-key",
		"EMAIL_PROVIDER":              "smtp",
		"EMAIL_PROVIDER_API_KEY":      "email-key",
		"POSTHOG_HOST":                "https://posthog.example",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
		"POSTHOG_SELF_HOSTED":         "true",
	} {
		t.Setenv(key, value)
	}

	loader, err := loadWorkerConfig("prod", dir)
	if err != nil {
		t.Fatalf("loadWorkerConfig: %v", err)
	}
	if err := loader.Validate(); err != nil {
		t.Fatalf("Validate with complete prod env: %v", err)
	}
	if got := loader.GetSecret("ai.gatewayApiKey").Reveal(); got != "ai-key" {
		t.Fatalf("worker did not load AI gateway secret; got %q", got)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
