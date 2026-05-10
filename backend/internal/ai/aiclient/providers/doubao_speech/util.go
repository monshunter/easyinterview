package doubaospeech

import (
	"encoding/base64"
	"io"
)

// encodeBase64Audio base64-encodes audio bytes for JSON transport.
func encodeBase64Audio(audio []byte) string {
	return base64.StdEncoding.EncodeToString(audio)
}

// decodeBase64Audio decodes base64-encoded audio bytes from JSON transport.
func decodeBase64Audio(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// readAll is a minimal io.ReadAll wrapper for Go compatibility.
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
