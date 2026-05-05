package aiclient

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// Client implements AIClient by dispatching each call to the Provider named
// in the resolved ModelProfile. It is the only concrete AIClient
// implementation in plan 001.
type Client struct {
	cfg       Config
	resolver  ProfileResolver
	providers map[string]Provider
	builder   metaBuilder

	// taskRunWriter and auditWriter are populated by Phase 3's decorator
	// wiring; plan 001 leaves the call-time observability hookup to
	// observability.Wrap so that this Client struct stays small.
	taskRunWriter AITaskRunWriter
	auditWriter   AuditEventWriter
}

// New builds a Client. Spec D-4 / plan 4.1 fail-fast rules:
//
//   - cfg.AppEnv == "test" + WithStubAllowed(true): success regardless of
//     provider config (single-process unit / contract tests).
//   - cfg.AppEnv == "test" without WithStubAllowed and missing
//     ProviderRegistryPath/ModelProfilePath: ErrMissingProviderConfig.
//   - cfg.AppEnv != "test" with missing ProviderRegistryPath or
//     ModelProfilePath: ErrMissingProviderConfig regardless of
//     WithStubAllowed; non-test deployments must load registry-backed
//     providers and capability profiles.
func New(cfg Config, opts ...Option) (*Client, error) {
	o := &clientOptions{
		providers: map[string]Provider{},
	}
	for _, opt := range opts {
		opt(o)
	}

	if cfg.AppEnv == AppEnvTest {
		if !o.stubAllowed && (cfg.ProviderRegistryPath == "" || cfg.ModelProfilePath == "") {
			return nil, ErrMissingProviderConfig
		}
	} else {
		if cfg.ProviderRegistryPath == "" || cfg.ModelProfilePath == "" {
			return nil, ErrMissingProviderConfig
		}
	}

	c := &Client{
		cfg:           cfg,
		resolver:      o.resolver,
		providers:     o.providers,
		taskRunWriter: o.taskRunWriter,
		auditWriter:   o.auditWriter,
	}
	return c, nil
}

// Resolver exposes the configured ProfileResolver. The decorator and tests
// use it to introspect profiles without re-implementing resolution.
func (c *Client) Resolver() ProfileResolver { return c.resolver }

// Providers exposes the registered providers keyed by Provider.Name().
func (c *Client) Providers() map[string]Provider { return c.providers }

// AITaskRunWriter returns the configured ai_task_runs writer (may be nil).
func (c *Client) AITaskRunWriter() AITaskRunWriter { return c.taskRunWriter }

// AuditEventWriter returns the configured audit_events writer (may be nil).
func (c *Client) AuditEventWriter() AuditEventWriter { return c.auditWriter }

// Complete implements AIClient.
func (c *Client) Complete(ctx context.Context, profileName string, payload CompletePayload) (CompleteResponse, AICallMeta, error) {
	if len(payload.Messages) == 0 {
		return CompleteResponse{}, AICallMeta{
			ModelProfileName: profileName,
			ValidationStatus: ValidationStatusInvalid,
			ErrorCode:        sharederrors.CodeAiOutputInvalid,
		}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "messages must be non-empty", false)
	}

	profile, provider, err := c.dispatch(profileName, CapabilityChat)
	if err != nil {
		return CompleteResponse{}, failureMeta(profileName, profile, err), err
	}

	resp, partial, err := executeWithFallback(profile, provider, c.providers, func(p Provider, attempt *ModelProfile) (CompleteResponse, AICallMeta, error) {
		return p.Complete(ctx, attempt, payload)
	})
	meta, mergeErr := c.builder.merge(profile, payload.Metadata, partial)
	if mergeErr != nil && err == nil {
		err = mergeErr
	}
	return resp, meta, err
}

// Embed implements AIClient.
func (c *Client) Embed(ctx context.Context, profileName string, input EmbedInput) (EmbedResponse, AICallMeta, error) {
	if len(input.Texts) == 0 {
		return EmbedResponse{}, AICallMeta{
			ModelProfileName: profileName,
			ValidationStatus: ValidationStatusInvalid,
			ErrorCode:        sharederrors.CodeAiOutputInvalid,
		}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "texts must be non-empty", false)
	}

	profile, provider, err := c.dispatch(profileName, CapabilityEmbed)
	if err != nil {
		return EmbedResponse{}, failureMeta(profileName, profile, err), err
	}

	resp, partial, err := executeWithFallback(profile, provider, c.providers, func(p Provider, attempt *ModelProfile) (EmbedResponse, AICallMeta, error) {
		return p.Embed(ctx, attempt, input)
	})
	meta, mergeErr := c.builder.merge(profile, input.Metadata, partial)
	if mergeErr != nil && err == nil {
		err = mergeErr
	}
	return resp, meta, err
}

// Stream implements AIClient.
func (c *Client) Stream(ctx context.Context, profileName string, payload CompletePayload) (<-chan AIStreamEvent, error) {
	if len(payload.Messages) == 0 {
		return nil, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "messages must be non-empty", false)
	}
	profile, provider, err := c.dispatch(profileName, CapabilityChat)
	if err != nil {
		return nil, err
	}
	return provider.Stream(ctx, profile, payload)
}

