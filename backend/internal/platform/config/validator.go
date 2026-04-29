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
		problems = append(problems, l.checkRequiredSecret("auth.sessionCookieSecret", "SESSION_COOKIE_SECRET")...)
		problems = append(problems, l.checkRequiredSecret("auth.challengeTokenPepper", "AUTH_CHALLENGE_TOKEN_PEPPER")...)
		problems = append(problems, l.checkRequiredSecret("ai.gatewayApiKey", "AI_GATEWAY_API_KEY")...)
		problems = append(problems, l.checkRequiredSecret("email.providerApiKey", "EMAIL_PROVIDER_API_KEY")...)
		if strings.EqualFold(l.GetString("featureFlag.source"), "posthog") {
			problems = append(problems, l.checkRequiredSecret("featureFlag.posthogProjectApiKey", "POSTHOG_PROJECT_API_KEY")...)
			if !l.GetBool("featureFlag.posthogSelfHosted") {
				problems = append(problems, "POSTHOG_SELF_HOSTED must be true in staging/prod (spec §4.1)")
			}
		}
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
