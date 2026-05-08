package targetjob

import (
	"fmt"
	"strings"
	"time"
)

// URL fetch boundary constants. These are code-owned, not app-level config:
// per spec D-7 they form the auditable SSRF / latency / size envelope and
// must be readable by `grep` in this package alone. See plan 3.3 for the
// SSRF test matrix that consumes these values.
const (
	// URLFetchTimeout caps each JD source HTTP request. Spec D-7 fixes 10s.
	URLFetchTimeout = 10 * time.Second

	// URLFetchBodyCap is the maximum body (in bytes) read from a JD source
	// before the response is rejected as oversized. Spec D-7 fixes 1 MiB.
	URLFetchBodyCap = 1 << 20

	// URLFetchUserAgentTemplate is the explicit crawler identifier. The
	// %s placeholder receives the running build version; an empty version
	// defaults to "dev". Spec D-7 forbids spoofing other UAs.
	URLFetchUserAgentTemplate = "EasyInterview JD-Crawler/%s (+https://easyinterview.local/crawler)"
)

// URLFetchUserAgent returns the canonical UA string to attach to outbound
// JD fetch HTTP requests. version "" maps to "dev" so unit tests and
// uninitialized boot paths still produce a valid identifier.
func URLFetchUserAgent(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}
	return fmt.Sprintf(URLFetchUserAgentTemplate, v)
}

// MustNotIntroduceAppLevelConfigKey is a runtime tripwire: any code that
// tries to register a new app-level config key from inside the targetjob
// domain panics with a clear "revise A4 first" message. New env-driven
// configuration must originate from docs/spec/secrets-and-config (A4) so
// the central inventory stays accurate.
func MustNotIntroduceAppLevelConfigKey(key string) {
	panic(fmt.Sprintf(
		"targetjob: refusing to register app-level config key %q from this domain; "+
			"revise docs/spec/secrets-and-config (A4) before adding it",
		key,
	))
}

// IsTestAppEnv reports whether the supplied APP_ENV value is the only
// environment in which stub AI providers are permitted (spec C-10 / plan
// 1.2). All other values must select real providers and fail-closed when
// secrets are missing.
func IsTestAppEnv(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "test")
}
