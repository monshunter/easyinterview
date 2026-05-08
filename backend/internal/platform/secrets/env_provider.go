package secrets

import (
	"fmt"
	"os"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

// EnvSecretSource resolves runtime secrets from process environment variables.
type EnvSecretSource struct{}

// Get returns the value for name or config.ErrSecretMissing when the variable
// is unset or empty.
func (EnvSecretSource) Get(name string) (string, error) {
	key := strings.TrimSpace(name)
	if key == "" {
		return "", fmt.Errorf("%w: empty secret name", config.ErrSecretMissing)
	}
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return "", fmt.Errorf("%w: %s", config.ErrSecretMissing, key)
	}
	return value, nil
}
