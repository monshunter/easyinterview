package objectstore

import (
	"context"
	"time"
)

type PresignResult struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
	// LocalPath is populated only by the filesystem fallback so unit tests can
	// write bytes without a network service.
	LocalPath string
}

type ObjectStore interface {
	Presign(ctx context.Context, objectKey, contentType string, byteSize int64, ttl time.Duration) (PresignResult, error)
	Delete(ctx context.Context, objectKey string) error
	Exists(ctx context.Context, objectKey string) (bool, error)
}
