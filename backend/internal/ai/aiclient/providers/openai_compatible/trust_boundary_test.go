package openaicompatible_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/openai_compatible"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const responseBodyLimit = 256

func TestCompleteUsesTrustedModelProvenance(t *testing.T) {
	const untrustedModel = "provider-model-RAW-MARKER"
	body := sizedChatResponse(t, responseBodyLimit, untrustedModel)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	adapter := newAdapterAt(t, server.URL, server.Client())
	resp, meta, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.FinishReason != "stop" {
		t.Fatalf("finish reason=%q, want stop", resp.FinishReason)
	}
	if meta.ModelID != chatModelID || meta.ModelFamily != chatModelFamily {
		t.Fatalf("untrusted response model changed provenance: %+v", meta)
	}
	assertDoesNotContain(t, mustJSON(t, meta), untrustedModel)
}

func TestCompleteKeepsProviderErrorMessagePrivate(t *testing.T) {
	const rawMarker = "PROVIDER-PRIVATE-RAW-MARKER"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"AI_PROVIDER_TIMEOUT","message":"`+rawMarker+`"}}`)
	}))
	defer server.Close()

	adapter := newAdapterAt(t, server.URL, server.Client())
	_, meta, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
	assertStableAPIError(t, err, sharederrors.CodeAiProviderTimeout, rawMarker)
	assertDoesNotContain(t, mustJSON(t, meta), rawMarker)
}

func TestCompleteRejectsUntrustedRouteAndFallbackHeaders(t *testing.T) {
	const rawMarker = "sk-UNTRUSTED_HEADER_RAW_MARKER"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(openaicompatible.HeaderRoute, rawMarker)
		w.Header().Set(openaicompatible.HeaderFallbackFrom, rawMarker)
		w.Header().Set(openaicompatible.HeaderFallbackTo, rawMarker)
		_, _ = w.Write(sizedChatResponse(t, 256, "malicious-response-model"))
	}))
	defer server.Close()

	adapter := newAdapterAt(t, server.URL, server.Client())
	_, meta, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if meta.Route != chatProfile(0).Route || len(meta.FallbackChain) != 0 {
		t.Fatalf("untrusted headers entered meta: %+v", meta)
	}
	assertDoesNotContain(t, mustJSON(t, meta), rawMarker)
}

func TestCompleteBoundsProviderFinishReason(t *testing.T) {
	const rawMarker = "UNTRUSTED-FINISH-REASON-RAW-MARKER"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"model":"ignored","choices":[{"message":{"content":"ok"},"finish_reason":"`+rawMarker+`"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
	}))
	defer server.Close()
	adapter := newAdapterAt(t, server.URL, server.Client())
	response, _, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if response.FinishReason != "unknown" || strings.Contains(response.FinishReason, rawMarker) {
		t.Fatalf("untrusted finish reason escaped boundary: %+v", response)
	}
}

func TestStreamUsesTrustedModelAndBoundsErrorBody(t *testing.T) {
	t.Run("response model cannot change provenance", func(t *testing.T) {
		const rawModel = "stream-model-RAW-MARKER"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, `data: {"model":"`+rawModel+`","choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}]}`+"\n\n")
		}))
		defer server.Close()

		adapter := newAdapterAt(t, server.URL, server.Client())
		ch, err := adapter.Stream(context.Background(), chatProfile(5000), samplePayload())
		if err != nil {
			t.Fatalf("Stream: %v", err)
		}
		events := collectStreamEvents(t, ch)
		meta := events[len(events)-1].Meta
		if meta == nil || meta.ModelID != chatModelID {
			t.Fatalf("untrusted stream model changed provenance: %+v", events)
		}
		assertDoesNotContain(t, mustJSON(t, meta), rawModel)
	})

	t.Run("maps a provider error to a stable code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"code":"AI_PROVIDER_TIMEOUT","message":"private"}}`)
		}))
		defer server.Close()
		adapter := newAdapterAt(t, server.URL, server.Client())
		ch, err := adapter.Stream(context.Background(), chatProfile(5000), samplePayload())
		if err != nil {
			t.Fatalf("Stream: %v", err)
		}
		events := collectStreamEvents(t, ch)
		if len(events) != 1 || events[0].ErrorCode != sharederrors.CodeAiProviderTimeout {
			t.Fatalf("events=%+v, want timeout error", events)
		}
	})
}

