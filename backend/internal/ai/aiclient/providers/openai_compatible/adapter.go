package openaicompatible

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	platformconfig "github.com/monshunter/easyinterview/backend/internal/platform/config"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
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
	providerRef          string
	baseURL              string
	apiKey               string
	client               *http.Client
	maxResponseBodyBytes int64
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
		providerRef:          opts.Provider.Entry.Name,
		baseURL:              normalizeBaseURL(opts.Provider.BaseURL),
		apiKey:               opts.Provider.APIKey,
		client:               hc,
		maxResponseBodyBytes: opts.MaxResponseBodyBytes,
	}, nil
}

// Name implements aiclient.Provider.
func (a *Adapter) Name() string { return a.providerRef }

// Complete implements aiclient.Provider.
func (a *Adapter) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("openai_compatible: profile is nil")
	}

	req := chatCompletionsRequest{
		Model:          profile.Default.Model,
		Messages:       convertMessages(payload.Messages),
		Stream:         false,
		Tools:          convertTools(payload.Tools),
		ToolChoice:     convertToolChoice(payload.ToolChoice),
		ResponseFormat: responseFormatForPayload(payload),
	}
	if profile.MaxTokens > 0 {
		req.MaxTokens = profile.MaxTokens
	}
	if err := applyParams(profile.Default.Params, &req); err != nil {
		errCode := stableError(sharederrors.CodeAiProviderConfigInvalid)
		return aiclient.CompleteResponse{}, a.errMeta(profile, nil, 0, errCode), errCode
	}

	start := time.Now()
	resp, status, headers, err := a.postJSON(ctx, profile.TimeoutMs, PathChatCompletions, req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return aiclient.CompleteResponse{}, a.errMeta(profile, headers, latencyMs, err), err
	}

	if status >= 400 {
		errCode := mapHTTPError(status, resp)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}

	var body chatCompletionsResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		errCode := stableError(sharederrors.CodeAiOutputInvalid)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}
	if len(body.Choices) == 0 {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: response missing choices", false)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}

	choice := body.Choices[0]
	out := aiclient.CompleteResponse{
		Content:      choice.Message.Content,
		FinishReason: safeFinishReason(choice.FinishReason),
		ToolCalls:    convertToolCalls(choice.Message.ToolCalls),
	}
	meta := a.buildMeta(profile, headers, latencyMs, body.Usage.PromptTokens, body.Usage.CompletionTokens, out.ToolCalls)
	return out, meta, nil
}

// Transcribe implements aiclient.Provider using the OpenAI-compatible Audio
// Transcriptions multipart wire shape.
func (a *Adapter) Transcribe(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("openai_compatible: profile is nil")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writeTranscriptionForm(writer, profile.Default.Model, input); err != nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, err
	}
	if err := writer.Close(); err != nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, fmt.Errorf("openai_compatible: close transcription multipart: %w", err)
	}

	start := time.Now()
	resp, status, headers, err := a.postMultipart(ctx, profile.TimeoutMs, PathTranscriptions, writer.FormDataContentType(), &buf)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return aiclient.TranscriptionResponse{}, a.errMeta(profile, headers, latencyMs, err), err
	}

	if status >= 400 {
		errCode := mapHTTPError(status, resp)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.TranscriptionResponse{}, meta, errCode
	}

	var body transcriptionResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		errCode := stableError(sharederrors.CodeAiOutputInvalid)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.TranscriptionResponse{}, meta, errCode
	}
	if body.Text == "" {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: transcription response missing text", false)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.TranscriptionResponse{}, meta, errCode
	}

	out := aiclient.TranscriptionResponse{Text: body.Text}
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
	req := chatCompletionsRequest{
		Model:          profile.Default.Model,
		Messages:       convertMessages(payload.Messages),
		Stream:         true,
		Tools:          convertTools(payload.Tools),
		ToolChoice:     convertToolChoice(payload.ToolChoice),
		ResponseFormat: responseFormatForPayload(payload),
	}
	if profile.MaxTokens > 0 {
		req.MaxTokens = profile.MaxTokens
	}
	if err := applyParams(profile.Default.Params, &req); err != nil {
		return nil, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}

	streamCtx := ctx
	var cancel context.CancelFunc
	if profile.TimeoutMs > 0 {
		streamCtx, cancel = context.WithTimeout(ctx, time.Duration(profile.TimeoutMs)*time.Millisecond)
	}

	start := time.Now()
	resp, err := a.postStream(streamCtx, PathChatCompletions, req)
	latencyMs := time.Since(start).Milliseconds()
	ch := make(chan aiclient.AIStreamEvent, 4)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: errorCodeOf(err)}
		close(ch)
		return ch, nil
	}
	go a.consumeStream(streamCtx, cancel, resp, profile, latencyMs, ch)
	return ch, nil
}

