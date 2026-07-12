package openaicompatible_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible/mockserver"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const (
	chatModelID     = "chat-primary-2026-05-05"
	chatModelFamily = "chat-primary"
	sttModelID      = "stt-transcribe-1"
	sttModelFamily  = "stt-transcribe-1"
	providerRef     = "deepseek"
)

func chatProfile(timeoutMs int) *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:       "practice.chat.default",
		Capability: aiclient.CapabilityChat,
		Status:     aiclient.ProfileStatusActive,
		Default: aiclient.ProviderConfig{
			ProviderRef: providerRef,
			Model:       chatModelID,
		},
		TimeoutMs: timeoutMs,
		Route:     "practice.session.chat",
		Version:   "1.0.0",
	}
}

func sttProfile(timeoutMs int) *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:       "practice.voice.stt.default",
		Capability: aiclient.CapabilitySTT,
		Status:     aiclient.ProfileStatusActive,
		Default: aiclient.ProviderConfig{
			ProviderRef: providerRef,
			Model:       sttModelID,
		},
		TimeoutMs: timeoutMs,
		Route:     "practice.voice.stt",
		Version:   "1.0.0",
	}
}

func samplePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "you are an interviewer."},
			{Role: "user", Content: "Tell me about a project."},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.session.chat",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	}
}

func newAdapter(t *testing.T, srv *mockserver.Server) *openaicompatible.Adapter {
	t.Helper()
	a, err := openaicompatible.New(openaicompatible.Options{
		Provider: resolvedProvider(srv.URL()),
	})
	if err != nil {
		t.Fatalf("openai_compatible.New: %v", err)
	}
	return a
}

func resolvedProvider(baseURL string) providerregistry.ResolvedProvider {
	return providerregistry.ResolvedProvider{
		Entry: aiclient.ProviderRegistryEntry{
			Name:     providerRef,
			Protocol: aiclient.ProviderProtocolOpenAICompatible,
			Capabilities: []aiclient.Capability{
				aiclient.CapabilityChat,
				aiclient.CapabilitySTT,
			},
			Version: "1.0.0",
		},
		BaseURL: baseURL,
		APIKey:  "test-key",
	}
}

func TestComplete_NormalChatCompletion(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	resp, meta, err := a.Complete(aiclient.WithRequestID(context.Background(), "req-1"), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !strings.HasPrefix(resp.Content, "mock response for ") {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if meta.Provider != providerRef {
		t.Fatalf("expected meta.Provider=%q, got %q", providerRef, meta.Provider)
	}
	if meta.ModelID != chatModelID {
		t.Fatalf("expected ModelID=%q, got %q", chatModelID, meta.ModelID)
	}
	if meta.ModelFamily != chatModelFamily {
		t.Fatalf("expected ModelFamily=%q, got %q", chatModelFamily, meta.ModelFamily)
	}
	if meta.InputTokens == 0 {
		t.Fatalf("expected non-zero input tokens, got %+v", meta)
	}
	if meta.OutputTokens == 0 {
		t.Fatalf("expected non-zero output tokens, got %+v", meta)
	}
	if meta.LatencyMs < 0 {
		t.Fatalf("expected non-negative latency, got %d", meta.LatencyMs)
	}

	requests := srv.Captured()
	if len(requests) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(requests))
	}
	got := requests[0]
	if got.Path != "/v1/chat/completions" {
		t.Fatalf("expected /v1/chat/completions, got %q", got.Path)
	}
	if got.Authorization != "Bearer test-key" {
		t.Fatalf("expected Bearer auth, got %q", got.Authorization)
	}
	if got.ContentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got.ContentType)
	}
	if got.RequestID != "req-1" {
		t.Fatalf("expected X-Request-ID propagation, got %q", got.RequestID)
	}
}

func TestComplete_BaseURLMayIncludeV1Prefix(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a, err := openaicompatible.New(openaicompatible.Options{
		Provider: resolvedProvider(srv.URL() + "/v1"),
	})
	if err != nil {
		t.Fatalf("openai_compatible.New: %v", err)
	}

	_, _, err = a.Complete(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	requests := srv.Captured()
	if len(requests) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(requests))
	}
	if got := requests[0].Path; got != "/v1/chat/completions" {
		t.Fatalf("expected normalized /v1/chat/completions path, got %q", got)
	}
}