func TestStreamSuccessBodyLimit(t *testing.T) {
	body := sizedStreamResponse(t, responseBodyLimit+1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	adapter := newAdapterAt(t, server.URL, server.Client())
	ch, err := adapter.Stream(context.Background(), chatProfile(5000), samplePayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	events := collectStreamEvents(t, ch)
	if len(events) != 1 || events[0].Type != aiclient.StreamEventError || events[0].ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("events=%+v, want one output-invalid event", events)
	}
}

func TestCompleteDoesNotExposeTransportReadOrParseErrors(t *testing.T) {
	const rawMarker = "PRIVATE-URL-AND-BODY-RAW-MARKER"
	tests := []struct {
		name   string
		client *http.Client
	}{
		{
			name: "transport",
			client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New(rawMarker)
			})},
		},
		{
			name: "read",
			client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &errorReadCloser{err: errors.New(rawMarker)}}, nil
			})},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adapter := newAdapterAt(t, "https://provider.invalid/"+rawMarker, tc.client)
			_, meta, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
			assertStableAPIError(t, err, sharederrors.CodeAiProviderTimeout, rawMarker, "provider.invalid")
			assertDoesNotContain(t, mustJSON(t, meta), rawMarker)
		})
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, rawMarker)
	}))
	defer server.Close()
	adapter := newAdapterAt(t, server.URL, server.Client())
	_, _, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
	assertStableAPIError(t, err, sharederrors.CodeAiOutputInvalid, rawMarker)
}

func newAdapterAt(t *testing.T, baseURL string, client *http.Client) *openaicompatible.Adapter {
	t.Helper()
	provider := resolvedProvider(baseURL)
	adapter, err := openaicompatible.New(openaicompatible.Options{Provider: provider, HTTPClient: client, MaxResponseBodyBytes: responseBodyLimit})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return adapter
}

func sizedChatResponse(t *testing.T, size int, model string) []byte {
	t.Helper()
	prefix := `{"model":"` + model + `","choices":[{"message":{"content":"`
	suffix := `"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
	if size < len(prefix)+len(suffix) {
		t.Fatalf("response size %d too small", size)
	}
	return []byte(prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix)
}

func sizedStreamResponse(t *testing.T, size int) []byte {
	t.Helper()
	prefix := `data: {"model":"ignored","choices":[{"delta":{"content":"`
	suffix := `"},"finish_reason":"stop"}]}` + "\n\n"
	if size < len(prefix)+len(suffix) {
		t.Fatalf("stream response size %d too small", size)
	}
	return []byte(prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix)
}

func assertStableAPIError(t *testing.T, err error, code string, forbidden ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s error", code)
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	meta := sharederrors.CodeRegistry[code]
	if apiErr.Code != code || apiErr.Message != meta.Message || apiErr.Retryable != meta.Retryable {
		t.Fatalf("unstable API error: %+v, registry=%+v", apiErr, meta)
	}
	assertDoesNotContain(t, err.Error(), forbidden...)
}

func assertDoesNotContain(t *testing.T, value string, forbidden ...string) {
	t.Helper()
	for _, marker := range forbidden {
		if marker != "" && strings.Contains(value, marker) {
			t.Fatalf("value leaked %q: %s", marker, value)
		}
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(raw)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

type errorReadCloser struct{ err error }

func (r *errorReadCloser) Read([]byte) (int, error) { return 0, r.err }
func (r *errorReadCloser) Close() error             { return nil }
