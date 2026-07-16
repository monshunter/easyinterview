package openaicompatible

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/shared"
)

// Name is the adapter protocol name. Concrete Adapter instances identify
// themselves by Provider Registry ref so profiles route through provider_ref.
const Name = "openai_compatible"

// Path constants are exported so the mockserver helper can reuse them.
const (
	PathChatCompletions = "/v1/chat/completions"
	PathTranscriptions  = "/v1/audio/transcriptions"
)

// Header names retained for wire compatibility and mock fixtures. Response
// values are untrusted and never override canonical profile metadata.
const (
	HeaderRequestID    = "X-Request-ID"
	HeaderFallbackFrom = "X-Fallback-From"
	HeaderFallbackTo   = "X-Fallback-To"
	HeaderRoute        = "X-Route"
)

// Options configures the adapter.
type Options struct {
	Provider             providerregistry.ResolvedProvider
	HTTPClient           *http.Client
	MaxResponseBodyBytes int64
}

// Adapter is the concrete aiclient.Provider implementation.
type Adapter struct {
	providerRef string
	sdkClient   openai.Client
}

// New constructs an Adapter from a Provider Registry entry materialized by A4
// SecretSource. Raw global base URL / API key values are not accepted here.
func New(opts Options) (*Adapter, error) {
	if opts.Provider.Entry.Name == "" {
		return nil, errors.New("openai_compatible: resolved provider is required")
	}
	if opts.Provider.Entry.Protocol != aiclient.ProviderProtocolOpenAICompatible {
		return nil, fmt.Errorf("openai_compatible: provider %q protocol must be %q", opts.Provider.Entry.Name, aiclient.ProviderProtocolOpenAICompatible)
	}
	if opts.Provider.BaseURL == "" {
		return nil, errors.New("openai_compatible: resolved provider BaseURL is required")
	}
	if opts.Provider.APIKey == "" {
		return nil, errors.New("openai_compatible: resolved provider APIKey is required")
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

// Complete implements aiclient.Provider.
func (a *Adapter) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("openai_compatible: profile is nil")
	}

	req, requestOptions, err := sdkCompleteRequest(profile, payload)
	if err != nil {
		errCode := stableError(sharederrors.CodeAiProviderConfigInvalid)
		return aiclient.CompleteResponse{}, a.errMeta(profile, nil, 0, errCode), errCode
	}
	if requestID := aiclient.RequestIDFromContext(ctx); requestID != "" {
		requestOptions = append(requestOptions, option.WithHeader(HeaderRequestID, requestID))
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
	headers := responseHeaders(rawResponse)
	if err != nil {
		errCode := mapSDKError(requestContext, rawResponse, err)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}
	if completion == nil || len(completion.Choices) == 0 {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: response missing choices", false)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}

	choice := completion.Choices[0]
	out := aiclient.CompleteResponse{
		Content:      choice.Message.Content,
		FinishReason: safeFinishReason(choice.FinishReason),
		ToolCalls:    convertSDKToolCalls(choice.Message.ToolCalls),
	}
	meta := a.buildMeta(profile, headers, latencyMs, int(completion.Usage.PromptTokens), int(completion.Usage.CompletionTokens), out.ToolCalls)
	return out, meta, nil
}

// Transcribe implements aiclient.Provider using the OpenAI-compatible Audio
// Transcriptions multipart wire shape.
func (a *Adapter) Transcribe(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("openai_compatible: profile is nil")
	}

	request := openai.AudioTranscriptionNewParams{
		File:  openai.File(bytes.NewReader(input.Audio), input.Filename, input.ContentType),
		Model: profile.Default.Model,
	}
	if input.Language != "" {
		request.Language = openai.String(input.Language)
	}
	if input.Prompt != "" {
		request.Prompt = openai.String(input.Prompt)
	}

	requestContext := ctx
	if profile.TimeoutMs > 0 {
		var cancel context.CancelFunc
		requestContext, cancel = context.WithTimeout(ctx, time.Duration(profile.TimeoutMs)*time.Millisecond)
		defer cancel()
	}
	var rawResponse *http.Response
	requestOptions := []option.RequestOption{option.WithResponseInto(&rawResponse)}
	if requestID := aiclient.RequestIDFromContext(ctx); requestID != "" {
		requestOptions = append(requestOptions, option.WithHeader(HeaderRequestID, requestID))
	}
	start := time.Now()
	response, err := a.sdkClient.Audio.Transcriptions.New(requestContext, request, requestOptions...)
	latencyMs := time.Since(start).Milliseconds()
	headers := responseHeaders(rawResponse)
	if err != nil {
		errCode := mapSDKError(requestContext, rawResponse, err)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.TranscriptionResponse{}, meta, errCode
	}
	if response == nil || response.Text == "" {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: transcription response missing text", false)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.TranscriptionResponse{}, meta, errCode
	}

	out := aiclient.TranscriptionResponse{Text: response.Text}
	meta := a.buildMeta(profile, headers, latencyMs, len(input.Audio), len(out.Text), nil)
	return out, meta, nil
}

// Stream implements aiclient.Provider. Plan 001 honors the channel contract
// by consuming OpenAI-compatible SSE chunks and mapping them to the frozen
// AIStreamEvent contract.
func (a *Adapter) Stream(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	if profile == nil {
		return nil, errors.New("openai_compatible: profile is nil")
	}
	req, requestOptions, err := sdkCompleteRequest(profile, payload)
	if err != nil {
		return nil, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}
	if requestID := aiclient.RequestIDFromContext(ctx); requestID != "" {
		requestOptions = append(requestOptions, option.WithHeader(HeaderRequestID, requestID))
	}

	streamCtx := ctx
	var cancel context.CancelFunc
	if profile.TimeoutMs > 0 {
		streamCtx, cancel = context.WithTimeout(ctx, time.Duration(profile.TimeoutMs)*time.Millisecond)
	}

	start := time.Now()
	var rawResponse *http.Response
	requestOptions = append(requestOptions, option.WithResponseInto(&rawResponse))
	stream := a.sdkClient.Chat.Completions.NewStreaming(streamCtx, req, requestOptions...)
	latencyMs := time.Since(start).Milliseconds()
	ch := make(chan aiclient.AIStreamEvent, 4)
	if err := stream.Err(); err != nil {
		_ = stream.Close()
		if cancel != nil {
			cancel()
		}
		mapped := mapSDKError(streamCtx, rawResponse, err)
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: errorCodeOf(mapped)}
		close(ch)
		return ch, nil
	}
	go a.consumeSDKStream(streamCtx, cancel, stream, responseHeaders(rawResponse), profile, latencyMs, ch)
	return ch, nil
}

