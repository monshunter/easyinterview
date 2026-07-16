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
    resume: 2048
    privacyExport: 4096
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
	t.Setenv("EMAIL_PROVIDER", "smtp")
	t.Setenv("EMAIL_SMTP_HOST", "smtp.example.test")
	t.Setenv("EMAIL_SMTP_PORT", "587")
	t.Setenv("EMAIL_SMTP_USERNAME", "mailer")
	t.Setenv("EMAIL_SMTP_TLS_MODE", "starttls")
	t.Setenv("EMAIL_FROM_ADDRESS", "noreply@example.test")
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
	if got := envBindings["AI_DEBUG_CAPTURE_RAW_IO"]; got != "ai.debugCaptureRawIO" {
		t.Fatalf("AI_DEBUG_CAPTURE_RAW_IO binding = %q", got)
	}
	if got := envBindings["AI_DEBUG_RAW_IO_PATH"]; got != "ai.debugRawIOPath" {
		t.Fatalf("AI_DEBUG_RAW_IO_PATH binding = %q", got)
	}
	if _, exists := envBindings["AI_DEBUG_PRINT_RAW_OUTPUT"]; exists {
		t.Fatalf("legacy AI_DEBUG_PRINT_RAW_OUTPUT binding remains: %+v", envBindings)
	}

	secretBindings := config.DefaultSecretBindings()
	if _, ok := secretBindings["ai.providerApiKey"]; ok {
		t.Fatalf("out-of-scope ai.providerApiKey secret binding must not remain: %+v", secretBindings)
	}
	if got := secretBindings["ai.defaultProviderApiKey"]; got != "AI_PROVIDER_API_KEY" {
		t.Fatalf("ai.defaultProviderApiKey secret binding = %q", got)
	}
}

