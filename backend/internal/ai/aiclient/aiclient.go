package aiclient

import "context"

// AIClient is the only entry point business code may depend on for AI calls
// (spec §2.1 / D-1). The interface is provider-neutral: callers reference a
// Model Profile name and never see vendor-specific request or response
// shapes.
type AIClient interface {
	// Complete runs a chat / completion request and returns the structured
	// response plus the AICallMeta filled by the client.
	Complete(ctx context.Context, profileName string, payload CompletePayload) (CompleteResponse, AICallMeta, error)

	// Embed runs an embedding request.
	Embed(ctx context.Context, profileName string, input EmbedInput) (EmbedResponse, AICallMeta, error)

	// Transcribe runs an audio transcription request. The payload carries
	// audio bytes plus filename/content type metadata; callers still reference
	// only a Model Profile name.
	Transcribe(ctx context.Context, profileName string, input TranscriptionInput) (TranscriptionResponse, AICallMeta, error)

	// Stream returns a channel of AIStreamEvent values whose lifecycle is
	// frozen by spec §4.1: at most one terminal `done` or `error` event,
	// channel closes after the terminal event. Plan 001 only ships the type
	// contract; full provider-side streaming consumption lands in plan 002.
	Stream(ctx context.Context, profileName string, payload CompletePayload) (<-chan AIStreamEvent, error)
}

// Provider is the lower-level interface every concrete backend implements
// (stub, openai_compatible, ...). It is internal to the aiclient package
// tree and is never exposed to business code.
//
// A Provider receives the resolved *ModelProfile so it can translate the
// profile into the on-the-wire request without re-resolving by name.
// Providers MUST fill at minimum Provider, ModelFamily, ModelID, InputTokens,
// OutputTokens, and any FallbackChain entries reported by the provider
// endpoint. The Client's metaBuilder fills the remaining profile- and
// metadata-derived fields.
type Provider interface {
	Name() string
	Complete(ctx context.Context, profile *ModelProfile, payload CompletePayload) (CompleteResponse, AICallMeta, error)
	Embed(ctx context.Context, profile *ModelProfile, input EmbedInput) (EmbedResponse, AICallMeta, error)
	Transcribe(ctx context.Context, profile *ModelProfile, input TranscriptionInput) (TranscriptionResponse, AICallMeta, error)
	Stream(ctx context.Context, profile *ModelProfile, payload CompletePayload) (<-chan AIStreamEvent, error)
}

// ProviderResolver materializes providers by registry provider_ref. Production
// bootstrap uses it so profile hot reload can route to provider refs without
// pre-registering a static in-memory map.
type ProviderResolver interface {
	ResolveProvider(ref string) (Provider, error)
}
