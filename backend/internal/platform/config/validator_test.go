package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

func newProdLoader(t *testing.T, secrets mapSecret) *config.Loader {
	return newProdLoaderWithProviderBaseURL(t, secrets, "")
}

func newProdLoaderWithProviderBaseURL(t *testing.T, secrets mapSecret, providerBaseURL string) *config.Loader {
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
  providerRegistryPath: ""
  defaultProviderBaseURL: "`+providerBaseURL+`"
  defaultProviderApiKey: ""
  modelProfilePath: ""
email:
  provider: ""
  providerApiKey: ""
featureFlag:
  source: posthog
  posthogSelfHosted: true
  posthogHost: ""
  posthogProjectApiKey: ""
objectStorage:
  provider: minio
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: 2097152
    privacyExport: 5242880
resume:
  maxActive: 10
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
  leaseTimeoutSeconds: 300
  shutdownGraceSeconds: 10
  reaperIntervalSeconds: 60
  scanIntervalSeconds: 5
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
	t.Setenv("DATABASE_URL", "postgres://prod:secret@db.internal:5432/easyinterview?sslmode=require")
	t.Setenv("REDIS_URL", "redis://redis.internal:6379/0")
	t.Setenv("OBJECT_STORAGE_ENDPOINT", "https://s3.internal")
	t.Setenv("OBJECT_STORAGE_BUCKET", "easyinterview-prod")
	t.Setenv("OBJECT_STORAGE_ACCESS_KEY", "object-access")
	t.Setenv("OBJECT_STORAGE_SECRET_KEY", "object-secret")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("AI_PROVIDER_REGISTRY_PATH", "/etc/easyinterview/ai-providers.yaml")
	t.Setenv("AI_PROVIDER_BASE_URL", "https://provider.example")
	t.Setenv("AI_MODEL_PROFILE_PATH", "/etc/easyinterview/ai-profiles.yaml")
	t.Setenv("FEATURE_FLAG_SOURCE", "posthog")
	t.Setenv("POSTHOG_HOST", "https://posthog")
	t.Setenv("EMAIL_PROVIDER", "ses")
}

func TestDefaultEnvDictionaryOmitsWorkerListenAddr(t *testing.T) {
	envBindings := config.DefaultEnvBindings()
	if _, ok := envBindings["WORKER_LISTEN_ADDR"]; ok {
		t.Fatalf("WORKER_LISTEN_ADDR must not remain in the P0 env dictionary: %+v", envBindings)
	}
	for _, path := range envBindings {
		if path == "worker.listenAddr" {
			t.Fatalf("worker.listenAddr must not remain in canonical config bindings: %+v", envBindings)
		}
	}
}

func TestDefaultAIDictionaryUsesProviderRegistryPaths(t *testing.T) {
	envBindings := config.DefaultEnvBindings()
	if got := envBindings["AI_PROVIDER_REGISTRY_PATH"]; got != "ai.providerRegistryPath" {
		t.Fatalf("AI_PROVIDER_REGISTRY_PATH binding = %q", got)
	}
	if got := envBindings["AI_PROVIDER_BASE_URL"]; got != "ai.defaultProviderBaseURL" {
		t.Fatalf("AI_PROVIDER_BASE_URL binding = %q", got)
	}
	if got := envBindings["AI_PROVIDER_API_KEY"]; got != "ai.defaultProviderApiKey" {
		t.Fatalf("AI_PROVIDER_API_KEY binding = %q", got)
	}
	if got := envBindings["AI_MODEL_PROFILE_PATH"]; got != "ai.modelProfilePath" {
		t.Fatalf("AI_MODEL_PROFILE_PATH binding = %q", got)
	}
	if got := envBindings["AI_DEBUG_PRINT_RAW_OUTPUT"]; got != "ai.debugPrintRawOutput" {
		t.Fatalf("AI_DEBUG_PRINT_RAW_OUTPUT binding = %q", got)
	}

	secretBindings := config.DefaultSecretBindings()
	if _, ok := secretBindings["ai.providerApiKey"]; ok {
		t.Fatalf("out-of-scope ai.providerApiKey secret binding must not remain: %+v", secretBindings)
	}
	if got := secretBindings["ai.defaultProviderApiKey"]; got != "AI_PROVIDER_API_KEY" {
		t.Fatalf("ai.defaultProviderApiKey secret binding = %q", got)
	}
}

func TestAIDebugPrintRawOutputEnvBindingOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
ai:
  debugPrintRawOutput: false
