//go:build integration

package objectstore_test

import (
	"os"
	"testing"
)

func TestMinIO(t *testing.T) {
	if os.Getenv("OBJECT_STORAGE_ENDPOINT") == "" {
		t.Skip("OBJECT_STORAGE_ENDPOINT is not set; skipping MinIO smoke")
	}
	t.Skip("live MinIO SDK smoke is deferred until provider SDK is introduced")
}
