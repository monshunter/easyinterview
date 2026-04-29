package config_test

import (
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

// TestPackageSkeleton ensures the platform/config package skeleton (1.1) is
// importable and exports the entry-point types declared in spec §5.
func TestPackageSkeleton(t *testing.T) {
	t.Helper()
	var _ *config.Loader
	var _ config.RedactedString
}
