package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

func newProdLoader(t *testing.T, secrets mapSecret) *config.Loader {
	return newProdLoaderWithGatewayBaseURL(t, secrets, "")
}

func newProdLoaderWithGatewayBaseURL(t *testing.T, secrets mapSecret, gatewayBaseURL string) *config.Loader {
	t.Helper()
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
auth:
  sessionCookieName: ei_session
  sessionCookieSecret: ""
  challengeTokenPepper: ""
ai:
  gatewayBaseURL: "`+gatewayBaseURL+`"
  gatewayApiKey: ""
email:
  provider: ""
  providerApiKey: ""
featureFlag:
  source: posthog
  posthogSelfHosted: true
  posthogHost: ""
  posthogProjectApiKey: ""
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
`)
	loader, err := config.Load(config.Options{
		AppEnv:         "prod",
		ConfigDir:      dir,
		EnvBindings:    config.DefaultEnvBindings(),
		SecretBindings: config.DefaultSecretBindings(),
		SecretSource:   secrets,
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return loader
}

func setCompleteProdRuntimeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_LISTEN_ADDR", ":8080")
	t.Setenv("WORKER_LISTEN_ADDR", ":8081")
	t.Setenv("DATABASE_URL", "postgres://prod:secret@db.internal:5432/easyinterview?sslmode=require")
	t.Setenv("REDIS_URL", "redis://redis.internal:6379/0")
	t.Setenv("OBJECT_STORAGE_ENDPOINT", "https://s3.internal")
	t.Setenv("OBJECT_STORAGE_BUCKET", "easyinterview-prod")
	t.Setenv("OBJECT_STORAGE_ACCESS_KEY", "object-access")
	t.Setenv("OBJECT_STORAGE_SECRET_KEY", "object-secret")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("AI_GATEWAY_BASE_URL", "https://gateway")
	t.Setenv("AI_MODEL_PROFILE_PATH", "/etc/easyinterview/ai-profiles")
	t.Setenv("FEATURE_FLAG_SOURCE", "posthog")
	t.Setenv("POSTHOG_HOST", "https://posthog")
	t.Setenv("EMAIL_PROVIDER", "ses")
}

func TestValidateProdMissingSecretFailsFast(t *testing.T) {
	loader := newProdLoader(t, mapSecret{})
	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error in prod with missing secrets")
	}
	msg := err.Error()
	for _, key := range []string{"SESSION_COOKIE_SECRET", "AUTH_CHALLENGE_TOKEN_PEPPER", "AI_GATEWAY_API_KEY", "EMAIL_PROVIDER_API_KEY"} {
		if !strings.Contains(msg, key) {
			t.Errorf("error missing key %s: %s", key, msg)
		}
	}
	if !strings.Contains(msg, "missing required secret") {
		t.Errorf("error message format: %s", msg)
	}
}

func TestValidateProdAllSecretsPasses(t *testing.T) {
	setCompleteProdRuntimeEnv(t)
	loader := newProdLoaderWithGatewayBaseURL(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_GATEWAY_API_KEY":          "key",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	}, "https://gateway")
	if err := loader.Validate(); err != nil {
		t.Errorf("unexpected validate error: %v", err)
	}
}

func TestValidateProdRejectsDevDefaultDeploymentDependencies(t *testing.T) {
	loader := newProdLoaderWithGatewayBaseURL(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_GATEWAY_API_KEY":          "key",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	}, "https://gateway")

	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error when prod uses dev default deployment dependencies")
	}
	msg := err.Error()
	for _, key := range []string{
		"DATABASE_URL",
		"REDIS_URL",
		"OBJECT_STORAGE_ENDPOINT",
		"OBJECT_STORAGE_BUCKET",
		"OBJECT_STORAGE_ACCESS_KEY",
		"OBJECT_STORAGE_SECRET_KEY",
		"POSTHOG_HOST",
		"EMAIL_PROVIDER",
	} {
		if !strings.Contains(msg, key) {
			t.Errorf("error missing key %s: %s", key, msg)
		}
	}
}

func TestValidateProdMissingAIBaseURLFailsFast(t *testing.T) {
	loader := newProdLoader(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_GATEWAY_API_KEY":          "key",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	})

	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error when AI_GATEWAY_BASE_URL is missing")
	}
	if !strings.Contains(err.Error(), "AI_GATEWAY_BASE_URL") {
		t.Fatalf("error must mention AI_GATEWAY_BASE_URL: %v", err)
	}
}

func TestValidateTestEnvAllowsMissingAIAndSession(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
auth:
  sessionCookieName: ei_session
ai:
  gatewayBaseURL: ""
  gatewayApiKey: ""
featureFlag:
  source: file
  filePath: ""
`)
	loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := loader.Validate(); err != nil {
		t.Errorf("APP_ENV=test must allow missing AI/Email/Session secrets, got: %v", err)
	}
}

func TestValidateStagingPostHogSelfHostedFalseFailsFast(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
auth:
  sessionCookieName: ei_session
ai:
  gatewayBaseURL: "https://gateway"
featureFlag:
  source: posthog
  posthogSelfHosted: false
  posthogHost: "https://posthog"
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
`)
	loader, err := config.Load(config.Options{
		AppEnv:    "staging",
		ConfigDir: dir,
		SecretBindings: map[string]string{
			"auth.sessionCookieSecret":         "SESSION_COOKIE_SECRET",
			"auth.challengeTokenPepper":        "AUTH_CHALLENGE_TOKEN_PEPPER",
			"ai.gatewayApiKey":                 "AI_GATEWAY_API_KEY",
			"email.providerApiKey":             "EMAIL_PROVIDER_API_KEY",
			"featureFlag.posthogProjectApiKey": "POSTHOG_PROJECT_API_KEY",
		},
		SecretSource: mapSecret{
			"SESSION_COOKIE_SECRET":       "x",
			"AUTH_CHALLENGE_TOKEN_PEPPER": "x",
			"AI_GATEWAY_API_KEY":          "x",
			"EMAIL_PROVIDER_API_KEY":      "x",
			"POSTHOG_PROJECT_API_KEY":     "x",
		},
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	err = loader.Validate()
	if err == nil {
		t.Fatal("expected validate error when staging POSTHOG_SELF_HOSTED=false")
	}
	if !strings.Contains(err.Error(), "POSTHOG_SELF_HOSTED") {
		t.Errorf("error must mention POSTHOG_SELF_HOSTED: %v", err)
	}
}

func TestValidateAsyncQueueWeightsFailsFastWhenMissingOrNonPositive(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
auth:
  sessionCookieName: ei_session
ai:
  gatewayBaseURL: "https://gateway"
async:
  queueWeights:
    critical: 0
    default: 3
    low: 1
featureFlag:
  source: file
  filePath: "./feature-flags.yaml"
`)
	loader, err := config.Load(config.Options{AppEnv: "dev", ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := loader.Validate(); err == nil || !strings.Contains(err.Error(), "async.queueWeights") {
		t.Errorf("expected async.queueWeights validation error, got: %v", err)
	}
}
