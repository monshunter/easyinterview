package responsebody_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/internal/responsebody"
)

func TestReadOwnsProviderResponseBodyLimit(t *testing.T) {
	t.Run("reads a bounded identity body", func(t *testing.T) {
		response := responseWithBody("abc", "")
		body, err := responsebody.Read(response, 3)
		if err != nil || string(body) != "abc" {
			t.Fatalf("body=%q err=%v", body, err)
		}
	})

	for _, tc := range []struct {
		name     string
		encoding string
	}{
		{name: "rejects an oversized identity body"},
		{name: "rejects an oversized decompressed body", encoding: "gzip"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := responseWithBody("abcd", tc.encoding)
			if _, err := responsebody.Read(response, 3); !errors.Is(err, responsebody.ErrInvalid) {
				t.Fatalf("error=%v, want ErrInvalid", err)
			}
		})
	}
}

func responseWithBody(body, encoding string) *http.Response {
	var reader io.Reader = bytes.NewBufferString(body)
	if encoding == "gzip" {
		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		_, _ = writer.Write([]byte(body))
		_ = writer.Close()
		reader = &compressed
	}
	return &http.Response{
		Header: http.Header{"Content-Encoding": []string{encoding}},
		Body:   io.NopCloser(reader),
	}
}