func (a *Adapter) consumeSDKStream(ctx context.Context, cancel context.CancelFunc, stream *ssestream.Stream[openai.ChatCompletionChunk], headers http.Header, profile *aiclient.ModelProfile, latencyMs int64, ch chan<- aiclient.AIStreamEvent) {
	defer close(ch)
	defer stream.Close()
	if cancel != nil {
		defer cancel()
	}

	inputTokens := 0
	outputTokens := 0
	outputChars := 0
	emittedDone := false
	emitDone := func(errorCode, partialReason string) {
		if emittedDone {
			return
		}
		if outputTokens == 0 {
			outputTokens = outputChars
		}
		meta := a.buildMeta(profile, headers, latencyMs, inputTokens, outputTokens, nil)
		if errorCode != "" {
			meta.ValidationStatus = aiclient.ValidationStatusInvalid
			meta.ErrorCode = errorCode
			meta.PartialMetaReason = partialReason
		}
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventDone, Meta: &meta}
		emittedDone = true
	}

	for stream.Next() {
		chunk := stream.Current()
		if chunk.Usage.PromptTokens > 0 {
			inputTokens = int(chunk.Usage.PromptTokens)
		}
		if chunk.Usage.CompletionTokens > 0 {
			outputTokens = int(chunk.Usage.CompletionTokens)
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				outputChars += len(choice.Delta.Content)
				ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventDelta, Delta: choice.Delta.Content}
			}
			if choice.FinishReason != "" {
				emitDone("", "")
			}
		}
	}
	if ctx.Err() != nil {
		emitDone(sharederrors.CodeAiProviderTimeout, "context_cancelled")
		return
	}
	if err := stream.Err(); err != nil {
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: sdkStreamErrorCode(err)}
		return
	}
	emitDone("", "")
}

