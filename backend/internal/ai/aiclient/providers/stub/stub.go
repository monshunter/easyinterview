package stub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// Name is the canonical provider name. It also matches the value Model
// Profiles set in Default.ProviderRef when they want to route to stub.
const Name = "stub"

// ProviderName is exposed for cross-package references that prefer a typed
// name over the bare const.
func ProviderName() string { return Name }

// ErrNotAllowed is returned by New when stub instantiation is rejected
// because the resolved AppEnv is not "test" and the caller did not pass
// WithAllowed.
var ErrNotAllowed = errors.New("stub provider only allowed when WithAppEnv(test) or WithAllowed(true) is set")

type options struct {
	allowed bool
	appEnv  string
}

// Option mutates stub construction.
type Option func(*options)

// WithAllowed permits the stub provider outside APP_ENV=test. Plan 001 reads
// the AppEnv directly from the aiclient.Config; this option is the override
// the AIClient uses for callers who explicitly opt in (e.g. integration
// tests running with APP_ENV=ci).
func WithAllowed(allowed bool) Option {
	return func(o *options) { o.allowed = allowed }
}

// WithAppEnv passes the boot AppEnv into the stub factory. Required by
// secrets-and-config spec §4.1 boundary lint: the stub must not read
// os.Getenv directly. AIClient passes the loader-resolved value.
func WithAppEnv(env string) Option {
	return func(o *options) { o.appEnv = env }
}

// Provider is the deterministic stub.
type Provider struct{}

// New constructs a Provider after validating that stub instantiation is
// allowed for the current environment. Callers must pass WithAppEnv from
// the resolved Loader (or WithAllowed(true) for explicit overrides);
// reading os.Getenv directly is forbidden by the secrets-and-config
// boundary lint.
func New(opts ...Option) (*Provider, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if o.appEnv != aiclient.AppEnvTest && !o.allowed {
		return nil, ErrNotAllowed
	}
	return &Provider{}, nil
}

// Name implements aiclient.Provider.
func (p *Provider) Name() string { return Name }

// Complete implements aiclient.Provider.
func (p *Provider) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, fmt.Errorf("stub: profile is nil")
	}
	seed, err := canonicalSeed(profile.Name, payload)
	if err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
	}
	content := fmt.Sprintf("stub:%s:%s", profile.Name, seed[:16])
	resp := aiclient.CompleteResponse{
		Content:      content,
		FinishReason: "stop",
	}
	meta := aiclient.AICallMeta{
		Provider:     Name,
		ModelFamily:  "stub",
		ModelID:      profile.Default.Model,
		InputTokens:  countTokens(payload),
		OutputTokens: len(content),
		LatencyMs:    1,
	}
	return resp, meta, nil
}

// Embed implements aiclient.Provider.
func (p *Provider) Embed(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.EmbedInput) (aiclient.EmbedResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.EmbedResponse{}, aiclient.AICallMeta{}, fmt.Errorf("stub: profile is nil")
	}
	vectors := make([][]float64, len(input.Texts))
	totalIn := 0
	for i, text := range input.Texts {
		seed := sha256.Sum256([]byte(profile.Name + "\n" + text))
		vec := make([]float64, 4)
		for j := range vec {
			vec[j] = float64(seed[j]) / 255.0
		}
		vectors[i] = vec
		totalIn += len(text)
	}
	meta := aiclient.AICallMeta{
		Provider:     Name,
		ModelFamily:  "stub",
		ModelID:      profile.Default.Model,
		InputTokens:  totalIn,
		OutputTokens: len(vectors),
		LatencyMs:    1,
	}
	return aiclient.EmbedResponse{Vectors: vectors}, meta, nil
}

// Transcribe implements aiclient.Provider with deterministic transcript text.
func (p *Provider) Transcribe(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, fmt.Errorf("stub: profile is nil")
	}
	seed := transcriptionSeed(profile.Name, input)
	text := fmt.Sprintf("stub transcript:%s:%s", profile.Name, seed[:16])
	meta := aiclient.AICallMeta{
		Provider:     Name,
		ModelFamily:  "stub",
		ModelID:      profile.Default.Model,
		InputTokens:  len(input.Audio),
		OutputTokens: len(text),
		LatencyMs:    1,
	}
	return aiclient.TranscriptionResponse{Text: text}, meta, nil
}

// Stream implements aiclient.Provider as a one-shot terminal event channel.
// Plan 001 freezes the stream contract; plan 002 will replace this with a
// real SSE / chunked consumer.
func (p *Provider) Stream(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	resp, meta, err := p.Complete(ctx, profile, payload)
	ch := make(chan aiclient.AIStreamEvent, 1)
	if err != nil {
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: err.Error()}
		close(ch)
		return ch, nil
	}
	finalMeta := meta
	finalMeta.OutputTokens = len(resp.Content)
	ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventDone, Meta: &finalMeta}
	close(ch)
	return ch, nil
}

func canonicalSeed(profileName string, payload aiclient.CompletePayload) (string, error) {
	canonical := struct {
		Profile  string                `json:"profile"`
		Messages []aiclient.Message    `json:"messages"`
		Metadata aiclient.CallMetadata `json:"metadata"`
	}{
		Profile:  profileName,
		Messages: payload.Messages,
		Metadata: payload.Metadata,
	}
	b, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func countTokens(payload aiclient.CompletePayload) int {
	n := 0
	for _, m := range payload.Messages {
		n += len(m.Content)
	}
	return n
}

func transcriptionSeed(profileName string, input aiclient.TranscriptionInput) string {
	sum := sha256.New()
	sum.Write([]byte(profileName))
	sum.Write([]byte{0})
	sum.Write([]byte(input.Filename))
	sum.Write([]byte{0})
	sum.Write([]byte(input.ContentType))
	sum.Write([]byte{0})
	sum.Write([]byte(input.Language))
	sum.Write([]byte{0})
	sum.Write([]byte(input.Prompt))
	sum.Write([]byte{0})
	sum.Write(input.Audio)
	return hex.EncodeToString(sum.Sum(nil))
}
