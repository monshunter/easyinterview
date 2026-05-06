package openaicompatible

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// Name is the adapter protocol name. Concrete Adapter instances identify
// themselves by Provider Registry ref so profiles route through provider_ref.
const Name = "openai_compatible"

// Path constants are exported so the mockserver helper can reuse them.
const (
	PathChatCompletions = "/v1/chat/completions"
	PathEmbeddings      = "/v1/embeddings"
)

// Header names used for fallback/route metadata. The provider endpoint
// populates these on responses; A3 client only reads them.
const (
	HeaderRequestID    = "X-Request-ID"
	HeaderFallbackFrom = "X-Fallback-From"
	HeaderFallbackTo   = "X-Fallback-To"
	HeaderRoute        = "X-Route"
)

// Options configures the adapter.
type Options struct {
	Provider   providerregistry.ResolvedProvider
	HTTPClient *http.Client
}

// Adapter is the concrete aiclient.Provider implementation.
type Adapter struct {
	providerRef string
	baseURL     string
	apiKey      string
	client      *http.Client
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
	return &Adapter{
		providerRef: opts.Provider.Entry.Name,
		baseURL:     normalizeBaseURL(opts.Provider.BaseURL),
		apiKey:      opts.Provider.APIKey,
		client:      hc,
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
		Model:      profile.Default.Model,
		Messages:   convertMessages(payload.Messages),
		Stream:     false,
		Tools:      convertTools(payload.Tools),
		ToolChoice: convertToolChoice(payload.ToolChoice),
	}
	if profile.MaxTokens > 0 {
		req.MaxTokens = profile.MaxTokens
	}
	applyParams(profile.Default.Params, &req)

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
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: parse response: "+err.Error(), false)
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
		FinishReason: choice.FinishReason,
		ToolCalls:    convertToolCalls(choice.Message.ToolCalls),
	}
	meta := a.buildMeta(profile, headers, latencyMs, body.Model, body.Usage.PromptTokens, body.Usage.CompletionTokens, out.ToolCalls)
	return out, meta, nil
}

// Embed implements aiclient.Provider.
func (a *Adapter) Embed(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.EmbedInput) (aiclient.EmbedResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.EmbedResponse{}, aiclient.AICallMeta{}, errors.New("openai_compatible: profile is nil")
	}

	req := embeddingsRequest{
		Model: profile.Default.Model,
		Input: input.Texts,
	}

	start := time.Now()
	resp, status, headers, err := a.postJSON(ctx, profile.TimeoutMs, PathEmbeddings, req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return aiclient.EmbedResponse{}, a.errMeta(profile, headers, latencyMs, err), err
	}

	if status >= 400 {
		errCode := mapHTTPError(status, resp)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.EmbedResponse{}, meta, errCode
	}

	var body embeddingsResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: parse response: "+err.Error(), false)
		meta := a.errMeta(profile, headers, latencyMs, errCode)
		return aiclient.EmbedResponse{}, meta, errCode
	}

	vectors := make([][]float64, len(body.Data))
	for i, d := range body.Data {
		vectors[i] = d.Embedding
	}
	out := aiclient.EmbedResponse{Vectors: vectors}
	meta := a.buildMeta(profile, headers, latencyMs, body.Model, body.Usage.PromptTokens, 0, nil)
	meta.OutputTokens = len(vectors)
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
		Model:      profile.Default.Model,
		Messages:   convertMessages(payload.Messages),
		Stream:     true,
		Tools:      convertTools(payload.Tools),
		ToolChoice: convertToolChoice(payload.ToolChoice),
	}
	if profile.MaxTokens > 0 {
		req.MaxTokens = profile.MaxTokens
	}
	applyParams(profile.Default.Params, &req)

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
			return nil, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "openai_compatible: timeout", true)
		}
		return nil, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "openai_compatible: transport error: "+err.Error(), true)
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
		body, _ := io.ReadAll(resp.Body)
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: errorCodeOf(mapHTTPError(resp.StatusCode, body))}
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	modelID := profile.Default.Model
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
		meta := a.buildMeta(profile, resp.Header, latencyMs, modelID, inputTokens, outputTokens, nil)
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
		if chunk.Model != "" {
			modelID = chunk.Model
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
		return nil, 0, nil, fmt.Errorf("openai_compatible: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("openai_compatible: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if rid := aiclient.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, 0, nil, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "openai_compatible: timeout", true)
		}
		return nil, 0, nil, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "openai_compatible: transport error: "+err.Error(), true)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "openai_compatible: read response: "+err.Error(), true)
	}
	return respBody, resp.StatusCode, resp.Header, nil
}

func mapHTTPError(status int, body []byte) error {
	if status >= 500 {
		return sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, fmt.Sprintf("openai_compatible: upstream %d", status), true)
	}
	// 4xx: try to parse body for an error_code field; otherwise fall back to
	// AI_OUTPUT_INVALID for shape errors and a generic wire error otherwise.
	var env errorEnvelope
	if json.Unmarshal(body, &env) == nil && env.Error.Code != "" {
		if meta, ok := sharederrors.CodeRegistry[env.Error.Code]; ok {
			msg := env.Error.Message
			if msg == "" {
				msg = meta.Message
			}
			return sharederrors.Wrap(env.Error.Code, msg, meta.Retryable)
		}
	}
	return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, fmt.Sprintf("openai_compatible: upstream %d", status), false)
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

func (a *Adapter) buildMeta(profile *aiclient.ModelProfile, headers http.Header, latencyMs int64, modelID string, inputTokens, outputTokens int, toolCalls []aiclient.ToolCall) aiclient.AICallMeta {
	if modelID == "" {
		modelID = profile.Default.Model
	}
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

func mergeFallbackHeaders(profile *aiclient.ModelProfile, headers http.Header, meta *aiclient.AICallMeta) {
	if headers == nil {
		return
	}
	if route := headers.Get(HeaderRoute); route != "" {
		meta.Route = route
	} else if profile != nil && profile.Route != "" {
		meta.Route = profile.Route
	}
	from := headers.Get(HeaderFallbackFrom)
	to := headers.Get(HeaderFallbackTo)
	if from != "" || to != "" {
		chain := []string{}
		if from != "" {
			chain = append(chain, from)
		}
		if to != "" {
			chain = append(chain, to)
		}
		meta.FallbackChain = chain
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

func applyParams(params map[string]any, req *chatCompletionsRequest) {
	if params == nil {
		return
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
	if strings.HasSuffix(base, "/v1") {
		base = strings.TrimSuffix(base, "/v1")
	}
	return base
}
