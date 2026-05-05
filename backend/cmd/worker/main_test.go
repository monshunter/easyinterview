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
  providerRegistryPath: ""
  defaultProviderBaseURL: ""
  defaultProviderApiKey: ""
  modelProfilePath: ""
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
		"APP_LISTEN_ADDR":             ":8080",
		"WORKER_LISTEN_ADDR":          ":8081",
		"DATABASE_URL":                "postgres://prod:secret@db.internal:5432/easyinterview?sslmode=require",
		"REDIS_URL":                   "redis://redis.internal:6379/0",
		"OBJECT_STORAGE_ENDPOINT":     "https://s3.internal",
		"OBJECT_STORAGE_BUCKET":       "easyinterview-prod",
		"OBJECT_STORAGE_ACCESS_KEY":   "object-access",
		"OBJECT_STORAGE_SECRET_KEY":   "object-secret",
		"LOG_LEVEL":                   "info",
		"SESSION_COOKIE_SECRET":       "session-secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_PROVIDER_REGISTRY_PATH":   "/etc/easyinterview/ai-providers.yaml",
		"AI_PROVIDER_BASE_URL":        "https://provider.example",
		"AI_PROVIDER_API_KEY":         "ai-key",
		"AI_MODEL_PROFILE_PATH":       "/etc/easyinterview/ai-profiles.yaml",
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
	if got := loader.GetString("ai.providerRegistryPath"); got != "/etc/easyinterview/ai-providers.yaml" {
		t.Fatalf("worker did not load AI provider registry path; got %q", got)
	}
	if got := loader.GetString("ai.modelProfilePath"); got != "/etc/easyinterview/ai-profiles.yaml" {
		t.Fatalf("worker did not load AI model profile path; got %q", got)
	}
	if got := loader.GetSecret("ai.defaultProviderApiKey").Reveal(); got != "ai-key" {
		t.Fatalf("worker did not load AI provider secret; got %q", got)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