func sdkStreamErrorCode(err error) string {
	if errors.Is(err, responsebody.ErrInvalid) {
		return sharederrors.CodeAiOutputInvalid
	}
	if errors.Is(err, responsebody.ErrRead) {
		return sharederrors.CodeAiProviderTimeout
	}
	var streamError *ssestream.StreamError
	if errors.As(err, &streamError) {
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(streamError.Event.Data, &envelope) == nil && envelope.Error.Code != "" {
			return sharedOrOutputInvalid(envelope.Error.Code)
		}
		return sharederrors.CodeAiOutputInvalid
	}
	return sharederrors.CodeAiOutputInvalid
}

// Synthesize implements aiclient.Provider. The openai_compatible protocol does
// not support TTS synthesis; speech adapters use doubao_speech / minimax_speech.
func (a *Adapter) Synthesize(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, a.errMeta(profile, nil, 0, sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "openai_compatible protocol does not support TTS synthesis", false)), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "openai_compatible protocol does not support TTS synthesis", false)
}

func sdkCompleteRequest(profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (openai.ChatCompletionNewParams, []option.RequestOption, error) {
	request := openai.ChatCompletionNewParams{
		Model:    profile.Default.Model,
		Messages: convertSDKMessages(payload.Messages),
	}
	if profile.MaxTokens > 0 {
		request.MaxTokens = openai.Int(int64(profile.MaxTokens))
	}
	if value, ok := profile.Default.Params["temperature"]; ok {
		if number, valid := toFloat(value); valid {
			request.Temperature = openai.Float(number)
		}
	}
	if value, ok := profile.Default.Params["top_p"]; ok {
		if number, valid := toFloat(value); valid {
			request.TopP = openai.Float(number)
		}
	}

	tools, err := convertSDKTools(payload.Tools)
	if err != nil {
		return openai.ChatCompletionNewParams{}, nil, err
	}
	request.Tools = tools
	request.ToolChoice = convertSDKToolChoice(payload.ToolChoice)
	if responseFormatForPayload(payload) != nil {
		format := shared.NewResponseFormatJSONObjectParam()
		request.ResponseFormat.OfJSONObject = &format
	}

	var requestOptions []option.RequestOption
	if raw, ok := profile.Default.Params["thinking"]; ok {
		mode, valid := raw.(string)
		if !valid || (mode != "enabled" && mode != "disabled") {
			return openai.ChatCompletionNewParams{}, nil, errors.New("thinking must be enabled or disabled")
		}
		requestOptions = append(requestOptions, option.WithJSONSet("thinking", map[string]string{"type": mode}))
	}
	return request, requestOptions, nil
}

func convertSDKMessages(messages []aiclient.Message) []openai.ChatCompletionMessageParamUnion {
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

func convertSDKTools(tools []aiclient.Tool) ([]openai.ChatCompletionToolUnionParam, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		definition := shared.FunctionDefinitionParam{Name: tool.Name}
		if tool.Description != "" {
			definition.Description = openai.String(tool.Description)
		}
		if len(tool.Parameters) > 0 {
			var parameters map[string]any
			if err := json.Unmarshal(tool.Parameters, &parameters); err != nil {
				return nil, err
			}
			definition.Parameters = parameters
		}
		out = append(out, openai.ChatCompletionFunctionTool(definition))
	}
	return out, nil
}

func convertSDKToolChoice(choice *aiclient.ToolChoice) openai.ChatCompletionToolChoiceOptionUnionParam {
	if choice == nil {
		return openai.ChatCompletionToolChoiceOptionUnionParam{}
	}
	switch choice.Mode {
	case aiclient.ToolChoiceModeAuto:
		return openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("auto")}
	case aiclient.ToolChoiceModeNone:
		return openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("none")}
	case aiclient.ToolChoiceModeTool:
		return openai.ToolChoiceOptionFunctionToolChoice(openai.ChatCompletionNamedToolChoiceFunctionParam{Name: choice.Name})
	default:
		return openai.ChatCompletionToolChoiceOptionUnionParam{}
	}
}

