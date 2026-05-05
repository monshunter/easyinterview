package aiclient

import (
	"context"
	"fmt"

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
//     ProviderBaseURL/APIKey: ErrMissingProviderConfig.
//   - cfg.AppEnv != "test" with missing ProviderBaseURL or ProviderAPIKey:
//     ErrMissingProviderConfig regardless of WithStubAllowed; non-test
//     deployments must point at a real OpenAI-compatible endpoint.
func New(cfg Config, opts ...Option) (*Client, error) {
	o := &clientOptions{
		providers: map[string]Provider{},
	}
	for _, opt := range opts {
		opt(o)
	}

	if cfg.AppEnv == AppEnvTest {
		if !o.stubAllowed && (cfg.ProviderBaseURL == "" || cfg.ProviderAPIKey == "") {
			return nil, ErrMissingProviderConfig
		}
	} else {
		if cfg.ProviderBaseURL == "" || cfg.ProviderAPIKey == "" {
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
		return CompleteResponse{}, AICallMeta{ModelProfileName: profileName, ErrorCode: err.Error()}, err
	}

	resp, partial, err := provider.Complete(ctx, profile, payload)
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
		return EmbedResponse{}, AICallMeta{ModelProfileName: profileName, ErrorCode: err.Error()}, err
	}

	resp, partial, err := provider.Embed(ctx, profile, input)
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
	if profile.Capability == CapabilitySTT {
		return nil, nil, ErrCapabilityNotImplemented
	}
	if expectedCapability != "" && profile.Capability != expectedCapability {
		return nil, nil, fmt.Errorf("aiclient: profile %q has capability %q, caller expected %q", profileName, profile.Capability, expectedCapability)
	}
	provider, ok := c.providers[profile.Default.ProviderRef]
	if !ok {
		return nil, nil, fmt.Errorf("aiclient: provider %q not registered for profile %q", profile.Default.ProviderRef, profileName)
	}
	return profile, provider, nil
}
