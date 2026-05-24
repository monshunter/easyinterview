// Package judgecompatible is the A3/F3 cross-owner additive provider adapter
// for the CapabilityJudge protocol (ProviderProtocolJudgeCompatible). It speaks
// the OpenAI-compatible Chat Completions JSON subset on the wire so a real
// LLM judge endpoint can be driven by the offline eval harness, while keeping
// protocol/capability/meta tagged as judge so judge traffic never masquerades
// as a business chat call. Plan prompt-rubric-registry/004 §2.2.
//
// Only Complete is implemented; Transcribe/Stream/Synthesize fail-close with
// AI_UNSUPPORTED_CAPABILITY. Raw judge prompt/output content is never logged or
// persisted by this adapter (privacy red-line continued from A3).
package judgecompatible

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// Name is the adapter protocol name. Concrete adapters identify themselves by
// Provider Registry ref so profiles route through provider_ref.
const Name = "judge_compatible"

// PathChatCompletions is the OpenAI-compatible endpoint the judge wire posts to.
const PathChatCompletions = "/v1/chat/completions"

// Options configures the adapter.
type Options struct {
	Provider   providerregistry.ResolvedProvider
	HTTPClient *http.Client
}

// Adapter is the concrete aiclient.Provider implementation for judge calls.
type Adapter struct {
	providerRef string
	baseURL     string
	apiKey      string
	client      *http.Client
}

// New constructs an Adapter from a Provider Registry entry materialized by A4
// SecretSource. Missing base URL / API key fail fast (spec §4.4 / A3 secret
// fail-fast continued).
func New(opts Options) (*Adapter, error) {
	if opts.Provider.Entry.Name == "" {
		return nil, errors.New("judge_compatible: resolved provider is required")
	}
	if opts.Provider.Entry.Protocol != aiclient.ProviderProtocolJudgeCompatible {
		return nil, fmt.Errorf("judge_compatible: provider %q protocol must be %q", opts.Provider.Entry.Name, aiclient.ProviderProtocolJudgeCompatible)
	}
	if opts.Provider.BaseURL == "" {
		return nil, errors.New("judge_compatible: resolved provider BaseURL is required")
	}
	if opts.Provider.APIKey == "" {
		return nil, errors.New("judge_compatible: resolved provider APIKey is required")
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{}
	}
	return &Adapter{
		providerRef: opts.Provider.Entry.Name,
		baseURL:     normalizeBaseURL(opts.Provider.BaseURL),
		apiKey:      opts.Provider.APIKey,
		client:      hc,
	}, nil
}

// Name implements aiclient.Provider.
func (a *Adapter) Name() string { return a.providerRef }

// Complete implements aiclient.Provider by posting the judge prompt to an
// OpenAI-compatible Chat Completions endpoint and returning the model content.
func (a *Adapter) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("judge_compatible: profile is nil")
	}

	req := chatCompletionsRequest{
		Model:    profile.Default.Model,
		Messages: convertMessages(payload.Messages),
		Stream:   false,
	}
	if profile.MaxTokens > 0 {
		req.MaxTokens = profile.MaxTokens
	}
	if temp, ok := floatParam(profile.Default.Params, "temperature"); ok {
		req.Temperature = &temp
	}

	start := time.Now()
	resp, status, err := a.postJSON(ctx, profile.TimeoutMs, PathChatCompletions, req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, err), err
	}
	if status >= 400 {
		errCode := mapHTTPError(status)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}

	var body chatCompletionsResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "judge_compatible: parse response: "+err.Error(), false)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}
	if len(body.Choices) == 0 {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "judge_compatible: response missing choices", false)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}

	choice := body.Choices[0]
	out := aiclient.CompleteResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
	}
	model := body.Model
	if model == "" {
		model = profile.Default.Model
	}
	meta := aiclient.AICallMeta{
		Provider:     a.providerRef,
		ModelID:      model,
		Capability:   aiclient.CapabilityJudge,
		InputTokens:  body.Usage.PromptTokens,
		OutputTokens: body.Usage.CompletionTokens,
		LatencyMs:    latencyMs,
	}
	return out, meta, nil
}

// Transcribe fail-closes: the judge protocol is chat-shaped only.
func (a *Adapter) Transcribe(_ context.Context, profile *aiclient.ModelProfile, _ aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, a.errMeta(profile, 0, nil), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "judge_compatible does not support transcription", false)
}

// Stream fail-closes: judge calls are single-shot.
func (a *Adapter) Stream(_ context.Context, _ *aiclient.ModelProfile, _ aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "judge_compatible does not support streaming", false)
}

// Synthesize fail-closes: the judge protocol is chat-shaped only.
func (a *Adapter) Synthesize(_ context.Context, profile *aiclient.ModelProfile, _ aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, a.errMeta(profile, 0, nil), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "judge_compatible does not support synthesis", false)
}

func (a *Adapter) postJSON(ctx context.Context, timeoutMs int, path string, payload any) ([]byte, int, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("judge_compatible: marshal request: %w", err)
	}
	reqCtx := ctx
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, a.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, fmt.Errorf("judge_compatible: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, 0, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "judge_compatible: request failed: "+err.Error(), true)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("judge_compatible: read response: %w", err)
	}
	return body, resp.StatusCode, nil
}

func (a *Adapter) errMeta(profile *aiclient.ModelProfile, latencyMs int64, err error) aiclient.AICallMeta {
	meta := aiclient.AICallMeta{
		Provider:         a.providerRef,
		Capability:       aiclient.CapabilityJudge,
		LatencyMs:        latencyMs,
		ValidationStatus: aiclient.ValidationStatusInvalid,
		ErrorCode:        errorCode(err),
	}
	if profile != nil {
		meta.ModelID = profile.Default.Model
		meta.ModelProfileName = profile.Name
		meta.ModelProfileVersion = profile.Version
		meta.Route = profile.Route
	}
	return meta
}

func mapHTTPError(status int) error {
	if status >= 500 {
		return sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, fmt.Sprintf("judge_compatible: upstream %d", status), true)
	}
	return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, fmt.Sprintf("judge_compatible: upstream %d", status), false)
}

func errorCode(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return err.Error()
}

func normalizeBaseURL(raw string) string {
	return strings.TrimRight(raw, "/")
}

func convertMessages(messages []aiclient.Message) []wireMessage {
	out := make([]wireMessage, len(messages))
	for i, m := range messages {
		out[i] = wireMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func floatParam(params map[string]any, key string) (float64, bool) {
	if params == nil {
		return 0, false
	}
	switch v := params[key].(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsRequest struct {
	Model       string        `json:"model"`
	Messages    []wireMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
	Temperature *float64      `json:"temperature,omitempty"`
}

type chatCompletionsResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message      wireMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}