`)
	t.Setenv("AI_DEBUG_PRINT_RAW_OUTPUT", "true")

	loader, err := config.Load(config.Options{
		AppEnv:      "test",
		ConfigDir:   dir,
		EnvBindings: config.DefaultEnvBindings(),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loader.GetBool("ai.debugPrintRawOutput"); !got {
		t.Fatalf("AI_DEBUG_PRINT_RAW_OUTPUT env binding = %v", got)
	}
}

func TestRepoLocalConfigEnablesRawOutputDebugOnlyForLocalEnvironments(t *testing.T) {
	configDir := filepath.Clean("../../../../config")

	for _, appEnv := range []string{"dev", "test"} {
		loader, err := config.Load(config.Options{
			AppEnv:    appEnv,
			ConfigDir: configDir,
		})
		if err != nil {
			t.Fatalf("Load(%s): %v", appEnv, err)
		}
		if got := loader.GetBool("ai.debugPrintRawOutput"); !got {
			t.Fatalf("%s ai.debugPrintRawOutput = %v, want true", appEnv, got)
		}
	}

	for _, appEnv := range []string{"staging", "prod"} {
		loader, err := config.Load(config.Options{
			AppEnv:    appEnv,
			ConfigDir: configDir,
		})
		if err != nil {
			t.Fatalf("Load(%s): %v", appEnv, err)
		}
		if got := loader.GetBool("ai.debugPrintRawOutput"); got {
			t.Fatalf("%s ai.debugPrintRawOutput = %v, want false", appEnv, got)
		}
	}
}

func TestDefaultEmailDictionaryIncludesMailpitSMTPBindings(t *testing.T) {
	envBindings := config.DefaultEnvBindings()
	for key, want := range map[string]string{
		"EMAIL_PROVIDER":         "email.provider",
		"EMAIL_SMTP_HOST":        "email.smtpHost",
		"EMAIL_SMTP_PORT":        "email.smtpPort",
		"EMAIL_FROM_ADDRESS":     "email.fromAddress",
		"EMAIL_VERIFY_BASE_URL":  "email.verifyBaseURL",
		"EMAIL_PROVIDER_API_KEY": "email.providerApiKey",
	} {
		if got := envBindings[key]; got != want {
			t.Fatalf("%s binding = %q, want %q", key, got, want)
		}
	}
}

func TestValidateProdMissingSecretFailsFast(t *testing.T) {
	loader := newProdLoader(t, mapSecret{})
	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error in prod with missing secrets")
	}
	msg := err.Error()
	for _, key := range []string{"SESSION_COOKIE_SECRET", "AUTH_CHALLENGE_TOKEN_PEPPER", "EMAIL_PROVIDER_API_KEY"} {
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
	loader := newProdLoaderWithProviderBaseURL(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_PROVIDER_API_KEY":         "key",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	}, "https://provider.example")
	if err := loader.Validate(); err != nil {
		t.Errorf("unexpected validate error: %v", err)
	}
}

func TestValidateProdDoesNotRequireDefaultProviderSecretGlobally(t *testing.T) {
	setCompleteProdRuntimeEnv(t)
	t.Setenv("AI_PROVIDER_BASE_URL", "")
	loader := newProdLoader(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	})

	if err := loader.Validate(); err != nil {
		t.Fatalf("default provider secret must be required by selected provider resolution, not global config validation: %v", err)
	}
}

func TestValidateProdMissingAIRegistryPathFailsFast(t *testing.T) {
	setCompleteProdRuntimeEnv(t)
	t.Setenv("AI_PROVIDER_REGISTRY_PATH", "")
	loader := newProdLoader(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	})

	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error when AI_PROVIDER_REGISTRY_PATH is missing")
	}
	if !strings.Contains(err.Error(), "AI_PROVIDER_REGISTRY_PATH") {
		t.Fatalf("error must mention AI_PROVIDER_REGISTRY_PATH: %v", err)
	}
}

func TestDefaultProviderSecretBindingIsStillAvailableWhenRegistryReferencesIt(t *testing.T) {
	setCompleteProdRuntimeEnv(t)
	loader := newProdLoader(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_PROVIDER_API_KEY":         "provider-key",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	})

	if got := loader.GetString("ai.defaultProviderApiKey"); got != "provider-key" {
		t.Fatalf("default provider API key binding = %q", got)
	}
}

func TestValidateProdRejectsDevDefaultDeploymentDependencies(t *testing.T) {
	loader := newProdLoaderWithProviderBaseURL(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"AI_PROVIDER_API_KEY":         "key",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	}, "https://provider.example")

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

func TestValidateProdMissingAIModelProfilePathFailsFast(t *testing.T) {
	setCompleteProdRuntimeEnv(t)
	t.Setenv("AI_MODEL_PROFILE_PATH", "")
	loader := newProdLoader(t, mapSecret{
		"SESSION_COOKIE_SECRET":       "secret",
		"AUTH_CHALLENGE_TOKEN_PEPPER": "pepper",
		"EMAIL_PROVIDER_API_KEY":      "ek",
		"POSTHOG_PROJECT_API_KEY":     "ph-key",
	})

	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error when AI_MODEL_PROFILE_PATH is missing")
	}
	if !strings.Contains(err.Error(), "AI_MODEL_PROFILE_PATH") {
		t.Fatalf("error must mention AI_MODEL_PROFILE_PATH: %v", err)
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
  providerRegistryPath: ""
  defaultProviderBaseURL: ""
  defaultProviderApiKey: ""
  modelProfilePath: ""
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
  providerRegistryPath: "config/ai-providers.yaml"
  defaultProviderBaseURL: "https://provider.example"
  modelProfilePath: "config/ai-profiles.yaml"
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
			"ai.defaultProviderApiKey":         "AI_PROVIDER_API_KEY",
			"email.providerApiKey":             "EMAIL_PROVIDER_API_KEY",
			"featureFlag.posthogProjectApiKey": "POSTHOG_PROJECT_API_KEY",
		},
		SecretSource: mapSecret{
			"SESSION_COOKIE_SECRET":       "x",
			"AUTH_CHALLENGE_TOKEN_PEPPER": "x",
			"AI_PROVIDER_API_KEY":         "x",
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
  providerRegistryPath: "config/ai-providers.yaml"
  defaultProviderBaseURL: "https://provider.example"
  modelProfilePath: "config/ai-profiles.yaml"
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

func TestDefaultUploadConfigPaths(t *testing.T) {
	loader, err := config.Load(config.Options{
		AppEnv:    "dev",
		ConfigDir: filepath.Join("..", "..", "..", "..", "config"),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := loader.GetString("objectStorage.provider"); got != "minio" {
		t.Fatalf("objectStorage.provider = %q", got)
	}
	if got := loader.GetInt("upload.presignTTLSeconds"); got != 600 {
		t.Fatalf("upload.presignTTLSeconds = %d", got)
	}
	for path, want := range map[string]int{
		"resume.maxActive":              10,
		"upload.maxBytes.resume":        2097152,
		"upload.maxBytes.privacyExport": 5242880,
	} {
		if got := loader.GetInt(path); got != want {
			t.Fatalf("%s = %d, want %d", path, got, want)
		}
	}
	if got := loader.GetInt("upload.maxBytes.targetJobAttachment"); got != 0 {
		t.Fatalf("removed upload.maxBytes.targetJobAttachment = %d, want 0", got)
	}
}

func TestValidateUploadConfigFailsFast(t *testing.T) {
	for name, yaml := range map[string]string{
		"provider": `
