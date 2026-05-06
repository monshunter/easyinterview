package config

import (
	"fmt"
	"sort"
	"strings"
)

// Validate enforces fail-fast required-field checks at process startup. The
// rules implement spec §3.1.1 / §3.1.2 / §6 (C-2, C-4, C-10, C-12):
//
//   - APP_ENV=staging|prod must have all auth, AI, email and (when
//     featureFlag.source=posthog) PostHog secrets populated by the
//     SecretSource. Missing keys are reported with their env-key names so
//     deployers know what to provision.
//   - APP_ENV=test allows missing AI / Email / Session secrets so unit
//     tests can run without provisioning real credentials.
//   - When featureFlag.source=posthog, staging/prod must set
//     featureFlag.posthogSelfHosted=true (D-Q3 / spec §4.1).
//   - async.queueWeights must declare three positive entries (D-9 / C-12).
//
// Validation only inspects the merged Loader state; it does not mutate it.
// The first call materializes the failure list deterministically (sorted)
// so error messages are stable across runs.
func (l *Loader) Validate() error {
	if l == nil {
		return fmt.Errorf("config: loader is nil")
	}
	env := strings.ToLower(strings.TrimSpace(l.appEnv))
	var problems []string

	if env == "staging" || env == "prod" {
		problems = append(problems, l.checkRequiredValue("app.listenAddr", "APP_LISTEN_ADDR")...)
		problems = append(problems, l.checkRequiredRuntimeValue("database.url", "DATABASE_URL")...)
		problems = append(problems, l.checkRequiredRuntimeValue("redis.url", "REDIS_URL")...)
		problems = append(problems, l.checkRequiredRuntimeValue("objectStorage.endpoint", "OBJECT_STORAGE_ENDPOINT")...)
		problems = append(problems, l.checkRequiredRuntimeValue("objectStorage.bucket", "OBJECT_STORAGE_BUCKET")...)
		problems = append(problems, l.checkRequiredRuntimeValue("objectStorage.accessKey", "OBJECT_STORAGE_ACCESS_KEY")...)
		problems = append(problems, l.checkRequiredRuntimeValue("objectStorage.secretKey", "OBJECT_STORAGE_SECRET_KEY")...)
		problems = append(problems, l.checkRequiredValue("log.level", "LOG_LEVEL")...)
		problems = append(problems, l.checkRequiredSecret("auth.sessionCookieSecret", "SESSION_COOKIE_SECRET")...)
		problems = append(problems, l.checkRequiredSecret("auth.challengeTokenPepper", "AUTH_CHALLENGE_TOKEN_PEPPER")...)
		problems = append(problems, l.checkRequiredValue("ai.providerRegistryPath", "AI_PROVIDER_REGISTRY_PATH")...)
		problems = append(problems, l.checkRequiredValue("ai.modelProfilePath", "AI_MODEL_PROFILE_PATH")...)
		problems = append(problems, l.checkRequiredValue("featureFlag.source", "FEATURE_FLAG_SOURCE")...)
		switch strings.ToLower(strings.TrimSpace(l.GetString("featureFlag.source"))) {
		case "file":
			problems = append(problems, l.checkRequiredValue("featureFlag.filePath", "FEATURE_FLAG_FILE_PATH")...)
		case "posthog":
			problems = append(problems, l.checkRequiredValue("featureFlag.posthogHost", "POSTHOG_HOST")...)
			problems = append(problems, l.checkRequiredSecret("featureFlag.posthogProjectApiKey", "POSTHOG_PROJECT_API_KEY")...)
			if !l.GetBool("featureFlag.posthogSelfHosted") {
				problems = append(problems, "POSTHOG_SELF_HOSTED must be true in staging/prod (spec §4.1)")
			}
		default:
			problems = append(problems, "FEATURE_FLAG_SOURCE must be file or posthog in staging/prod")
		}
		problems = append(problems, l.checkRequiredValue("email.provider", "EMAIL_PROVIDER")...)
		problems = append(problems, l.checkRequiredSecret("email.providerApiKey", "EMAIL_PROVIDER_API_KEY")...)
	}

	if env != "test" {
		// Async queue weights must declare three positive entries.
		if l.GetInt("async.queueWeights.critical") <= 0 ||
			l.GetInt("async.queueWeights.default") <= 0 ||
			l.GetInt("async.queueWeights.low") <= 0 {
			problems = append(problems, "async.queueWeights must declare positive critical/default/low values (spec C-12)")
		}
	}

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("config validation failed:\n  - %s", strings.Join(problems, "\n  - "))
}

func (l *Loader) checkRequiredSecret(dotPath, envKey string) []string {
	if v, ok := l.secrets[dotPath]; ok && v != "" {
		return nil
	}
	if l.GetString(dotPath) != "" {
		return nil
	}
	return []string{fmt.Sprintf("missing required secret: %s (config path %s)", envKey, dotPath)}
}

func (l *Loader) checkRequiredValue(dotPath, envKey string) []string {
	if l.GetString(dotPath) != "" {
		return nil
	}
	return []string{fmt.Sprintf("missing required config: %s (config path %s)", envKey, dotPath)}
}

func (l *Loader) checkRequiredRuntimeValue(dotPath, envKey string) []string {
	if l.GetString(dotPath) == "" {
		return []string{fmt.Sprintf("missing required config: %s (config path %s)", envKey, dotPath)}
	}
	if l.runtimeBound[dotPath] {
		return nil
	}
	return []string{fmt.Sprintf("missing required runtime override: %s (config path %s)", envKey, dotPath)}
}