// Synthesize implements aiclient.Provider. The openai_compatible protocol does
// not support TTS synthesis; speech adapters use doubao_speech / minimax_speech.
func (a *Adapter) Synthesize(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, a.errMeta(profile, nil, 0, sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "openai_compatible protocol does not support TTS synthesis", false)), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "openai_compatible protocol does not support TTS synthesis", false)
}

func (a *Adapter) postStream(ctx context.Context, path string, body any) (*http.Response, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai_compatible: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("openai_compatible: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if rid := aiclient.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, stableError(sharederrors.CodeAiProviderTimeout)
		}
		return nil, stableError(sharederrors.CodeAiProviderTimeout)
	}
	return resp, nil
}

func (a *Adapter) consumeStream(ctx context.Context, cancel context.CancelFunc, resp *http.Response, profile *aiclient.ModelProfile, latencyMs int64, ch chan<- aiclient.AIStreamEvent) {
	defer close(ch)
	defer resp.Body.Close()
	if cancel != nil {
		defer cancel()
	}

	if resp.StatusCode >= 400 {
		body, err := readResponseBody(resp, a.maxResponseBodyBytes)
		if err != nil {
			ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: errorCodeOf(err)}
			return
		}
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: errorCodeOf(mapHTTPError(resp.StatusCode, body))}
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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
		meta := a.buildMeta(profile, resp.Header, latencyMs, inputTokens, outputTokens, nil)
		if errorCode != "" {
			meta.ValidationStatus = aiclient.ValidationStatusInvalid
			meta.ErrorCode = errorCode
			meta.PartialMetaReason = partialReason
		}
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventDone, Meta: &meta}
		emittedDone = true
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			emitDone("", "")
			continue
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: sharederrors.CodeAiOutputInvalid}
			return
		}
		if chunk.Error.Code != "" {
			ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: sharedOrOutputInvalid(chunk.Error.Code)}
			return
		}
		if chunk.Usage.PromptTokens > 0 {
			inputTokens = chunk.Usage.PromptTokens
		}
		if chunk.Usage.CompletionTokens > 0 {
			outputTokens = chunk.Usage.CompletionTokens
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
	if err := scanner.Err(); err != nil {
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: sharederrors.CodeAiProviderTimeout}
		return
	}
	emitDone("", "")
}

func (a *Adapter) postJSON(ctx context.Context, timeoutMs int, path string, body any) ([]byte, int, http.Header, error) {
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, 0, nil, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, 0, nil, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if rid := aiclient.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, 0, nil, stableError(sharederrors.CodeAiProviderTimeout)
		}
		return nil, 0, nil, stableError(sharederrors.CodeAiProviderTimeout)
	}
	defer resp.Body.Close()
	respBody, err := readResponseBody(resp, a.maxResponseBodyBytes)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, err
	}
	return respBody, resp.StatusCode, resp.Header, nil
}

func (a *Adapter) postMultipart(ctx context.Context, timeoutMs int, path, contentType string, body io.Reader) ([]byte, int, http.Header, error) {
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, body)
	if err != nil {
		return nil, 0, nil, stableError(sharederrors.CodeAiProviderConfigInvalid)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if rid := aiclient.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, 0, nil, stableError(sharederrors.CodeAiProviderTimeout)
		}
		return nil, 0, nil, stableError(sharederrors.CodeAiProviderTimeout)
	}
	defer resp.Body.Close()
	respBody, err := readResponseBody(resp, a.maxResponseBodyBytes)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, err
	}
	return respBody, resp.StatusCode, resp.Header, nil
}