func (c *Client) dispatch(profileName string, expectedCapability Capability) (*ModelProfile, Provider, error) {
	if c.resolver == nil {
		return nil, nil, fmt.Errorf("aiclient: no profile resolver configured")
	}
	profile, err := c.resolver.Resolve(profileName)
	if err != nil {
		return nil, nil, err
	}
	if profile.Status != ProfileStatusActive {
		return profile, nil, unsupportedCapabilityError(
			"profile %q has status %q: %s",
			profileName,
			profile.Status,
			profile.UnsupportedReason,
		)
	}
	if len(profile.Fallback) > 2 {
		return profile, nil, fallbackExhaustedError("profile %q fallback chain has %d hops; maximum is 2", profileName, len(profile.Fallback))
	}
	if expectedCapability != "" && profile.Capability != expectedCapability {
		return profile, nil, unsupportedCapabilityError(
			"profile %q has capability %q, caller expected %q",
			profileName,
			profile.Capability,
			expectedCapability,
		)
	}
	provider, ok := c.providers[profile.Default.ProviderRef]
	if !ok {
		return profile, nil, unsupportedCapabilityError(
			"profile %q references inactive provider %q for capability %q",
			profileName,
			profile.Default.ProviderRef,
			profile.Capability,
		)
	}
	return profile, provider, nil
}

func unsupportedCapabilityError(format string, args ...any) error {
	return sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, fmt.Sprintf(format, args...), false)
}

func fallbackExhaustedError(format string, args ...any) error {
	return sharederrors.Wrap(sharederrors.CodeAiFallbackExhausted, fmt.Sprintf(format, args...), true)
}

func executeWithFallback[T any](
	profile *ModelProfile,
	primary Provider,
	providers map[string]Provider,
	invoke func(Provider, *ModelProfile) (T, AICallMeta, error),
) (T, AICallMeta, error) {
	chain := []string{fallbackHop(profile.Default)}
	out, meta, err := invoke(primary, profileWithProviderConfig(profile, profile.Default))
	meta = attemptMeta(meta, profile.Default, chain)
	if err == nil {
		return out, meta, nil
	}

	lastMeta := meta
	lastErr := err
	attemptedFallback := false
	var zero T

	for _, fb := range profile.Fallback {
		if !fallbackConditionMatches(lastErr, fb.When) {
			break
		}
		attemptedFallback = true
		chain = append(chain, fallbackHop(fb.ProviderConfig))
		provider, ok := providers[fb.ProviderRef]
		if !ok {
			exhausted := fallbackExhaustedError("fallback provider %q is not registered for profile %q", fb.ProviderRef, profile.Name)
			lastMeta = fallbackFailureMeta(profile, fb.ProviderConfig, chain, exhausted)
			return zero, lastMeta, exhausted
		}

		out, meta, err = invoke(provider, profileWithProviderConfig(profile, fb.ProviderConfig))
		meta = attemptMeta(meta, fb.ProviderConfig, chain)
		if err == nil {
			return out, meta, nil
		}
		lastMeta = meta
		lastErr = err
	}

	if attemptedFallback {
		exhausted := fallbackExhaustedError("fallback chain exhausted for profile %q after %d attempts", profile.Name, len(chain))
		lastMeta.ErrorCode = sharederrors.CodeAiFallbackExhausted
		lastMeta.ValidationStatus = ValidationStatusInvalid
		lastMeta.FallbackChain = append([]string(nil), chain...)
		return zero, lastMeta, exhausted
	}
	return zero, lastMeta, lastErr
}

func fallbackConditionMatches(err error, when []string) bool {
	if len(when) == 0 {
		return true
	}
	code := errorCode(err)
	for _, raw := range when {
		cond := strings.ToLower(strings.TrimSpace(raw))
		switch cond {
		case "timeout", "provider_timeout", "ai_provider_timeout":
			if code == sharederrors.CodeAiProviderTimeout {
				return true
			}
		default:
			if strings.EqualFold(raw, code) {
				return true
			}
		}
	}
	return false
}

func profileWithProviderConfig(profile *ModelProfile, cfg ProviderConfig) *ModelProfile {
	attempt := *profile
	attempt.Default = cfg
	return &attempt
}

func attemptMeta(meta AICallMeta, cfg ProviderConfig, chain []string) AICallMeta {
	if meta.Provider == "" {
		meta.Provider = cfg.ProviderRef
	}
	if meta.ModelID == "" {
		meta.ModelID = cfg.Model
	}
	meta.FallbackChain = append([]string(nil), chain...)
	return meta
}

func fallbackFailureMeta(profile *ModelProfile, cfg ProviderConfig, chain []string, err error) AICallMeta {
	return AICallMeta{
		Provider:            cfg.ProviderRef,
		ModelID:             cfg.Model,
		Capability:          profile.Capability,
		ModelProfileName:    profile.Name,
		ModelProfileVersion: profile.Version,
		FallbackChain:       append([]string(nil), chain...),
		Route:               profile.Route,
		ValidationStatus:    ValidationStatusInvalid,
		ErrorCode:           errorCode(err),
	}
}

func fallbackHop(cfg ProviderConfig) string {
	if cfg.Model == "" {
		return cfg.ProviderRef
	}
	return cfg.ProviderRef + "/" + cfg.Model
}

func failureMeta(profileName string, profile *ModelProfile, err error) AICallMeta {
	out := AICallMeta{
		ModelProfileName: profileName,
		ValidationStatus: ValidationStatusInvalid,
		ErrorCode:        errorCode(err),
	}
	if profile == nil {
		return out
	}
	out.Capability = profile.Capability
	out.Provider = profile.Default.ProviderRef
	out.ModelID = profile.Default.Model
	out.ModelProfileVersion = profile.Version
	out.Route = profile.Route
	if profile.Default.ProviderRef != "" {
		out.FallbackChain = []string{profile.Default.ProviderRef}
	}
	return out
}

func errorCode(err error) string {
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}
