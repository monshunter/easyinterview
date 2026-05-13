package objectstore

import (
	"fmt"
	"io"
)

func readAllLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("object read limit must be positive")
	}
	limited := io.LimitReader(r, maxBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maxBytes {
		return nil, ErrObjectTooLarge
	}
	return raw, nil
}
