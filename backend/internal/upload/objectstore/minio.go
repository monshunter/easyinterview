package objectstore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
}

type MinIOStore struct {
	cfg    MinIOConfig
	client *minio.Client
}

func NewMinIOStore(cfg MinIOConfig) *MinIOStore {
	return &MinIOStore{cfg: cfg}
}

func (s *MinIOStore) Presign(ctx context.Context, objectKey, contentType string, _ int64, ttl time.Duration) (PresignResult, error) {
	if err := s.validate(); err != nil {
		return PresignResult{}, err
	}
	client, err := s.clientFor()
	if err != nil {
		return PresignResult{}, err
	}
	headers := http.Header{}
	headers.Set("Content-Type", contentType)
	headers.Set("x-amz-server-side-encryption", "AES256")
	u, err := client.PresignHeader(ctx, http.MethodPut, s.cfg.Bucket, objectKey, ttl, nil, headers)
	if err != nil {
		return PresignResult{}, fmt.Errorf("presign minio put object: %w", err)
	}
	return PresignResult{
		URL:       u.String(),
		Method:    "PUT",
		Headers:   map[string]string{"Content-Type": contentType, "x-amz-server-side-encryption": "AES256"},
		ExpiresAt: time.Now().UTC().Add(ttl),
	}, nil
}

func (s *MinIOStore) Delete(ctx context.Context, objectKey string) error {
	if err := s.validate(); err != nil {
		return err
	}
	client, err := s.clientFor()
	if err != nil {
		return err
	}
	if err := client.RemoveObject(ctx, s.cfg.Bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete minio object: %w", err)
	}
	return nil
}

func (s *MinIOStore) Exists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.Stat(ctx, objectKey)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrObjectNotFound) {
		return false, nil
	}
	return false, err
}

func (s *MinIOStore) Stat(ctx context.Context, objectKey string) (ObjectInfo, error) {
	if err := s.validate(); err != nil {
		return ObjectInfo{}, err
	}
	client, err := s.clientFor()
	if err != nil {
		return ObjectInfo{}, err
	}
	info, err := client.StatObject(ctx, s.cfg.Bucket, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return ObjectInfo{Size: info.Size}, nil
	}
	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey" {
		return ObjectInfo{}, ErrObjectNotFound
	}
	return ObjectInfo{}, fmt.Errorf("stat minio object: %w", err)
}

func (s *MinIOStore) validate() error {
	if s == nil {
		return fmt.Errorf("minio object store is nil")
	}
	if strings.TrimSpace(s.cfg.Endpoint) == "" || strings.TrimSpace(s.cfg.Bucket) == "" {
		return fmt.Errorf("minio endpoint and bucket are required")
	}
	if strings.TrimSpace(s.cfg.AccessKey) == "" || strings.TrimSpace(s.cfg.SecretKey) == "" {
		return fmt.Errorf("minio credentials are required")
	}
	return nil
}

func (s *MinIOStore) clientFor() (*minio.Client, error) {
	if s.client != nil {
		return s.client, nil
	}
	u, err := url.Parse(strings.TrimRight(s.cfg.Endpoint, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse minio endpoint: %w", err)
	}
	endpoint := u.Host
	secure := u.Scheme == "https"
	if endpoint == "" {
		endpoint = u.Path
		secure = false
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.cfg.AccessKey, s.cfg.SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	s.client = client
	return client, nil
}
