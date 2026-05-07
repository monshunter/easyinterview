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
	toolCalls := stubToolCalls(profile.Name, seed, payload)
	content := fmt.Sprintf("stub:%s:%s", profile.Name, seed[:16])
	resp := aiclient.CompleteResponse{
		Content:      content,
		FinishReason: "stop",
		ToolCalls:    toolCalls,
	}
	if len(toolCalls) > 0 {
		resp.FinishReason = "tool_calls"
	}
	meta := aiclient.AICallMeta{
		Provider:        Name,
		ModelFamily:     "stub",
		ModelID:         profile.Default.Model,
		InputTokens:     countTokens(payload),
		OutputTokens:    len(content),
		LatencyMs:       1,
		ToolInvocations: summarizeToolCalls(toolCalls),
	}
	return resp, meta, nil
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
		Profile    string                `json:"profile"`
		Messages   []aiclient.Message    `json:"messages"`
		Metadata   aiclient.CallMetadata `json:"metadata"`
		Tools      []aiclient.Tool       `json:"tools,omitempty"`
		ToolChoice *aiclient.ToolChoice  `json:"toolChoice,omitempty"`
	}{
		Profile:    profileName,
		Messages:   payload.Messages,
		Metadata:   payload.Metadata,
		Tools:      payload.Tools,
		ToolChoice: payload.ToolChoice,
	}
	b, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func stubToolCalls(profileName, seed string, payload aiclient.CompletePayload) []aiclient.ToolCall {
	if len(payload.Tools) == 0 {
		return nil
	}
	tool := payload.Tools[0]
	if payload.ToolChoice != nil {
		switch payload.ToolChoice.Mode {
		case aiclient.ToolChoiceModeNone:
			return nil
		case aiclient.ToolChoiceModeTool:
			found, ok := findTool(payload.Tools, payload.ToolChoice.Name)
			if !ok {
				return nil
			}
			tool = found
		}
	}
	argumentSeed := sha256.Sum256([]byte(profileName + "\n" + seed + "\n" + tool.Name + "\n" + string(tool.Parameters)))
	args, err := json.Marshal(map[string]any{
		"stub_seed": hex.EncodeToString(argumentSeed[:])[:16],
	})
	if err != nil {
		return nil
	}
	return []aiclient.ToolCall{{
		ID:        "stub_" + hex.EncodeToString(argumentSeed[:])[:12],
		Name:      tool.Name,
		Arguments: json.RawMessage(args),
	}}
}

func findTool(tools []aiclient.Tool, name string) (aiclient.Tool, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return aiclient.Tool{}, false
}

func summarizeToolCalls(calls []aiclient.ToolCall) []aiclient.ToolInvocationMeta {
	if len(calls) == 0 {
		return nil
	}
	out := make([]aiclient.ToolInvocationMeta, 0, len(calls))
	for _, call := range calls {
		sum := sha256.Sum256(call.Arguments)
		out = append(out, aiclient.ToolInvocationMeta{
			Name:            call.Name,
			ArgumentsHash:   hex.EncodeToString(sum[:]),
			ArgumentsLength: len(call.Arguments),
		})
	}
	return out
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
