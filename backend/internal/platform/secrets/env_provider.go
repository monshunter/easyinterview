package secrets

import (
	"os"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

// EnvSecretSource resolves secret names from environment variables.
type EnvSecretSource struct{}

func (EnvSecretSource) Get(name string) (string, error) {
	key := strings.TrimSpace(name)
	if key == "" {
		return "", config.ErrSecretMissing
	}
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return "", config.ErrSecretMissing
	}
	return value, nil
}
