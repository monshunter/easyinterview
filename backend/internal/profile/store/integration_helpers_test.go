//go:build integration

package store_test

import (
	"encoding/base64"
	"encoding/json"
)

func decodeBase64URL(raw string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(raw)
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