func TestAIDebugRawCaptureEnvBindingsOverrideDefaults(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
ai:
  debugCaptureRawIO: false
  debugRawIOPath: .test-output/local-dev/ai-raw.ndjson
`)
	wantPath := filepath.Join(t.TempDir(), "override.ndjson")
	t.Setenv("AI_DEBUG_CAPTURE_RAW_IO", "true")
	t.Setenv("AI_DEBUG_RAW_IO_PATH", wantPath)

	loader, err := config.Load(config.Options{
		AppEnv:      "test",
		ConfigDir:   dir,
		EnvBindings: config.DefaultEnvBindings(),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loader.GetBool("ai.debugCaptureRawIO"); !got {
		t.Fatalf("AI_DEBUG_CAPTURE_RAW_IO env binding = %v", got)
	}
	if got := loader.GetString("ai.debugRawIOPath"); got != wantPath {
		t.Fatalf("AI_DEBUG_RAW_IO_PATH env binding = %q, want %q", got, wantPath)
	}
}

func TestRepoLocalConfigEnablesRawCaptureOnlyForLocalEnvironments(t *testing.T) {
	configDir := filepath.Clean("../../../../config")

	for _, appEnv := range []string{"dev", "test"} {
		loader, err := config.Load(config.Options{
			AppEnv:    appEnv,
			ConfigDir: configDir,
		})
		if err != nil {
			t.Fatalf("Load(%s): %v", appEnv, err)
		}
		if got := loader.GetBool("ai.debugCaptureRawIO"); !got {
			t.Fatalf("%s ai.debugCaptureRawIO = %v, want true", appEnv, got)
		}
		if got := loader.GetString("ai.debugRawIOPath"); !filepath.IsAbs(got) {
			t.Fatalf("%s ai.debugRawIOPath = %q, want absolute path", appEnv, got)
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
		if got := loader.GetBool("ai.debugCaptureRawIO"); got {
			t.Fatalf("%s ai.debugCaptureRawIO = %v, want false", appEnv, got)
		}
	}
}

func TestDefaultEmailDictionaryIncludesSMTPProviderBindings(t *testing.T) {
	envBindings := config.DefaultEnvBindings()
	for key, want := range map[string]string{
		"EMAIL_PROVIDER":        "email.provider",
		"EMAIL_SMTP_HOST":       "email.smtpHost",
		"EMAIL_SMTP_PORT":       "email.smtpPort",
		"EMAIL_SMTP_USERNAME":   "email.smtpUsername",
		"EMAIL_SMTP_PASSWORD":   "email.smtpPassword",
		"EMAIL_SMTP_TLS_MODE":   "email.smtpTLSMode",
		"EMAIL_FROM_ADDRESS":    "email.fromAddress",
		"EMAIL_VERIFY_BASE_URL": "email.verifyBaseURL",
	} {
		if got := envBindings[key]; got != want {
			t.Fatalf("%s binding = %q, want %q", key, got, want)
		}
	}
	if _, ok := envBindings["EMAIL_PROVIDER_API_KEY"]; ok {
		t.Fatal("EMAIL_PROVIDER_API_KEY must not remain in the current email dictionary")
	}
	if got := config.DefaultSecretBindings()["email.smtpPassword"]; got != "EMAIL_SMTP_PASSWORD" {
		t.Fatalf("email.smtpPassword secret binding = %q", got)
	}
}

func TestValidateEmailProviderContractInDev(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		host     string
		port     string
		username string
		password string
		tlsMode  string
		wantKey  string
	}{
		{name: "mailpit plain without auth", provider: "mailpit", host: "127.0.0.1", port: "1025", tlsMode: "none"},
		{name: "smtp starttls with auth", provider: "smtp", host: "smtp.example.test", port: "587", username: "mailer", password: "secret", tlsMode: "starttls"},
		{name: "smtp implicit tls with auth", provider: "smtp", host: "smtp.example.test", port: "465", username: "mailer", password: "secret", tlsMode: "tls"},
		{name: "unknown provider", provider: "ses", host: "smtp.example.test", port: "587", username: "mailer", password: "secret", tlsMode: "starttls", wantKey: "EMAIL_PROVIDER"},
		{name: "smtp rejects none tls", provider: "smtp", host: "smtp.example.test", port: "587", username: "mailer", password: "secret", tlsMode: "none", wantKey: "EMAIL_SMTP_TLS_MODE"},
		{name: "smtp requires username", provider: "smtp", host: "smtp.example.test", port: "587", password: "secret", tlsMode: "starttls", wantKey: "EMAIL_SMTP_USERNAME"},
		{name: "smtp requires password", provider: "smtp", host: "smtp.example.test", port: "587", username: "mailer", tlsMode: "starttls", wantKey: "EMAIL_SMTP_PASSWORD"},
		{name: "rejects invalid port", provider: "mailpit", host: "127.0.0.1", port: "70000", tlsMode: "none", wantKey: "EMAIL_SMTP_PORT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EMAIL_PROVIDER", tt.provider)
			t.Setenv("EMAIL_SMTP_HOST", tt.host)
			t.Setenv("EMAIL_SMTP_PORT", tt.port)
			t.Setenv("EMAIL_SMTP_USERNAME", tt.username)
			t.Setenv("EMAIL_SMTP_TLS_MODE", tt.tlsMode)
			t.Setenv("EMAIL_FROM_ADDRESS", "noreply@example.test")
			loader, err := config.LoadCanonical(config.CanonicalOptions{
				AppEnv:    "dev",
				ConfigDir: filepath.Clean("../../../../config"),
				SecretSource: mapSecret{
					"EMAIL_SMTP_PASSWORD": tt.password,
				},
			})
			if err != nil {
				t.Fatalf("LoadCanonical: %v", err)
			}
			err = loader.Validate()
			if tt.wantKey == "" {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantKey) {
				t.Fatalf("Validate error = %v, want key %s", err, tt.wantKey)
			}
			if tt.password != "" && strings.Contains(err.Error(), tt.password) {
				t.Fatal("validation error leaked SMTP password")
			}
		})
	}
}

func TestValidateProdMissingSecretFailsFast(t *testing.T) {
	setCompleteProdRuntimeEnv(t)
	loader := newProdLoader(t, mapSecret{})
	err := loader.Validate()
	if err == nil {
		t.Fatal("expected validate error in prod with missing secrets")
	}
	msg := err.Error()
	for _, key := range []string{"SESSION_COOKIE_SECRET", "AUTH_CHALLENGE_TOKEN_PEPPER", "EMAIL_SMTP_PASSWORD"} {
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
		"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
		"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
		"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
		"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
		"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
		"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
email:
  provider: smtp
  smtpHost: smtp.example.test
  smtpPort: 587
  smtpUsername: mailer
  smtpTLSMode: starttls
  fromAddress: noreply@example.test
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
			"email.smtpPassword":               "EMAIL_SMTP_PASSWORD",
			"featureFlag.posthogProjectApiKey": "POSTHOG_PROJECT_API_KEY",
		},
		SecretSource: mapSecret{
			"SESSION_COOKIE_SECRET":       "x",
			"AUTH_CHALLENGE_TOKEN_PEPPER": "x",
			"AI_PROVIDER_API_KEY":         "x",
			"EMAIL_SMTP_PASSWORD":         "smtp-secret",
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
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
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
				!strings.Contains(err.Error(), "objectStorage.provider") {
				t.Fatalf("error must mention upload config boundary: %v", err)
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

func TestLocalAIRawCaptureConfigContract(t *testing.T) {
	t.Run("canonical env dictionary replaces stderr debug key", func(t *testing.T) {
		bindings := config.DefaultEnvBindings()
		if got := bindings["AI_DEBUG_CAPTURE_RAW_IO"]; got != "ai.debugCaptureRawIO" {
			t.Fatalf("AI_DEBUG_CAPTURE_RAW_IO binding = %q", got)
		}
		if got := bindings["AI_DEBUG_RAW_IO_PATH"]; got != "ai.debugRawIOPath" {
			t.Fatalf("AI_DEBUG_RAW_IO_PATH binding = %q", got)
		}
		if _, exists := bindings["AI_DEBUG_PRINT_RAW_OUTPUT"]; exists {
			t.Fatalf("legacy stderr raw-output binding remains: %+v", bindings)
		}
	})

	t.Run("repo dev and test default enabled with ConfigDir parent anchor", func(t *testing.T) {
		t.Setenv("AI_DEBUG_CAPTURE_RAW_IO", "")
		t.Setenv("AI_DEBUG_RAW_IO_PATH", "")
		configDir, err := filepath.Abs(filepath.Clean("../../../../config"))
		if err != nil {
			t.Fatalf("resolve repo config dir: %v", err)
		}
		wantPath := filepath.Join(filepath.Dir(configDir), ".test-output", "local-dev", "ai-raw.ndjson")
		for _, appEnv := range []string{"dev", "test"} {
			loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: appEnv, ConfigDir: configDir})
			if err != nil {
				t.Fatalf("LoadCanonical(%s): %v", appEnv, err)
			}
			if !loader.GetBool("ai.debugCaptureRawIO") {
				t.Errorf("%s ai.debugCaptureRawIO = false, want true", appEnv)
			}
			if got := loader.GetString("ai.debugRawIOPath"); got != wantPath {
				t.Errorf("%s effective raw path = %q, want ConfigDir-parent anchored %q", appEnv, got, wantPath)
			}
		}
	})

	t.Run("repo staging and prod default disabled", func(t *testing.T) {
		t.Setenv("AI_DEBUG_CAPTURE_RAW_IO", "")
		t.Setenv("AI_DEBUG_RAW_IO_PATH", "")
		configDir := filepath.Clean("../../../../config")
		for _, appEnv := range []string{"staging", "prod"} {
			loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: appEnv, ConfigDir: configDir})
			if err != nil {
				t.Fatalf("LoadCanonical(%s): %v", appEnv, err)
			}
			if loader.GetBool("ai.debugCaptureRawIO") {
				t.Errorf("%s ai.debugCaptureRawIO = true, want false", appEnv)
			}
		}
	})

	t.Run("test accepts an absolute override", func(t *testing.T) {
		dir := t.TempDir()
		writeYAML(t, filepath.Join(dir, "config.yaml"), "ai:\n  debugCaptureRawIO: false\n  debugRawIOPath: .test-output/local-dev/ai-raw.ndjson\n")
		wantPath := filepath.Join(t.TempDir(), "custom-ai-raw.ndjson")
		t.Setenv("AI_DEBUG_CAPTURE_RAW_IO", "true")
		t.Setenv("AI_DEBUG_RAW_IO_PATH", wantPath)
		loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: "test", ConfigDir: dir})
		if err != nil {
			t.Fatalf("LoadCanonical: %v", err)
		}
		if !loader.GetBool("ai.debugCaptureRawIO") || loader.GetString("ai.debugRawIOPath") != wantPath {
			t.Fatalf("legal override not applied: enabled=%v path=%q", loader.GetBool("ai.debugCaptureRawIO"), loader.GetString("ai.debugRawIOPath"))
		}
		if err := loader.Validate(); err != nil {
			t.Fatalf("valid test raw capture override rejected: %v", err)
		}
	})

	t.Run("enabled capture requires a path", func(t *testing.T) {
		dir := t.TempDir()
		writeYAML(t, filepath.Join(dir, "config.yaml"), "ai:\n  debugCaptureRawIO: true\n  debugRawIOPath: \"\"\n")
		loader, err := config.Load(config.Options{AppEnv: "test", ConfigDir: dir})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if err := loader.Validate(); err == nil || !strings.Contains(err.Error(), "AI_DEBUG_RAW_IO_PATH") {
			t.Fatalf("enabled empty raw path error = %v, want AI_DEBUG_RAW_IO_PATH", err)
		}
	})

	t.Run("staging and prod reject explicit enable", func(t *testing.T) {
		for _, appEnv := range []string{"staging", "prod"} {
			t.Run(appEnv, func(t *testing.T) {
				dir := t.TempDir()
				writeYAML(t, filepath.Join(dir, "config.yaml"), `
ai:
  debugCaptureRawIO: false
  debugRawIOPath: .test-output/local-dev/ai-raw.ndjson
`)
				t.Setenv("AI_DEBUG_CAPTURE_RAW_IO", "true")
				t.Setenv("AI_DEBUG_RAW_IO_PATH", filepath.Join(t.TempDir(), "forbidden.ndjson"))
				loader, err := config.LoadCanonical(config.CanonicalOptions{AppEnv: appEnv, ConfigDir: dir})
				if err != nil {
					t.Fatalf("LoadCanonical: %v", err)
				}
				if err := loader.Validate(); err == nil || !strings.Contains(err.Error(), "AI_DEBUG_CAPTURE_RAW_IO") {
					t.Fatalf("%s explicit capture error = %v, want AI_DEBUG_CAPTURE_RAW_IO", appEnv, err)
				}
			})
		}
	})
}