func TestTranscribe_PostsMultipartAudioAndReturnsTranscript(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	input := aiclient.TranscriptionInput{
		Audio:       []byte("fake-audio-bytes"),
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Language:    "en",
		Prompt:      "interview answer",
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.voice.stt",
			PromptVersion: "stt-p1",
			Language:      "en",
		},
	}
	resp, meta, err := a.Transcribe(aiclient.WithRequestID(context.Background(), "req-stt-1"), sttProfile(5000), input)
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	if resp.Text != "mock transcript for answer.webm" {
		t.Fatalf("unexpected transcript: %q", resp.Text)
	}
	if meta.Provider != providerRef || meta.ModelID != sttModelID || meta.ModelFamily != sttModelFamily {
		t.Fatalf("meta mismatch: %+v", meta)
	}
	if meta.InputTokens != len(input.Audio) || meta.OutputTokens != len(resp.Text) {
		t.Fatalf("usage meta mismatch: %+v", meta)
	}

	requests := srv.Captured()
	if len(requests) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(requests))
	}
	got := requests[0]
	if got.Path != "/v1/audio/transcriptions" {
		t.Fatalf("expected /v1/audio/transcriptions, got %q", got.Path)
	}
	if got.Authorization != "Bearer test-key" || got.RequestID != "req-stt-1" {
		t.Fatalf("headers not propagated: %+v", got)
	}
	form := parseMultipartRequest(t, got.ContentType, got.Body)
	if form.values["model"] != sttModelID || form.values["language"] != "en" || form.values["prompt"] != "interview answer" {
		t.Fatalf("multipart fields mismatch: %+v", form.values)
	}
	if form.fileName != "answer.webm" || form.fileContentType != "audio/webm" || string(form.fileBytes) != string(input.Audio) {
		t.Fatalf("multipart file mismatch: %+v", form)
	}
}

func TestTranscribe_ProviderErrorReturnsSharedCode(t *testing.T) {
	srv := mockserver.New()
	srv.SetTranscriptionBehavior(mockserver.Behavior{StatusCode: 503, ErrorBody: `{"error":{"code":"AI_PROVIDER_TIMEOUT"}}`})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Transcribe(context.Background(), sttProfile(5000), aiclient.TranscriptionInput{
		Audio:       []byte("fake"),
		Filename:    "answer.webm",
		ContentType: "audio/webm",
	})
	if err == nil {
		t.Fatalf("expected provider error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("expected AI_PROVIDER_TIMEOUT, got %v", err)
	}
	if meta.ErrorCode != sharederrors.CodeAiProviderTimeout || meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected timeout meta, got %+v", meta)
	}
}

