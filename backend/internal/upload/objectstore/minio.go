package objectstore

import (
	"context"
	"errors"
	"fmt"
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
	u, err := client.PresignedPutObject(ctx, s.cfg.Bucket, objectKey, ttl)
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
	if err := s.validate(); err != nil {
		return false, err
	}
	client, err := s.clientFor()
	if err != nil {
		return false, err
	}
	_, err = client.StatObject(ctx, s.cfg.Bucket, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey" {
		return false, nil
	}
	return false, fmt.Errorf("stat minio object: %w", err)
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