func convertSDKToolCalls(calls []openai.ChatCompletionMessageToolCallUnion) []aiclient.ToolCall {
	if len(calls) == 0 {
		return nil
	}
	out := make([]aiclient.ToolCall, 0, len(calls))
	for _, call := range calls {
		if call.Type != "" && call.Type != "function" {
			continue
		}
		arguments := json.RawMessage(call.Function.Arguments)
		if len(arguments) == 0 {
			arguments = nil
		}
		out = append(out, aiclient.ToolCall{ID: call.ID, Name: call.Function.Name, Arguments: arguments})
	}
	return out
}

func mapSDKError(ctx context.Context, response *http.Response, err error) error {
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
	if errors.As(err, &apiError) {
		if apiError.StatusCode >= http.StatusInternalServerError {
			return stableError(sharederrors.CodeAiProviderTimeout)
		}
		if metadata, ok := sharederrors.CodeRegistry[apiError.Code]; ok {
			return sharederrors.Wrap(apiError.Code, metadata.Message, metadata.Retryable)
		}
		return stableError(sharederrors.CodeAiOutputInvalid)
	}
	if response == nil {
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	return stableError(sharederrors.CodeAiOutputInvalid)
}

func responseHeaders(response *http.Response) http.Header {
	if response == nil {
		return nil
	}
	return response.Header
}

func (a *Adapter) errMeta(profile *aiclient.ModelProfile, headers http.Header, latencyMs int64, err error) aiclient.AICallMeta {
	meta := aiclient.AICallMeta{
		Provider:         a.providerRef,
		ModelID:          profile.Default.Model,
		LatencyMs:        latencyMs,
		ValidationStatus: aiclient.ValidationStatusInvalid,
		ErrorCode:        errorCodeOf(err),
	}
	mergeFallbackHeaders(profile, headers, &meta)
	return meta
}

func (a *Adapter) buildMeta(profile *aiclient.ModelProfile, headers http.Header, latencyMs int64, inputTokens, outputTokens int, toolCalls []aiclient.ToolCall) aiclient.AICallMeta {
	modelID := profile.Default.Model
	meta := aiclient.AICallMeta{
		Provider:        a.providerRef,
		ModelFamily:     modelFamily(modelID),
		ModelID:         modelID,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		LatencyMs:       latencyMs,
		ToolInvocations: summarizeToolCalls(toolCalls),
	}
	mergeFallbackHeaders(profile, headers, &meta)
	return meta
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

func mergeFallbackHeaders(profile *aiclient.ModelProfile, _ http.Header, meta *aiclient.AICallMeta) {
	if profile != nil {
		meta.Route = profile.Route
	}
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

func errorCodeOf(err error) string {
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return ""
}

func sharedOrOutputInvalid(code string) string {
	if _, ok := sharederrors.CodeRegistry[code]; ok {
		return code
	}
	return sharederrors.CodeAiOutputInvalid
}

func modelFamily(modelID string) string {
	if modelID == "" {
		return ""
	}
	parts := strings.Split(modelID, "-")
	if len(parts) >= 4 && isDateSuffix(parts[len(parts)-3], parts[len(parts)-2], parts[len(parts)-1]) {
		return strings.Join(parts[:len(parts)-3], "-")
	}
	return modelID
}

func isDateSuffix(year, month, day string) bool {
	return len(year) == 4 && len(month) == 2 && len(day) == 2 &&
		allDigits(year) && allDigits(month) && allDigits(day)
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func responseFormatForPayload(payload aiclient.CompletePayload) any {
	if len(payload.Metadata.OutputSchema) == 0 {
		return nil
	}
	var schema struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload.Metadata.OutputSchema, &schema); err != nil {
		return nil
	}
	if schema.Type != "object" {
		return nil
	}
	return map[string]string{"type": "json_object"}
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	}
	return 0, false
}