func TestTranscribe_MissingTextReturnsAIOutputInvalid(t *testing.T) {
	srv := mockserver.New()
	srv.SetTranscriptionBodyOverride(func() string {
		return `{"duration":1.2}`
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Transcribe(context.Background(), sttProfile(5000), aiclient.TranscriptionInput{
		Audio:       []byte("fake"),
		Filename:    "answer.webm",
		ContentType: "audio/webm",
	})
	if err == nil {
		t.Fatalf("expected output invalid error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
}

func TestComplete_MapsToolsAndParsesToolCalls(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBodyOverride(func() string {
		return `{
			"id":"mock-id-tools",
			"model":"` + chatModelID + `",
			"choices":[{
				"index":0,
				"message":{
					"role":"assistant",
					"content":"",
					"tool_calls":[{
						"id":"call_1",
						"type":"function",
						"function":{
							"name":"extract_signal",
							"arguments":"{\"signal\":\"scope\"}"
						}
					}]
				},
				"finish_reason":"tool_calls"
			}],
			"usage":{"prompt_tokens":11,"completion_tokens":3,"total_tokens":14}
		}`
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	payload := samplePayload()
	payload.Tools = []aiclient.Tool{{
		Name:        "extract_signal",
		Description: "Extract structured signal.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"signal":{"type":"string"}}}`),
	}}
	payload.ToolChoice = &aiclient.ToolChoice{Mode: aiclient.ToolChoiceModeTool, Name: "extract_signal"}

	resp, meta, err := a.Complete(context.Background(), chatProfile(5000), payload)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.FinishReason != "tool_calls" {
		t.Fatalf("expected finish_reason tool_calls, got %q", resp.FinishReason)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected one tool call, got %+v", resp.ToolCalls)
	}
	if resp.ToolCalls[0].Name != "extract_signal" {
		t.Fatalf("tool call not parsed: %+v", resp.ToolCalls[0])
	}
	if string(resp.ToolCalls[0].Arguments) != `{"signal":"scope"}` {
		t.Fatalf("tool arguments mismatch: %s", resp.ToolCalls[0].Arguments)
	}
	if len(meta.ToolInvocations) != 1 {
		t.Fatalf("expected one tool invocation summary, got %+v", meta.ToolInvocations)
	}
	if meta.ToolInvocations[0].Name != "extract_signal" {
		t.Fatalf("tool invocation name mismatch: %+v", meta.ToolInvocations[0])
	}
	if meta.ToolInvocations[0].ArgumentsHash == "" || meta.ToolInvocations[0].ArgumentsLength == 0 {
		t.Fatalf("tool invocation must include arguments hash and length: %+v", meta.ToolInvocations[0])
	}
	if strings.Contains(meta.ToolInvocations[0].ArgumentsHash, "scope") {
		t.Fatalf("tool invocation hash leaked raw arguments: %+v", meta.ToolInvocations[0])
	}

	var wire map[string]any
	if err := json.Unmarshal(srv.Captured()[0].Body, &wire); err != nil {
		t.Fatalf("unmarshal captured request: %v", err)
	}
	if _, ok := wire["tools"].([]any); !ok {
		t.Fatalf("expected request tools array, got %s", srv.Captured()[0].Body)
	}
	choice, ok := wire["tool_choice"].(map[string]any)
	if !ok || choice["type"] != "function" {
		t.Fatalf("expected function tool_choice, got %+v", wire["tool_choice"])
	}
}

func TestComplete_RequestsJSONObjectWhenOutputSchemaIsPresent(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	payload := samplePayload()
	payload.Metadata.OutputSchema = json.RawMessage(`{"type":"object","required":["basics"],"properties":{"basics":{"type":"object"}}}`)

	if _, _, err := a.Complete(context.Background(), chatProfile(5000), payload); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	var wire map[string]any
	if err := json.Unmarshal(srv.Captured()[0].Body, &wire); err != nil {
		t.Fatalf("unmarshal captured request: %v", err)
	}
	responseFormat, ok := wire["response_format"].(map[string]any)
	if !ok {
		t.Fatalf("response_format missing from request: %+v", wire)
	}
	if responseFormat["type"] != "json_object" {
		t.Fatalf("response_format.type = %v", responseFormat["type"])
	}
}

func TestComplete_LeavesResponseFormatUnsetForArrayOutputSchema(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	payload := samplePayload()
	payload.Metadata.OutputSchema = json.RawMessage(`{"type":"array","items":{"type":"object","required":["jobMatchId"],"properties":{"jobMatchId":{"type":"string"}}}}`)

	if _, _, err := a.Complete(context.Background(), chatProfile(5000), payload); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	var wire map[string]any
	if err := json.Unmarshal(srv.Captured()[0].Body, &wire); err != nil {
		t.Fatalf("unmarshal captured request: %v", err)
	}
	if _, ok := wire["response_format"]; ok {
		t.Fatalf("array output schema must not force object response_format: %+v", wire["response_format"])
	}
}

func TestComplete_TimeoutMapsToAIProviderTimeout(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{SleepBeforeRespond: 200 * time.Millisecond})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(50), samplePayload())
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("expected %q, got %q", sharederrors.CodeAiProviderTimeout, apiErr.Code)
	}
	if meta.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("meta.ErrorCode mismatch: %q", meta.ErrorCode)
	}
}

func TestComplete_5xxMapsToAIProviderTimeout(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{StatusCode: 503, ErrorBody: "upstream gone"})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected error for 5xx")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("expected AI_PROVIDER_TIMEOUT, got %v", err)
	}
}

