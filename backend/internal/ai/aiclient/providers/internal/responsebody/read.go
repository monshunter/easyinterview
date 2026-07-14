package responsebody

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

var (
	ErrInvalid = errors.New("invalid provider response body")
	ErrRead    = errors.New("failed to read provider response body")
)

func Read(response *http.Response, maxBytes int64) ([]byte, error) {
	if response == nil || response.Body == nil || maxBytes <= 0 {
		return nil, ErrInvalid
	}

	reader := io.Reader(response.Body)
	compressed := false
	if !response.Uncompressed {
		switch strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Encoding"))) {
		case "", "identity":
		case "gzip":
			gzipReader, err := gzip.NewReader(response.Body)
			if err != nil {
				return nil, ErrInvalid
			}
			defer gzipReader.Close()
			reader = gzipReader
			compressed = true
		default:
			return nil, ErrInvalid
		}
	}

	body, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		if compressed {
			return nil, ErrInvalid
		}
		return nil, ErrRead
	}
	if int64(len(body)) > maxBytes {
		return nil, ErrInvalid
	}
	return body, nil
}
