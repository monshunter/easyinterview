package targetjob

import (
	"strings"
)

// IsTestAppEnv reports whether the supplied APP_ENV value is the only
// environment in which stub AI providers are permitted (spec C-10 / plan
// 1.2). All other values must select real providers and fail-closed when
// secrets are missing.
func IsTestAppEnv(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "test")
}
