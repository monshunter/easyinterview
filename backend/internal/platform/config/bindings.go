package config

// CanonicalOptions configures LoadCanonical. It wires the P0 env dictionary
// from secrets-and-config spec §3.1.1 into the generic Loader.
type CanonicalOptions struct {
	AppEnv       string
	ConfigDir    string
	SecretSource SecretSource
}

// LoadCanonical loads the repository's P0 config schema with the canonical
// env and secret bindings. Backend runtime entrypoints use this helper so
// startup validation sees the same code-side env dictionary.
func LoadCanonical(opts CanonicalOptions) (*Loader, error) {
	return Load(Options{
		AppEnv:         opts.AppEnv,
		ConfigDir:      opts.ConfigDir,
		EnvBindings:    DefaultEnvBindings(),
		SecretBindings: DefaultSecretBindings(),
		SecretSource:   opts.SecretSource,
	})
}

// DefaultEnvBindings maps env keys from spec §3.1.1 to canonical dot paths.
// It intentionally includes secret keys as env bindings too: env vars are the
// P0 SecretSource backend, while future Vault/SOPS sources can override these
// same dot paths through DefaultSecretBindings.
func DefaultEnvBindings() map[string]string {
	return cloneStringMap(defaultEnvBindings)
}

// DefaultSecretBindings maps secret config paths to their env-backed secret
// names. Runtime secret sources have higher priority than plain env bindings.
func DefaultSecretBindings() map[string]string {
	return cloneStringMap(defaultSecretBindings)
}

var defaultEnvBindings = map[string]string{
	"APP_ENV":                     "app.env",
	"APP_LISTEN_ADDR":             "app.listenAddr",
	"DATABASE_URL":                "database.url",
	"REDIS_URL":                   "redis.url",
	"OBJECT_STORAGE_ENDPOINT":     "objectStorage.endpoint",
	"OBJECT_STORAGE_BUCKET":       "objectStorage.bucket",
	"OBJECT_STORAGE_ACCESS_KEY":   "objectStorage.accessKey",
	"OBJECT_STORAGE_SECRET_KEY":   "objectStorage.secretKey",
	"OTEL_EXPORTER_OTLP_ENDPOINT": "observability.otlpEndpoint",
	"LOG_LEVEL":                   "log.level",
	"SESSION_COOKIE_SECRET":       "auth.sessionCookieSecret",
	"AUTH_CHALLENGE_TOKEN_PEPPER": "auth.challengeTokenPepper",
	"AI_PROVIDER_REGISTRY_PATH":   "ai.providerRegistryPath",
	"AI_PROVIDER_BASE_URL":        "ai.defaultProviderBaseURL",
	"AI_PROVIDER_API_KEY":         "ai.defaultProviderApiKey",
	"AI_MODEL_PROFILE_PATH":       "ai.modelProfilePath",
	"AI_DEBUG_CAPTURE_RAW_IO":     "ai.debugCaptureRawIO",
	"AI_DEBUG_RAW_IO_PATH":        "ai.debugRawIOPath",
	"FEATURE_FLAG_SOURCE":         "featureFlag.source",
	"FEATURE_FLAG_FILE_PATH":      "featureFlag.filePath",
	"POSTHOG_HOST":                "featureFlag.posthogHost",
	"POSTHOG_SELF_HOSTED":         "featureFlag.posthogSelfHosted",
	"POSTHOG_PROJECT_API_KEY":     "featureFlag.posthogProjectApiKey",
	"POSTHOG_PUBLIC_KEY":          "featureFlag.posthogPublicKey",
	"EMAIL_PROVIDER":              "email.provider",
	"EMAIL_SMTP_HOST":             "email.smtpHost",
	"EMAIL_SMTP_PORT":             "email.smtpPort",
	"EMAIL_SMTP_USERNAME":         "email.smtpUsername",
	"EMAIL_SMTP_PASSWORD":         "email.smtpPassword",
	"EMAIL_SMTP_TLS_MODE":         "email.smtpTLSMode",
	"EMAIL_FROM_ADDRESS":          "email.fromAddress",
	"EMAIL_VERIFY_BASE_URL":       "email.verifyBaseURL",
}

var defaultSecretBindings = map[string]string{
	"database.url":                     "DATABASE_URL",
	"redis.url":                        "REDIS_URL",
	"objectStorage.accessKey":          "OBJECT_STORAGE_ACCESS_KEY",
	"objectStorage.secretKey":          "OBJECT_STORAGE_SECRET_KEY",
	"auth.sessionCookieSecret":         "SESSION_COOKIE_SECRET",
	"auth.challengeTokenPepper":        "AUTH_CHALLENGE_TOKEN_PEPPER",
	"ai.defaultProviderApiKey":         "AI_PROVIDER_API_KEY",
	"featureFlag.posthogProjectApiKey": "POSTHOG_PROJECT_API_KEY",
	"email.smtpPassword":               "EMAIL_SMTP_PASSWORD",
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
