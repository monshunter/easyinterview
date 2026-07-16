// Package openaisdk owns the shared OpenAI Go SDK transport policy used only
// by OpenAI-compatible provider adapters.
package openaisdk

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/internal/responsebody"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// MaxRetries is the adapter-owned same-provider retry budget. Cross-provider
// and cross-model fallback remains owned by AIClient.
const MaxRetries = 2

// NewClient constructs an SDK client without reading provider configuration
// from business code. baseURL may be either the provider root or its /v1 URL.
func NewClient(baseURL, apiKey string, httpClient *http.Client, maxResponseBodyBytes int64) openai.Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return openai.NewClient(
		option.WithBaseURL(NormalizeBaseURL(baseURL)),
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
		option.WithMaxRetries(MaxRetries),
		option.WithMiddleware(limitResponseBody(maxResponseBodyBytes)),
	)
}

// NormalizeBaseURL returns the SDK service root expected by relative paths such
// as chat/completions. It deliberately preserves no endpoint-specific suffix.
func NormalizeBaseURL(raw string) string {
	base := strings.TrimRight(raw, "/")
	base = strings.TrimSuffix(base, "/v1")
	return base + "/v1/"
}

func limitResponseBody(maxBytes int64) option.Middleware {
	return func(request *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		response, err := next(request)
		if err != nil || response == nil || response.Body == nil || maxBytes <= 0 {
			return response, err
		}
		if strings.HasPrefix(strings.ToLower(response.Header.Get("Content-Type")), "text/event-stream") {
			response.Body = &eventBoundedReadCloser{body: response.Body, maxBytes: maxBytes}
			return response, nil
		}
		// Several OpenAI-compatible providers omit or mislabel otherwise valid JSON
		// responses. Preserve the adapter's prior tolerant wire contract.
		response.Header.Set("Content-Type", "application/json")
		response.Body = &boundedReadCloser{body: response.Body, remaining: maxBytes}
		return response, nil
	}
}

type boundedReadCloser struct {
	body      io.ReadCloser
	remaining int64
}

func (r *boundedReadCloser) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	if r.remaining == 0 {
		var extra [1]byte
		n, err := r.body.Read(extra[:])
		if n > 0 {
			return 0, responsebody.ErrInvalid
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, io.EOF
			}
			return 0, errors.Join(responsebody.ErrRead, err)
		}
		return 0, responsebody.ErrRead
	}
	if int64(len(buffer)) > r.remaining {
		buffer = buffer[:r.remaining]
	}
	n, err := r.body.Read(buffer)
	r.remaining -= int64(n)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, errors.Join(responsebody.ErrRead, err)
	}
	return n, err
}

func (r *boundedReadCloser) Close() error { return r.body.Close() }

type eventBoundedReadCloser struct {
	body           io.ReadCloser
	maxBytes       int64
	eventBytes     int64
	lineHasContent bool
	failed         bool
}

func (r *eventBoundedReadCloser) Read(buffer []byte) (int, error) {
	if r.failed {
		return 0, responsebody.ErrInvalid
	}
	n, err := r.body.Read(buffer)
	for index, value := range buffer[:n] {
		r.eventBytes++
		if r.eventBytes > r.maxBytes {
			r.failed = true
			if index == 0 {
				return 0, responsebody.ErrInvalid
			}
			return index, nil
		}
		switch value {
		case '\n':
			if !r.lineHasContent {
				r.eventBytes = 0
			}
			r.lineHasContent = false
		case '\r':
		default:
			r.lineHasContent = true
		}
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return n, errors.Join(responsebody.ErrRead, err)
	}
	return n, err
}

func (r *eventBoundedReadCloser) Close() error { return r.body.Close() }
