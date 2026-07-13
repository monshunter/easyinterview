package judgecompatible_test

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
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	judgecompatible "github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/judge_compatible"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const judgeResponseBodyLimit = 4 << 20

func TestCompleteResponseLimitAndTrustedModelProvenance(t *testing.T) {
	const rawModel = "JUDGE-PROVIDER-MODEL-RAW-MARKER"
	for _, tc := range []struct {
		name     string
		size     int
		gzip     bool
		wantCode string
	}{
		{name: "exact limit", size: judgeResponseBodyLimit},
		{name: "one byte over", size: judgeResponseBodyLimit + 1, wantCode: sharederrors.CodeAiOutputInvalid},
		{name: "one byte over after gzip decompression", size: judgeResponseBodyLimit + 1, gzip: true, wantCode: sharederrors.CodeAiOutputInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := sizedJudgeResponse(t, tc.size, rawModel)
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

			adapter := newJudgeAdapterAt(t, server.URL, server.Client())
			_, meta, err := adapter.Complete(context.Background(), judgeProfile(5000), judgePayload())
			if tc.wantCode != "" {
				assertJudgeStableError(t, err, tc.wantCode, rawModel)
				return
			}
			if err != nil {
				t.Fatalf("Complete: %v", err)
			}
			if meta.ModelID != judgeProfile(0).Default.Model {
				t.Fatalf("untrusted response model changed provenance: %+v", meta)
			}
			assertJudgeDoesNotContain(t, judgeMustJSON(t, meta), rawModel)
		})
	}
}

func TestCompleteErrorResponseLimit(t *testing.T) {
	base := []byte(`{"error":{"code":"private","message":"private"}}`)
	for _, tc := range []struct {
		name     string
		size     int
		wantCode string
	}{
		{name: "exact limit", size: judgeResponseBodyLimit, wantCode: sharederrors.CodeAiProviderTimeout},
		{name: "one byte over", size: judgeResponseBodyLimit + 1, wantCode: sharederrors.CodeAiOutputInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write(judgePadJSON(t, base, tc.size))
			}))
			defer server.Close()
			adapter := newJudgeAdapterAt(t, server.URL, server.Client())
			_, _, err := adapter.Complete(context.Background(), judgeProfile(5000), judgePayload())
			assertJudgeStableError(t, err, tc.wantCode, "private")
		})
	}
}

func TestCompleteDoesNotExposeTransportReadOrParseErrors(t *testing.T) {
	const rawMarker = "JUDGE-PRIVATE-URL-AND-BODY-RAW-MARKER"
	for _, tc := range []struct {
		name   string
		client *http.Client
	}{
		{
			name: "transport",
			client: &http.Client{Transport: judgeRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New(rawMarker)
			})},
		},
		{
			name: "read",
			client: &http.Client{Transport: judgeRoundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &judgeErrorReadCloser{err: errors.New(rawMarker)}}, nil
			})},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			adapter := newJudgeAdapterAt(t, "https://provider.invalid/"+rawMarker, tc.client)
			_, meta, err := adapter.Complete(context.Background(), judgeProfile(5000), judgePayload())
			assertJudgeStableError(t, err, sharederrors.CodeAiProviderTimeout, rawMarker, "provider.invalid")
			assertJudgeDoesNotContain(t, judgeMustJSON(t, meta), rawMarker)
		})
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, rawMarker)
	}))
	defer server.Close()
	adapter := newJudgeAdapterAt(t, server.URL, server.Client())
	_, _, err := adapter.Complete(context.Background(), judgeProfile(5000), judgePayload())
	assertJudgeStableError(t, err, sharederrors.CodeAiOutputInvalid, rawMarker)
}

func TestCompleteBoundsProviderFinishReason(t *testing.T) {
	const rawMarker = "JUDGE-UNTRUSTED-FINISH-REASON-RAW-MARKER"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"model":"ignored","choices":[{"message":{"content":"{}"},"finish_reason":"`+rawMarker+`"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
	}))
	defer server.Close()
	adapter := newJudgeAdapterAt(t, server.URL, server.Client())
	response, _, err := adapter.Complete(context.Background(), judgeProfile(5000), judgePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if response.FinishReason != "unknown" || strings.Contains(response.FinishReason, rawMarker) {
		t.Fatalf("untrusted finish reason escaped boundary: %+v", response)
	}
}

func newJudgeAdapterAt(t *testing.T, baseURL string, client *http.Client) *judgecompatible.Adapter {
	t.Helper()
	provider := providerregistry.ResolvedProvider{Entry: judgeEntry(), BaseURL: baseURL, APIKey: "k"}
	adapter, err := judgecompatible.New(judgecompatible.Options{Provider: provider, HTTPClient: client})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return adapter
}

func judgePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{Messages: []aiclient.Message{{Role: "user", Content: "score this"}}}
}

func sizedJudgeResponse(t *testing.T, size int, model string) []byte {
	t.Helper()
	prefix := `{"model":"` + model + `","choices":[{"message":{"content":"`
	suffix := `"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
	if size < len(prefix)+len(suffix) {
		t.Fatalf("response size %d too small", size)
	}
	return []byte(prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix)
}

func judgePadJSON(t *testing.T, body []byte, size int) []byte {
	t.Helper()
	if len(body) > size {
		t.Fatalf("body size %d exceeds target %d", len(body), size)
	}
	return append(append([]byte(nil), body...), bytes.Repeat([]byte(" "), size-len(body))...)
}

func assertJudgeStableError(t *testing.T, err error, code string, forbidden ...string) {
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
	assertJudgeDoesNotContain(t, err.Error(), forbidden...)
}

func assertJudgeDoesNotContain(t *testing.T, value string, forbidden ...string) {
	t.Helper()
	for _, marker := range forbidden {
		if marker != "" && strings.Contains(value, marker) {
			t.Fatalf("value leaked %q: %s", marker, value)
		}
	}
}

func judgeMustJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(raw)
}

type judgeRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn judgeRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

type judgeErrorReadCloser struct{ err error }

func (r *judgeErrorReadCloser) Read([]byte) (int, error) { return 0, r.err }
func (r *judgeErrorReadCloser) Close() error             { return nil }
