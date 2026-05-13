package objectstore_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
)

func TestFilesystemStorePresignPutExistsAndDelete(t *testing.T) {
	store := objectstore.NewFilesystemStore(t.TempDir())
	ctx := context.Background()

	presign, err := store.Presign(ctx, "user-1/resume/file-1.pdf", "application/pdf", 3, time.Minute)
	if err != nil {
		t.Fatalf("Presign: %v", err)
	}
	if presign.Method != "PUT" || presign.URL == "" || presign.ExpiresAt.IsZero() {
		t.Fatalf("presign = %+v", presign)
	}
	if err := os.WriteFile(presign.LocalPath, []byte("pdf"), 0o600); err != nil {
		t.Fatalf("write signed local path: %v", err)
	}
	ok, err := store.Exists(ctx, "user-1/resume/file-1.pdf")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("expected object to exist")
	}
	if err := store.Delete(ctx, "user-1/resume/file-1.pdf"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, err = store.Exists(ctx, "user-1/resume/file-1.pdf")
	if err != nil {
		t.Fatalf("Exists after delete: %v", err)
	}
	if ok {
		t.Fatal("expected object to be deleted")
	}
}

func TestFilesystemStoreReadRespectsLimit(t *testing.T) {
	store := objectstore.NewFilesystemStore(t.TempDir())
	ctx := context.Background()
	presign, err := store.Presign(ctx, "user-1/resume/file-1.txt", "text/plain", 11, time.Minute)
	if err != nil {
		t.Fatalf("Presign: %v", err)
	}
	if err := os.WriteFile(presign.LocalPath, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write signed local path: %v", err)
	}
	raw, err := store.Read(ctx, "user-1/resume/file-1.txt", 32)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(raw) != "hello world" {
		t.Fatalf("Read = %q", raw)
	}
	_, err = store.Read(ctx, "user-1/resume/file-1.txt", 5)
	if !errors.Is(err, objectstore.ErrObjectTooLarge) {
		t.Fatalf("Read oversize err = %v, want ErrObjectTooLarge", err)
	}
}
