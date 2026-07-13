package openaicompatible_test

import (
	"bytes"
	"compress/gzip"
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

const responseBodyLimit = 4 << 20

func TestCompleteResponseBodyLimitAndTrustedModelProvenance(t *testing.T) {
	const untrustedModel = "provider-model-RAW-MARKER"
	tests := []struct {
		name       string
		size       int
		gzip       bool
		wantCode   string
		wantModel  string
		wantFinish string
	}{
		{name: "exact limit", size: responseBodyLimit, wantModel: chatModelID, wantFinish: "stop"},
		{name: "one byte over", size: responseBodyLimit + 1, wantCode: sharederrors.CodeAiOutputInvalid},
		{name: "one byte over after gzip decompression", size: responseBodyLimit + 1, gzip: true, wantCode: sharederrors.CodeAiOutputInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := sizedChatResponse(t, tc.size, untrustedModel)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if tc.gzip {
					w.Header().Set("Content-Encoding", "gzip")
					zw := gzip.NewWriter(w)
					_, _ = zw.Write(body)
					_ = zw.Close()
					return
				}
				_, _ = w.Write(body)
			}))
			defer server.Close()

			adapter := newAdapterAt(t, server.URL, server.Client())
			resp, meta, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
			if tc.wantCode != "" {
				assertStableAPIError(t, err, tc.wantCode, untrustedModel)
				return
			}
			if err != nil {
				t.Fatalf("Complete: %v", err)
			}
			if resp.FinishReason != tc.wantFinish {
				t.Fatalf("finish reason=%q, want %q", resp.FinishReason, tc.wantFinish)
			}
			if meta.ModelID != tc.wantModel || meta.ModelFamily != chatModelFamily {
				t.Fatalf("untrusted response model changed provenance: %+v", meta)
			}
			assertDoesNotContain(t, mustJSON(t, meta), untrustedModel)
		})
	}
}

func TestCompleteErrorBodyLimitAndProviderMessageIsolation(t *testing.T) {
	const rawMarker = "PROVIDER-PRIVATE-RAW-MARKER"
	base := []byte(`{"error":{"code":"AI_PROVIDER_TIMEOUT","message":"` + rawMarker + `"}}`)
	tests := []struct {
		name     string
		size     int
		wantCode string
	}{
		{name: "exact limit", size: responseBodyLimit, wantCode: sharederrors.CodeAiProviderTimeout},
		{name: "one byte over", size: responseBodyLimit + 1, wantCode: sharederrors.CodeAiOutputInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(padJSON(t, base, tc.size))
			}))
			defer server.Close()

			adapter := newAdapterAt(t, server.URL, server.Client())
			_, meta, err := adapter.Complete(context.Background(), chatProfile(5000), samplePayload())
			assertStableAPIError(t, err, tc.wantCode, rawMarker)
			assertDoesNotContain(t, mustJSON(t, meta), rawMarker)
		})
	}
}

func TestTranscribeResponseBodyLimit(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantCode string
	}{
		{name: "exact limit", size: responseBodyLimit},
		{name: "one byte over", size: responseBodyLimit + 1, wantCode: sharederrors.CodeAiOutputInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := sizedTranscriptionResponse(t, tc.size)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(body)
			}))
			defer server.Close()

			adapter := newAdapterAt(t, server.URL, server.Client())
			resp, _, err := adapter.Transcribe(context.Background(), sttProfile(5000), aiclient.TranscriptionInput{
				Audio: []byte("audio"), Filename: "answer.webm", ContentType: "audio/webm",
			})
			if tc.wantCode != "" {
				assertStableAPIError(t, err, tc.wantCode)
				return
			}
			if err != nil || resp.Text == "" {
				t.Fatalf("Transcribe exact-limit response failed: text=%d err=%v", len(resp.Text), err)
			}
		})
	}
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

	base := []byte(`{"error":{"code":"AI_PROVIDER_TIMEOUT","message":"private"}}`)
	for _, tc := range []struct {
		name     string
		size     int
		wantCode string
	}{
		{name: "exact error limit", size: responseBodyLimit, wantCode: sharederrors.CodeAiProviderTimeout},
		{name: "error one byte over", size: responseBodyLimit + 1, wantCode: sharederrors.CodeAiOutputInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(padJSON(t, base, tc.size))
			}))
			defer server.Close()
			adapter := newAdapterAt(t, server.URL, server.Client())
			ch, err := adapter.Stream(context.Background(), chatProfile(5000), samplePayload())
			if err != nil {
				t.Fatalf("Stream: %v", err)
			}
			events := collectStreamEvents(t, ch)
			if len(events) != 1 || events[0].ErrorCode != tc.wantCode {
				t.Fatalf("events=%+v, want error code %s", events, tc.wantCode)
			}
		})
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
	adapter, err := openaicompatible.New(openaicompatible.Options{Provider: provider, HTTPClient: client})
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

func sizedTranscriptionResponse(t *testing.T, size int) []byte {
	t.Helper()
	prefix, suffix := `{"text":"`, `"}`
	if size < len(prefix)+len(suffix) {
		t.Fatalf("response size %d too small", size)
	}
	return []byte(prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix)
}

func padJSON(t *testing.T, body []byte, size int) []byte {
	t.Helper()
	if len(body) > size {
		t.Fatalf("body size %d exceeds target %d", len(body), size)
	}
	return append(append([]byte(nil), body...), bytes.Repeat([]byte(" "), size-len(body))...)
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
