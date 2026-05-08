// Package bootstrap wires the production AIClient runtime from A3/A4
// registry, profile, and secret truth sources.
package bootstrap

import (
	"fmt"
	"net/http"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	doubaospeech "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/doubao_speech"
	minimaxspeech "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/minimax_speech"
	openaicompatible "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/stub"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// Options configures NewClient.
type Options struct {
	Config aiclient.Config

	SecretSource providerregistry.SecretSource
	HTTPClient   *http.Client
	OnWarn       func(error)

	ProviderRegistryPollInterval time.Duration
	ModelProfilePollInterval     time.Duration

	// AllowStubProvider is reserved for explicit mock/offline runtime wiring.
	// Non-test deployments must leave it false.
	AllowStubProvider bool
}

// Runtime owns the AIClient and the hot-reloading truth-source loaders it uses.
type Runtime struct {
	Client *aiclient.Client

	registry *providerregistry.Loader
	profiles *profile.Loader
}

// Close stops background reload loops.
func (r *Runtime) Close() {
	if r == nil {
		return
	}
	if r.registry != nil {
		r.registry.Close()
	}
	if r.profiles != nil {
		r.profiles.Close()
	}
}

// NewClient builds an AIClient from AI_PROVIDER_REGISTRY_PATH and
// AI_MODEL_PROFILE_PATH. It validates active profiles at startup so non-test
// deployments fail before serving if selected provider secrets are missing.
func NewClient(opts Options) (*Runtime, error) {
	registry, err := providerregistry.NewLoader(providerregistry.Options{
		Path:         opts.Config.ProviderRegistryPath,
		PollInterval: opts.ProviderRegistryPollInterval,
		OnWarn:       opts.OnWarn,
	})
	if err != nil {
		return nil, err
	}

	profiles, err := profile.NewLoader(profile.Options{
		Path:         opts.Config.ModelProfilePath,
		PollInterval: opts.ModelProfilePollInterval,
		OnWarn:       opts.OnWarn,
	})
	if err != nil {
		registry.Close()
		return nil, err
	}

	resolver := &providerResolver{
		registry:          registry,
		appEnv:            opts.Config.AppEnv,
		secrets:           opts.SecretSource,
		httpClient:        opts.HTTPClient,
		allowStubProvider: opts.AllowStubProvider,
	}

	if err := validateActiveProfiles(profiles, registry, opts); err != nil {
		profiles.Close()
		registry.Close()
		return nil, err
	}

	client, err := aiclient.New(
		opts.Config,
		aiclient.WithProfileResolver(profiles),
		aiclient.WithProviderResolver(resolver),
		aiclient.WithStubAllowed(opts.AllowStubProvider),
	)
	if err != nil {
		profiles.Close()
		registry.Close()
		return nil, err
	}

	return &Runtime{Client: client, registry: registry, profiles: profiles}, nil
}

func validateActiveProfiles(profiles *profile.Loader, registry *providerregistry.Loader, opts Options) error {
	for _, name := range profiles.Names() {
		p, err := profiles.Resolve(name)
		if err != nil {
			return err
		}
		if p.Status != aiclient.ProfileStatusActive {
			continue
		}
		resolved, err := registry.ResolveSelectedProviders(p, opts.Config.AppEnv, opts.SecretSource)
		if err != nil {
			return err
		}
		if opts.Config.AppEnv != aiclient.AppEnvTest && !opts.AllowStubProvider {
			for _, provider := range resolved {
				if provider.Entry.Protocol == aiclient.ProviderProtocolStub {
					return fmt.Errorf("%w: active profile %q selects stub provider %q outside APP_ENV=test", providerregistry.ErrProviderConfigInvalid, p.Name, provider.Entry.Name)
				}
			}
		}
	}
	return nil
}

type providerResolver struct {
	registry          *providerregistry.Loader
	appEnv            string
	secrets           providerregistry.SecretSource
	httpClient        *http.Client
	allowStubProvider bool
}

func (r *providerResolver) ResolveProvider(ref string) (aiclient.Provider, error) {
	entry, ok := r.registry.Provider(ref)
	if !ok {
		return nil, fmt.Errorf("%w: provider %q not found", providerregistry.ErrProviderConfigInvalid, ref)
	}

	switch entry.Protocol {
	case aiclient.ProviderProtocolStub:
		return stub.New(stub.WithAppEnv(r.appEnv), stub.WithAllowed(r.allowStubProvider))
	case aiclient.ProviderProtocolOpenAICompatible:
		resolved, err := providerregistry.ResolveProviderEntry(entry, r.appEnv, r.secrets)
		if err != nil {
			return nil, err
		}
		return openaicompatible.New(openaicompatible.Options{
			Provider:   resolved,
			HTTPClient: r.httpClient,
		})
	case aiclient.ProviderProtocolDoubaoSpeech:
		resolved, err := providerregistry.ResolveProviderEntry(entry, r.appEnv, r.secrets)
		if err != nil {
			return nil, err
		}
		return doubaospeech.New(doubaospeech.Options{
			Provider:   resolved,
			HTTPClient: r.httpClient,
		})
	case aiclient.ProviderProtocolMinimaxSpeech:
		resolved, err := providerregistry.ResolveProviderEntry(entry, r.appEnv, r.secrets)
		if err != nil {
			return nil, err
		}
		return minimaxspeech.New(minimaxspeech.Options{
			Provider:   resolved,
			HTTPClient: r.httpClient,
		})
	case aiclient.ProviderProtocolRealtimeAudio,
		aiclient.ProviderProtocolJudgeCompatible:
		return nil, sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, fmt.Sprintf("provider protocol %q is not implemented", entry.Protocol), false)
	default:
		return nil, fmt.Errorf("%w: unsupported provider protocol %q", providerregistry.ErrProviderConfigInvalid, entry.Protocol)
	}
}
