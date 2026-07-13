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
	"compress/gzip"
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

const maxResponseBodyBytes int64 = 4 << 20

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
	if err := applyParams(profile.Default.Params, &req); err != nil {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "judge_compatible: invalid profile params: "+err.Error(), false)
		return aiclient.CompleteResponse{}, a.errMeta(profile, 0, errCode), errCode
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
		errCode := stableError(sharederrors.CodeAiOutputInvalid)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}
	if len(body.Choices) == 0 {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "judge_compatible: response missing choices", false)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}

	choice := body.Choices[0]
	meta := aiclient.AICallMeta{
		Provider:     a.providerRef,
		ModelID:      profile.Default.Model,
		Capability:   aiclient.CapabilityJudge,
		InputTokens:  body.Usage.PromptTokens,
		OutputTokens: body.Usage.CompletionTokens,
		LatencyMs:    latencyMs,
	}
	if strings.TrimSpace(choice.Message.Content) == "" {
		finishReason := safeFinishReason(choice.FinishReason)
		errCode := sharederrors.Wrap(
			sharederrors.CodeAiOutputInvalid,
			fmt.Sprintf(
				"judge_compatible: empty response content finish_reason=%s completion_tokens=%d reasoning_content_present=%t",
				finishReason,
				body.Usage.CompletionTokens,
				choice.Message.ReasoningContent != "",
			),
			false,
		)
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
		meta.ErrorCode = errorCode(errCode)
		meta.ModelProfileName = profile.Name
		meta.ModelProfileVersion = profile.Version
		meta.Route = profile.Route
		return aiclient.CompleteResponse{}, meta, errCode
	}
	out := aiclient.CompleteResponse{
		Content:      choice.Message.Content,
		FinishReason: safeFinishReason(choice.FinishReason),
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
		return nil, 0, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}
	reqCtx := ctx
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, a.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, 0, stableError(sharederrors.CodeAiProviderTimeout)
	}
	defer resp.Body.Close()
	body, err := readResponseBody(resp)
	if err != nil {
		return nil, resp.StatusCode, err
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
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	return stableError(sharederrors.CodeAiOutputInvalid)
}

func errorCode(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return ""
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	reader := io.Reader(resp.Body)
	compressed := false
	if !resp.Uncompressed {
		switch strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Encoding"))) {
		case "", "identity":
		case "gzip":
			zr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return nil, stableError(sharederrors.CodeAiOutputInvalid)
			}
			defer zr.Close()
			reader = zr
			compressed = true
		default:
			return nil, stableError(sharederrors.CodeAiOutputInvalid)
		}
	}
	body, err := io.ReadAll(io.LimitReader(reader, maxResponseBodyBytes+1))
	if err != nil {
		if compressed {
			return nil, stableError(sharederrors.CodeAiOutputInvalid)
		}
		return nil, stableError(sharederrors.CodeAiProviderTimeout)
	}
	if int64(len(body)) > maxResponseBodyBytes {
		return nil, stableError(sharederrors.CodeAiOutputInvalid)
	}
	return body, nil
}

func stableError(code string) *sharederrors.APIError {
	meta := sharederrors.CodeRegistry[code]
	return sharederrors.Wrap(code, meta.Message, meta.Retryable)
}

func safeFinishReason(value string) string {
	switch value {
	case "", "stop", "length", "tool_calls", "content_filter", "function_call":
		return value
	default:
		return "unknown"
	}
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

func applyParams(params map[string]any, req *chatCompletionsRequest) error {
	if params == nil {
		return nil
	}
	if temp, ok := floatParam(params, "temperature"); ok {
		req.Temperature = &temp
	}
	if raw, ok := params["thinking"]; ok {
		mode, ok := raw.(string)
		if !ok || (mode != "enabled" && mode != "disabled") {
			return errors.New("thinking must be enabled or disabled")
		}
		req.Thinking = &thinkingConfig{Type: mode}
	}
	if raw, ok := params["response_format"]; ok {
		format, ok := raw.(string)
		if !ok || format != "json_object" {
			return errors.New("response_format must be json_object")
		}
		req.ResponseFormat = &responseFormat{Type: format}
	}
	return nil
}

type wireMessage struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type thinkingConfig struct {
	Type string `json:"type"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionsRequest struct {
	Model          string          `json:"model"`
	Messages       []wireMessage   `json:"messages"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Stream         bool            `json:"stream"`
	Temperature    *float64        `json:"temperature,omitempty"`
	Thinking       *thinkingConfig `json:"thinking,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
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
