//go:build integration

package objectstore_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
)

func TestMinIO(t *testing.T) {
	cfg := objectstore.MinIOConfig{
		Endpoint:  os.Getenv("OBJECT_STORAGE_ENDPOINT"),
		Bucket:    os.Getenv("OBJECT_STORAGE_BUCKET"),
		AccessKey: os.Getenv("OBJECT_STORAGE_ACCESS_KEY"),
		SecretKey: os.Getenv("OBJECT_STORAGE_SECRET_KEY"),
	}
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		t.Skip("OBJECT_STORAGE_ENDPOINT is not set; skipping MinIO smoke")
	}

	store := objectstore.NewMinIOStore(cfg)
	ctx := context.Background()
	objectKey := fmt.Sprintf("smoke/%d/resume.pdf", time.Now().UnixNano())
	t.Cleanup(func() {
		_ = store.Delete(context.Background(), objectKey)
	})

	presign, err := store.Presign(ctx, objectKey, "application/pdf", int64(len("pdf")), 5*time.Second)
	if err != nil {
		t.Fatalf("Presign: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, presign.Method, presign.URL, stringsReader("pdf"))
	if err != nil {
		t.Fatalf("new PUT request: %v", err)
	}
	for key, value := range presign.Headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT signed URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("PUT signed URL status = %d body=%s", resp.StatusCode, string(body))
	}
	ok, err := store.Exists(ctx, objectKey)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("expected uploaded object to exist")
	}

	expired, err := store.Presign(ctx, objectKey+"-expired", "application/pdf", int64(len("pdf")), time.Second)
	if err != nil {
		t.Fatalf("Presign expired candidate: %v", err)
	}
	time.Sleep(2 * time.Second)
	expiredReq, err := http.NewRequestWithContext(ctx, expired.Method, expired.URL, stringsReader("pdf"))
	if err != nil {
		t.Fatalf("new expired PUT request: %v", err)
	}
	for key, value := range expired.Headers {
		expiredReq.Header.Set(key, value)
	}
	expiredResp, err := http.DefaultClient.Do(expiredReq)
	if err != nil {
		return
	}
	defer expiredResp.Body.Close()
	if expiredResp.StatusCode < 400 {
		t.Fatalf("expired signed URL status = %d, want rejection", expiredResp.StatusCode)
	}
}

func stringsReader(s string) io.Reader {
	return strings.NewReader(s)
}
