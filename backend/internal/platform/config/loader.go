package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ErrSecretMissing is returned by SecretSource implementations when the
// requested secret name is not provisioned. The loader propagates it so
// callers can distinguish "no secret available" from transport errors.
var ErrSecretMissing = errors.New("secret missing")

// SecretSource is implemented by runtime secret providers (D-3). It is
// re-exported from the secrets package indirectly via Options.SecretSource;
// the loader depends only on the minimal Get(name) (string, error) shape.
type SecretSource interface {
	Get(name string) (string, error)
}

// Loader exposes the merged configuration through typed Get* accessors. All
// business code reads configuration through Loader to enforce the
// os.Getenv boundary lint (spec §4.1).
type Loader struct {
	k        *koanf.Koanf
	secrets  map[string]string
	required map[string]bool
	appEnv   string
}

// AppEnv returns the active environment label, e.g. "dev"/"staging"/"prod".
func (l *Loader) AppEnv() string {
	return l.appEnv
}

// Options configures Load. The four layers are merged in spec D-1 priority:
// config.yaml -> {AppEnv}.yaml -> os env (via EnvBindings) -> SecretSource
// (via SecretBindings). Layers are applied serially; concurrent provider
// loading is forbidden because koanf merges last-write-wins.
type Options struct {
	// AppEnv selects {AppEnv}.yaml under ConfigDir. Empty value loads only
	// the base config.yaml file.
	AppEnv string
	// ConfigDir holds config.yaml and {AppEnv}.yaml. Required.
	ConfigDir string
	// EnvBindings maps environment variable name -> dot-path key.
	// Example: "APP_LISTEN_ADDR" -> "app.listenAddr".
	EnvBindings map[string]string
	// SecretBindings maps dot-path key -> secret name passed to SecretSource.
	// Example: "auth.sessionCookieSecret" -> "SESSION_COOKIE_SECRET".
	SecretBindings map[string]string
	// SecretSource resolves runtime secrets. When nil, secret bindings are
	// skipped (useful for unit tests with public-only fixtures).
	SecretSource SecretSource
	// RequiredKeys lists dot-path keys that must be non-empty after merge.
	// Validation is enforced in Loader.Validate (item 1.5).
	RequiredKeys []string
}

// Load merges the four configuration layers in spec D-1 priority and
// returns a typed Loader. Returns an error if a required file is missing
// or YAML parsing fails. Secret resolution errors are also surfaced.
func Load(opts Options) (*Loader, error) {
	if opts.ConfigDir == "" {
		return nil, fmt.Errorf("config: ConfigDir is required")
	}
	k := koanf.New(".")
	parser := yaml.Parser()

	base := filepath.Join(opts.ConfigDir, "config.yaml")
	if err := loadYAMLIfExists(k, base, parser); err != nil {
		return nil, fmt.Errorf("config: load base: %w", err)
	}

	if opts.AppEnv != "" {
		envFile := filepath.Join(opts.ConfigDir, opts.AppEnv+".yaml")
		if err := loadYAMLIfExists(k, envFile, parser); err != nil {
			return nil, fmt.Errorf("config: load %s override: %w", opts.AppEnv, err)
		}
	}

	for envKey, dotPath := range opts.EnvBindings {
		if v, ok := os.LookupEnv(envKey); ok && v != "" {
			if err := k.Set(dotPath, v); err != nil {
				return nil, fmt.Errorf("config: bind env %s -> %s: %w", envKey, dotPath, err)
			}
		}
	}

	secretValues := make(map[string]string, len(opts.SecretBindings))
	if len(opts.SecretBindings) > 0 && opts.SecretSource != nil {
		for dotPath, name := range opts.SecretBindings {
			value, err := opts.SecretSource.Get(name)
			if err != nil {
				if errors.Is(err, ErrSecretMissing) {
					continue
				}
				return nil, fmt.Errorf("config: read secret %s: %w", name, err)
			}
			if value != "" {
				if err := k.Set(dotPath, value); err != nil {
					return nil, fmt.Errorf("config: bind secret %s -> %s: %w", name, dotPath, err)
				}
				secretValues[dotPath] = value
			}
		}
	}

	required := make(map[string]bool, len(opts.RequiredKeys))
	for _, key := range opts.RequiredKeys {
		required[strings.TrimSpace(key)] = true
	}

	return &Loader{
		k:        k,
		secrets:  secretValues,
		required: required,
		appEnv:   opts.AppEnv,
	}, nil
}

func loadYAMLIfExists(k *koanf.Koanf, path string, parser koanf.Parser) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return k.Load(file.Provider(path), parser)
}
