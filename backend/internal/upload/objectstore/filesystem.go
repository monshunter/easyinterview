package objectstore

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FilesystemStore struct {
	root string
	now  func() time.Time
}

func NewFilesystemStore(root string) *FilesystemStore {
	return &FilesystemStore{root: root, now: func() time.Time { return time.Now().UTC() }}
}

func (s *FilesystemStore) Presign(_ context.Context, objectKey, contentType string, _ int64, ttl time.Duration) (PresignResult, error) {
	path, err := s.pathFor(objectKey)
	if err != nil {
		return PresignResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return PresignResult{}, fmt.Errorf("prepare filesystem object path: %w", err)
	}
	return PresignResult{
		URL:       "file://" + url.PathEscape(objectKey),
		Method:    "PUT",
		Headers:   map[string]string{"Content-Type": contentType},
		ExpiresAt: s.now().Add(ttl),
		LocalPath: path,
	}, nil
}

func (s *FilesystemStore) Delete(_ context.Context, objectKey string) error {
	path, err := s.pathFor(objectKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete filesystem object: %w", err)
	}
	return nil
}

func (s *FilesystemStore) Exists(_ context.Context, objectKey string) (bool, error) {
	_, err := s.Stat(context.Background(), objectKey)
	if err == nil {
		return true, nil
	}
	if err == ErrObjectNotFound {
		return false, nil
	}
	return false, err
}

func (s *FilesystemStore) Stat(_ context.Context, objectKey string) (ObjectInfo, error) {
	path, err := s.pathFor(objectKey)
	if err != nil {
		return ObjectInfo{}, err
	}
	info, err := os.Stat(path)
	if err == nil {
		return ObjectInfo{Size: info.Size()}, nil
	}
	if os.IsNotExist(err) {
		return ObjectInfo{}, ErrObjectNotFound
	}
	return ObjectInfo{}, fmt.Errorf("stat filesystem object: %w", err)
}

func (s *FilesystemStore) pathFor(objectKey string) (string, error) {
	if s == nil || strings.TrimSpace(s.root) == "" {
		return "", fmt.Errorf("filesystem object store root is required")
	}
	clean := filepath.Clean(strings.TrimPrefix(objectKey, "/"))
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("object key escapes filesystem root")
	}
	return filepath.Join(s.root, clean), nil
}