objectStorage:
  provider: s3
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: 2097152
    privacyExport: 5242880
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
featureFlag:
  source: file
  filePath: "./feature-flags.yaml"
`,
		"ttl": `
objectStorage:
  provider: minio
upload:
  presignTTLSeconds: 0
  maxBytes:
    resume: 2097152
    privacyExport: 5242880
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
featureFlag:
  source: file
  filePath: "./feature-flags.yaml"
`,
		"max-bytes": `
objectStorage:
  provider: filesystem
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: -1
    privacyExport: 5242880
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
featureFlag:
  source: file
  filePath: "./feature-flags.yaml"
`,
		"resume-max-active": `
objectStorage:
  provider: filesystem
upload:
  presignTTLSeconds: 600
  maxBytes:
    resume: 2097152
    privacyExport: 5242880
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
  leaseTimeoutSeconds: 300
  shutdownGraceSeconds: 10
  reaperIntervalSeconds: 60
  scanIntervalSeconds: 5
resume:
  maxActive: 0
featureFlag:
  source: file
  filePath: "./feature-flags.yaml"
`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeYAML(t, filepath.Join(dir, "config.yaml"), yaml)
			loader, err := config.Load(config.Options{AppEnv: "dev", ConfigDir: dir})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			err = loader.Validate()
			if err == nil {
				t.Fatal("expected upload config validation error")
			}
			if !strings.Contains(err.Error(), "upload") &&
				!strings.Contains(err.Error(), "objectStorage.provider") &&
				!strings.Contains(err.Error(), "resume.maxActive") {
				t.Fatalf("error must mention upload/resume config boundary: %v", err)
			}
		})
	}
}

func TestUploadConfigDoesNotAddEnvDictionaryKeys(t *testing.T) {
	envBindings := config.DefaultEnvBindings()
	for key := range envBindings {
		if strings.HasPrefix(key, "UPLOAD_") || strings.HasPrefix(key, "OBJECT_STORE_") {
			t.Fatalf("backend-upload must not add unregistered env key %s", key)
		}
	}
}
