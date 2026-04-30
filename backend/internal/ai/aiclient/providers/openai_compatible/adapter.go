package openaicompatible

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
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// Name is the canonical provider name. Profiles route to this adapter by
// setting Default.Provider to "openai_compatible".
const Name = "openai_compatible"

// Path constants are exported so the mockserver helper can reuse them.
const (
	PathChatCompletions = "/v1/chat/completions"
	PathEmbeddings      = "/v1/embeddings"
)

// Header names used for fallback/route metadata. The gateway populates
// these on responses; A3 client only reads them.
const (
	HeaderRequestID    = "X-Request-ID"
	HeaderFallbackFrom = "X-Fallback-From"
	HeaderFallbackTo   = "X-Fallback-To"
	HeaderRoute        = "X-Route"
)

// Options configures the adapter.
type Options struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// Adapter is the concrete aiclient.Provider implementation.
type Adapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// New constructs an Adapter. BaseURL and APIKey are required.
func New(opts Options) (*Adapter, error) {
	if opts.BaseURL == "" {
		return nil, errors.New("openai_compatible: BaseURL is required")
	}
	if opts.APIKey == "" {
		return nil, errors.New("openai_compatible: APIKey is required")
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{}
	}
	return &Adapter{
		baseURL: normalizeBaseURL(opts.BaseURL),
		apiKey:  opts.APIKey,
		client:  hc,
	}, nil
}

// Name implements aiclient.Provider.
func (a *Adapter) Name() string { return Name }

// Complete implements aiclient.Provider.
func (a *Adapter) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("openai_compatible: profile is nil")
	}

	req := chatCompletionsRequest{
		Model:    profile.Default.Model,
		Messages: convertMessages(payload.Messages),
		Stream:   false,
	}
	if profile.MaxTokens > 0 {
		req.MaxTokens = profile.MaxTokens
	}
	applyParams(profile.Default.Params, &req)

	start := time.Now()
	resp, status, headers, err := a.postJSON(ctx, profile.TimeoutMs, PathChatCompletions, req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return aiclient.CompleteResponse{}, errMeta(profile, headers, latencyMs, err), err
	}

	if status >= 400 {
		errCode := mapHTTPError(status, resp)
		meta := errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}

	var body chatCompletionsResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: parse response: "+err.Error(), false)
		meta := errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}
	if len(body.Choices) == 0 {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: response missing choices", false)
		meta := errMeta(profile, headers, latencyMs, errCode)
		return aiclient.CompleteResponse{}, meta, errCode
	}

	choice := body.Choices[0]
	out := aiclient.CompleteResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
	}
	meta := buildMeta(profile, headers, latencyMs, body.Model, body.Usage.PromptTokens, body.Usage.CompletionTokens)
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
		return aiclient.EmbedResponse{}, errMeta(profile, headers, latencyMs, err), err
	}

	if status >= 400 {
		errCode := mapHTTPError(status, resp)
		meta := errMeta(profile, headers, latencyMs, errCode)
		return aiclient.EmbedResponse{}, meta, errCode
	}

	var body embeddingsResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		errCode := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "openai_compatible: parse response: "+err.Error(), false)
		meta := errMeta(profile, headers, latencyMs, errCode)
		return aiclient.EmbedResponse{}, meta, errCode
	}

	vectors := make([][]float64, len(body.Data))
	for i, d := range body.Data {
		vectors[i] = d.Embedding
	}
	out := aiclient.EmbedResponse{Vectors: vectors}
	meta := buildMeta(profile, headers, latencyMs, body.Model, body.Usage.PromptTokens, 0)
	meta.OutputTokens = len(vectors)
	return out, meta, nil
}

// Stream implements aiclient.Provider. Plan 001 honors the channel contract
// by issuing the equivalent non-streaming request and emitting a single
// `done` event followed by close. Plan 002 replaces this body with a real
// SSE/chunked consumer.
func (a *Adapter) Stream(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	resp, meta, err := a.Complete(ctx, profile, payload)
	ch := make(chan aiclient.AIStreamEvent, 1)
	if err != nil {
		ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventError, ErrorCode: errorCodeOf(err)}
		close(ch)
		return ch, nil
	}
	final := meta
	final.OutputTokens = len(resp.Content)
	ch <- aiclient.AIStreamEvent{Type: aiclient.StreamEventDone, Meta: &final}
	close(ch)
	return ch, nil
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

func errMeta(profile *aiclient.ModelProfile, headers http.Header, latencyMs int64, err error) aiclient.AICallMeta {
	meta := aiclient.AICallMeta{
		Provider:         Name,
		ModelID:          profile.Default.Model,
		LatencyMs:        latencyMs,
		ValidationStatus: aiclient.ValidationStatusInvalid,
		ErrorCode:        errorCodeOf(err),
	}
	mergeFallbackHeaders(profile, headers, &meta)
	return meta
}

func buildMeta(profile *aiclient.ModelProfile, headers http.Header, latencyMs int64, modelID string, inputTokens, outputTokens int) aiclient.AICallMeta {
	if modelID == "" {
		modelID = profile.Default.Model
	}
	meta := aiclient.AICallMeta{
		Provider:     Name,
		ModelFamily:  modelFamily(modelID),
		ModelID:      modelID,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		LatencyMs:    latencyMs,
	}
	mergeFallbackHeaders(profile, headers, &meta)
	return meta
}

func mergeFallbackHeaders(profile *aiclient.ModelProfile, headers http.Header, meta *aiclient.AICallMeta) {
	if headers == nil {
		return
	}
	if route := headers.Get(HeaderRoute); route != "" {
		meta.Route = route
	} else if profile != nil && profile.GatewayRoute != "" {
		meta.Route = profile.GatewayRoute
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

func modelFamily(modelID string) string {
	if modelID == "" {
		return ""
	}
	// Strip everything after the last "-": "gpt-4-turbo-2024-04-09" → "gpt-4-turbo".
	if i := strings.LastIndex(modelID, "-"); i > 0 {
		return modelID[:i]
	}
	return modelID
}

func convertMessages(in []aiclient.Message) []wireMessage {
	out := make([]wireMessage, len(in))
	for i, m := range in {
		out[i] = wireMessage{Role: m.Role, Content: m.Content}
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
