package objectstore_test

import (
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
)

func TestNewFromConfigSelectsProvider(t *testing.T) {
	for _, provider := range []string{"filesystem", "minio"} {
		t.Run(provider, func(t *testing.T) {
			store, err := objectstore.NewFromConfig(objectstore.FactoryConfig{
				Provider:       provider,
				FilesystemRoot: t.TempDir(),
				MinIO: objectstore.MinIOConfig{
					Endpoint:  "http://localhost:9000",
					Bucket:    "easyinterview-dev",
					AccessKey: "dev-access-key",
					SecretKey: "dev-secret-key",
				},
			})
			if err != nil {
				t.Fatalf("NewFromConfig: %v", err)
			}
			if store == nil {
				t.Fatal("store is nil")
			}
		})
	}
}

func TestNewFromConfigRejectsUnknownProvider(t *testing.T) {
	if _, err := objectstore.NewFromConfig(objectstore.FactoryConfig{Provider: "s3"}); err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}
