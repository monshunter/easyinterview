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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/internal/openaisdk"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/internal/responsebody"
	platformconfig "github.com/monshunter/easyinterview/backend/internal/platform/config"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

// Name is the adapter protocol name. Concrete adapters identify themselves by
// Provider Registry ref so profiles route through provider_ref.
const Name = "judge_compatible"

// PathChatCompletions is the OpenAI-compatible endpoint the judge wire posts to.
const PathChatCompletions = "/v1/chat/completions"

// Options configures the adapter.
type Options struct {
	Provider             providerregistry.ResolvedProvider
	HTTPClient           *http.Client
	MaxResponseBodyBytes int64
}

// Adapter is the concrete aiclient.Provider implementation for judge calls.
type Adapter struct {
	providerRef string
	sdkClient   openai.Client
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
	if opts.MaxResponseBodyBytes <= 0 {
		opts.MaxResponseBodyBytes = platformconfig.DefaultContentLimits().AIProviderMaxResponseBodyBytes
	}
	return &Adapter{
		providerRef: opts.Provider.Entry.Name,
		sdkClient:   openaisdk.NewClient(opts.Provider.BaseURL, opts.Provider.APIKey, hc, opts.MaxResponseBodyBytes),
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

	req, requestOptions, err := sdkJudgeRequest(profile, payload)
	if err != nil {
		errCode := stableError(sharederrors.CodeAiOutputInvalid)
		return aiclient.CompleteResponse{}, a.errMeta(profile, 0, errCode), errCode
	}

	start := time.Now()
	requestContext := ctx
	if profile.TimeoutMs > 0 {
		var cancel context.CancelFunc
		requestContext, cancel = context.WithTimeout(ctx, time.Duration(profile.TimeoutMs)*time.Millisecond)
		defer cancel()
	}
	var rawResponse *http.Response
	requestOptions = append(requestOptions, option.WithResponseInto(&rawResponse))
	completion, err := a.sdkClient.Chat.Completions.New(requestContext, req, requestOptions...)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		errCode := mapJudgeSDKError(requestContext, rawResponse, err)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}
	if completion == nil || len(completion.Choices) == 0 {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "judge_compatible: response missing choices", false)
		return aiclient.CompleteResponse{}, a.errMeta(profile, latencyMs, errCode), errCode
	}

	choice := completion.Choices[0]
	meta := aiclient.AICallMeta{
		Provider:     a.providerRef,
		ModelID:      profile.Default.Model,
		Capability:   aiclient.CapabilityJudge,
		InputTokens:  int(completion.Usage.PromptTokens),
		OutputTokens: int(completion.Usage.CompletionTokens),
		LatencyMs:    latencyMs,
	}
	if strings.TrimSpace(choice.Message.Content) == "" {
		finishReason := safeFinishReason(choice.FinishReason)
		errCode := sharederrors.Wrap(
			sharederrors.CodeAiOutputInvalid,
			fmt.Sprintf(
				"judge_compatible: empty response content finish_reason=%s completion_tokens=%d reasoning_content_present=%t",
				finishReason,
				completion.Usage.CompletionTokens,
				reasoningContentPresent(choice.Message),
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

func sdkJudgeRequest(profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (openai.ChatCompletionNewParams, []option.RequestOption, error) {
	request := openai.ChatCompletionNewParams{
		Model:    profile.Default.Model,
		Messages: convertJudgeSDKMessages(payload.Messages),
	}
	if profile.MaxTokens > 0 {
		request.MaxTokens = openai.Int(int64(profile.MaxTokens))
	}
	if temperature, ok := floatParam(profile.Default.Params, "temperature"); ok {
		request.Temperature = openai.Float(temperature)
	}

	var requestOptions []option.RequestOption
	if raw, ok := profile.Default.Params["thinking"]; ok {
		mode, valid := raw.(string)
		if !valid || (mode != "enabled" && mode != "disabled") {
			return openai.ChatCompletionNewParams{}, nil, errors.New("thinking must be enabled or disabled")
		}
		requestOptions = append(requestOptions, option.WithJSONSet("thinking", map[string]string{"type": mode}))
	}
	if raw, ok := profile.Default.Params["response_format"]; ok {
		format, valid := raw.(string)
		if !valid || format != "json_object" {
			return openai.ChatCompletionNewParams{}, nil, errors.New("response_format must be json_object")
		}
		responseFormat := shared.NewResponseFormatJSONObjectParam()
		request.ResponseFormat.OfJSONObject = &responseFormat
	}
	return request, requestOptions, nil
}

func convertJudgeSDKMessages(messages []aiclient.Message) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, message := range messages {
		switch message.Role {
		case "system":
			out = append(out, openai.SystemMessage(message.Content))
		case "user":
			out = append(out, openai.UserMessage(message.Content))
		case "assistant":
			out = append(out, openai.AssistantMessage(message.Content))
		case "developer":
			out = append(out, openai.DeveloperMessage(message.Content))
		default:
			out = append(out, param.Override[openai.ChatCompletionMessageParamUnion](map[string]string{
				"role":    message.Role,
				"content": message.Content,
			}))
		}
	}
	return out
}

func reasoningContentPresent(message openai.ChatCompletionMessage) bool {
	var raw map[string]json.RawMessage
	if json.Unmarshal([]byte(message.RawJSON()), &raw) != nil {
		return false
	}
	value, ok := raw["reasoning_content"]
	return ok && len(value) > 0 && string(value) != "null" && string(value) != `""`
}

func mapJudgeSDKError(ctx context.Context, response *http.Response, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	if errors.Is(err, responsebody.ErrInvalid) {
		return stableError(sharederrors.CodeAiOutputInvalid)
	}
	if errors.Is(err, responsebody.ErrRead) {
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	var apiError *openai.Error
	if errors.As(err, &apiError) && apiError.StatusCode >= http.StatusInternalServerError {
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	if response == nil {
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	return stableError(sharederrors.CodeAiOutputInvalid)
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