func mapHTTPError(status int, body []byte) error {
	if status >= 500 {
		return stableError(sharederrors.CodeAiProviderTimeout)
	}
	// 4xx: accept only a registered stable error code; provider prose is never
	// returned or persisted. Unknown envelopes fail closed as AI_OUTPUT_INVALID.
	var env errorEnvelope
	if json.Unmarshal(body, &env) == nil && env.Error.Code != "" {
		if meta, ok := sharederrors.CodeRegistry[env.Error.Code]; ok {
			return sharederrors.Wrap(env.Error.Code, meta.Message, meta.Retryable)
		}
	}
	return stableError(sharederrors.CodeAiOutputInvalid)
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

func readResponseBody(resp *http.Response, maxResponseBodyBytes int64) ([]byte, error) {
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

func convertMessages(in []aiclient.Message) []wireMessage {
	out := make([]wireMessage, len(in))
	for i, m := range in {
		out[i] = wireMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func convertTools(in []aiclient.Tool) []wireTool {
	if len(in) == 0 {
		return nil
	}
	out := make([]wireTool, len(in))
	for i, tool := range in {
		out[i] = wireTool{
			Type: "function",
			Function: wireFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		}
	}
	return out
}

func convertToolChoice(choice *aiclient.ToolChoice) any {
	if choice == nil {
		return nil
	}
	switch choice.Mode {
	case aiclient.ToolChoiceModeAuto:
		return "auto"
	case aiclient.ToolChoiceModeNone:
		return "none"
	case aiclient.ToolChoiceModeTool:
		return map[string]any{
			"type": "function",
			"function": map[string]string{
				"name": choice.Name,
			},
		}
	default:
		return nil
	}
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

func convertToolCalls(in []wireToolCall) []aiclient.ToolCall {
	if len(in) == 0 {
		return nil
	}
	out := make([]aiclient.ToolCall, 0, len(in))
	for _, call := range in {
		if call.Type != "" && call.Type != "function" {
			continue
		}
		args := json.RawMessage(call.Function.Arguments)
		if len(args) == 0 {
			args = nil
		}
		out = append(out, aiclient.ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: args,
		})
	}
	return out
}

func writeTranscriptionForm(writer *multipart.Writer, model string, input aiclient.TranscriptionInput) error {
	if err := writer.WriteField("model", model); err != nil {
		return fmt.Errorf("openai_compatible: write transcription model: %w", err)
	}
	if input.Language != "" {
		if err := writer.WriteField("language", input.Language); err != nil {
			return fmt.Errorf("openai_compatible: write transcription language: %w", err)
		}
	}
	if input.Prompt != "" {
		if err := writer.WriteField("prompt", input.Prompt); err != nil {
			return fmt.Errorf("openai_compatible: write transcription prompt: %w", err)
		}
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeMultipartValue(input.Filename)))
	header.Set("Content-Type", input.ContentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("openai_compatible: create transcription file part: %w", err)
	}
	if _, err := part.Write(input.Audio); err != nil {
		return fmt.Errorf("openai_compatible: write transcription audio: %w", err)
	}
	return nil
}

func escapeMultipartValue(s string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(s)
}

func applyParams(params map[string]any, req *chatCompletionsRequest) error {
	if params == nil {
		return nil
	}
	if v, ok := params["temperature"]; ok {
		if f, ok := toFloat(v); ok {
			req.Temperature = &f
		}
	}
	if v, ok := params["top_p"]; ok {
		if f, ok := toFloat(v); ok {
			req.TopP = &f
		}
	}
	if raw, ok := params["thinking"]; ok {
		mode, ok := raw.(string)
		if !ok || (mode != "enabled" && mode != "disabled") {
			return errors.New("thinking must be enabled or disabled")
		}
		req.Thinking = &thinkingMode{Type: mode}
	}
	return nil
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

func normalizeBaseURL(raw string) string {
	base := strings.TrimRight(raw, "/")
	return strings.TrimSuffix(base, "/v1")
}