func TestComplete_4xxParsesErrorEnvelope(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"AI_OUTPUT_INVALID","message":"bad input"}}`,
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected error for 4xx")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected upstream code passthrough %q, got %q", sharederrors.CodeAiOutputInvalid, apiErr.Code)
	}
}

func TestComplete_4xxUnknownErrorCodeFallsBackToAIOutputInvalid(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{
		StatusCode: 400,
		ErrorBody:  `{"error":{"code":"PROVIDER_PRIVATE_BAD_REQUEST","message":"bad input"}}`,
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected error for 4xx")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected unknown upstream code to fall back to %q, got %q", sharederrors.CodeAiOutputInvalid, apiErr.Code)
	}
	if meta.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected meta.ErrorCode=%q, got %q", sharederrors.CodeAiOutputInvalid, meta.ErrorCode)
	}
}

func TestComplete_FallbackHeadersPopulateMeta(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{
		FallbackFrom: "primary/chat",
		FallbackTo:   "fallback/chat",
		Route:        "practice.session.chat",
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if len(meta.FallbackChain) != 2 || meta.FallbackChain[0] != "primary/chat" || meta.FallbackChain[1] != "fallback/chat" {
		t.Fatalf("fallback chain not populated: %+v", meta.FallbackChain)
	}
	if meta.Route != "practice.session.chat" {
		t.Fatalf("route not populated: %q", meta.Route)
	}
}

func TestComplete_MissingChoicesReturnsAIOutputInvalid(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatBehavior(mockserver.Behavior{MissingChoices: true})
	defer srv.Close()
	a := newAdapter(t, srv)

	_, _, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err == nil {
		t.Fatalf("expected AI_OUTPUT_INVALID for missing choices")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
}

func TestComplete_NoFallbackHeadersUsesProfileRoute(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()
	a := newAdapter(t, srv)

	_, meta, err := a.Complete(context.Background(), chatProfile(2000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if meta.Route != "practice.session.chat" {
		t.Fatalf("expected fallback route from profile, got %q", meta.Route)
	}
	if len(meta.FallbackChain) != 0 {
		t.Fatalf("expected empty fallback chain when headers absent, got %+v", meta.FallbackChain)
	}
}

func TestStream_ParsesSSEDeltaAndDone(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatStreamChunks([]string{
		`{"model":"` + chatModelID + `","choices":[{"delta":{"content":"hello "}}]}`,
		`{"model":"` + chatModelID + `","choices":[{"delta":{"content":"world"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}`,
		`[DONE]`,
	})
	defer srv.Close()
	a := newAdapter(t, srv)

	ch, err := a.Stream(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	events := collectStreamEvents(t, ch)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %+v", events)
	}
	if events[0].Type != aiclient.StreamEventDelta || events[0].Delta != "hello " {
		t.Fatalf("first delta mismatch: %+v", events[0])
	}
	if events[1].Type != aiclient.StreamEventDelta || events[1].Delta != "world" {
		t.Fatalf("second delta mismatch: %+v", events[1])
	}
	if events[2].Type != aiclient.StreamEventDone || events[2].Meta == nil {
		t.Fatalf("expected done with meta, got %+v", events[2])
	}
	if events[2].Meta.Provider != providerRef || events[2].Meta.ModelID != chatModelID {
		t.Fatalf("done meta mismatch: %+v", events[2].Meta)
	}
	if events[2].Meta.InputTokens != 5 || events[2].Meta.OutputTokens != 2 {
		t.Fatalf("usage meta mismatch: %+v", events[2].Meta)
	}
}

func TestStream_ErrorChunksEmitSharedError(t *testing.T) {
	tests := []struct {
		name      string
		chunk     string
		errorCode string
	}{
		{name: "malformed chunk", chunk: `{"choices":[`, errorCode: sharederrors.CodeAiOutputInvalid},
		{name: "provider error event", chunk: `{"error":{"code":"AI_PROVIDER_TIMEOUT","message":"upstream timeout"}}`, errorCode: sharederrors.CodeAiProviderTimeout},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := mockserver.New()
			srv.SetChatStreamChunks([]string{tc.chunk})
			defer srv.Close()
			a := newAdapter(t, srv)

			ch, err := a.Stream(context.Background(), chatProfile(5000), samplePayload())
			if err != nil {
				t.Fatalf("Stream: %v", err)
			}
			events := collectStreamEvents(t, ch)
			if len(events) != 1 || events[0].Type != aiclient.StreamEventError {
				t.Fatalf("expected one error event, got %+v", events)
			}
			if events[0].ErrorCode != tc.errorCode {
				t.Fatalf("expected %s, got %q", tc.errorCode, events[0].ErrorCode)
			}
		})
	}
}

func TestStream_ContextCancelEmitsPartialDoneMeta(t *testing.T) {
	srv := mockserver.New()
	srv.SetChatStreamChunks([]string{
		`{"model":"` + chatModelID + `","choices":[{"delta":{"content":"partial"}}]}`,
		`{"model":"` + chatModelID + `","choices":[{"delta":{"content":" after cancel"},"finish_reason":"stop"}]}`,
	})
	srv.SetChatStreamDelay(200 * time.Millisecond)
	defer srv.Close()
	a := newAdapter(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := a.Stream(ctx, chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	first := receiveStreamEvent(t, ch)
	if first.Type != aiclient.StreamEventDelta || first.Delta != "partial" {
		t.Fatalf("expected first partial delta, got %+v", first)
	}

	cancel()
	events := collectStreamEvents(t, ch)
	if len(events) != 1 || events[0].Type != aiclient.StreamEventDone || events[0].Meta == nil {
		t.Fatalf("expected one partial done event, got %+v", events)
	}
	meta := events[0].Meta
	if meta.ErrorCode != sharederrors.CodeAiProviderTimeout || meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected invalid timeout meta, got %+v", meta)
	}
	if meta.PartialMetaReason != "context_cancelled" || meta.OutputTokens != len("partial") {
		t.Fatalf("expected partial meta reason and token estimate, got %+v", meta)
	}
}

func collectStreamEvents(t *testing.T, ch <-chan aiclient.AIStreamEvent) []aiclient.AIStreamEvent {
	t.Helper()
	var events []aiclient.AIStreamEvent
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}

func receiveStreamEvent(t *testing.T, ch <-chan aiclient.AIStreamEvent) aiclient.AIStreamEvent {
	t.Helper()
	select {
	case ev, ok := <-ch:
		if !ok {
			t.Fatalf("stream closed before event")
		}
		return ev
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for stream event")
		return aiclient.AIStreamEvent{}
	}
}

type multipartCapture struct {
	values          map[string]string
	fileName        string
	fileContentType string
	fileBytes       []byte
}

func parseMultipartRequest(t *testing.T, contentType string, body []byte) multipartCapture {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("ParseMediaType: %v", err)
	}
	if mediaType != "multipart/form-data" {
		t.Fatalf("expected multipart/form-data, got %q", mediaType)
	}
	reader := multipart.NewReader(strings.NewReader(string(body)), params["boundary"])
	out := multipartCapture{values: map[string]string{}}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("NextPart: %v", err)
		}
		data, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("ReadAll multipart part: %v", err)
		}
		if part.FormName() == "file" {
			out.fileName = part.FileName()
			out.fileContentType = part.Header.Get("Content-Type")
			out.fileBytes = data
			continue
		}
		out.values[part.FormName()] = string(data)
	}
	return out
}

func TestNew_RequiresOpenAICompatibleResolvedProviderSecret(t *testing.T) {
	if _, err := openaicompatible.New(openaicompatible.Options{}); err == nil {
		t.Fatalf("expected error when resolved provider is missing")
	}
	missingBaseURL := resolvedProvider("")
	if _, err := openaicompatible.New(openaicompatible.Options{Provider: missingBaseURL}); err == nil {
		t.Fatalf("expected error when resolved provider BaseURL missing")
	}
	missingAPIKey := resolvedProvider("http://x")
	missingAPIKey.APIKey = ""
	if _, err := openaicompatible.New(openaicompatible.Options{Provider: missingAPIKey}); err == nil {
		t.Fatalf("expected error when resolved provider APIKey missing")
	}
	wrongProtocol := resolvedProvider("http://x")
	wrongProtocol.Entry.Protocol = aiclient.ProviderProtocolStub
	if _, err := openaicompatible.New(openaicompatible.Options{Provider: wrongProtocol}); err == nil {
		t.Fatalf("expected error when provider protocol is not openai_compatible")
	}
}
