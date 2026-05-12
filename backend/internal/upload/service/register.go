package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

var ErrValidationFailed = errors.New("upload validation failed")
var ErrRetryableDelete = errors.New("upload delete retryable")

type Repository interface {
	Create(ctx context.Context, in store.CreateInput) error
	RegisterUploaded(ctx context.Context, fileObjectID, ownerUserID string, expectedPurpose store.Purpose, now time.Time, exists func(context.Context, string) (bool, error)) (store.FileObject, error)
}

type ObjectStore interface {
	Presign(ctx context.Context, objectKey, contentType string, byteSize int64, ttl time.Duration) (objectstore.PresignResult, error)
	Exists(ctx context.Context, objectKey string) (bool, error)
}

type PrivacyRepository interface {
	ListFileObjectsForUser(ctx context.Context, userID string) ([]store.DeletedFileObject, error)
	HardDelete(ctx context.Context, fileObjectID string) error
	InsertAuditTombstone(ctx context.Context, in store.AuditTombstoneInput) error
}

type DeleteObjectStore interface {
	Delete(ctx context.Context, objectKey string) error
}

type Options struct {
	Repository Repository
	Objects    ObjectStore
	Now        func() time.Time
	NewID      func() string
}

type Service struct {
	repository Repository
	objects    ObjectStore
	now        func() time.Time
	newID      func() string
}

func New(opts Options) *Service {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	return &Service{repository: opts.Repository, objects: opts.Objects, now: now, newID: newID}
}

type CreatePresignInput struct {
	UserID         string
	IdempotencyKey string
	Purpose        string
	FileName       string
	ContentType    string
	ByteSize       int64
	PresignTTL     time.Duration
	MaxBytes       int64
}

func (s *Service) CreateUploadPresign(ctx context.Context, in CreatePresignInput) (api.UploadPresign, error) {
	if s == nil || s.repository == nil || s.objects == nil {
		return api.UploadPresign{}, fmt.Errorf("upload presign service is not configured")
	}
	userID := strings.TrimSpace(in.UserID)
	purpose, ok := publicPurpose(strings.TrimSpace(in.Purpose))
	fileName := strings.TrimSpace(in.FileName)
	contentType := strings.TrimSpace(in.ContentType)
	if userID == "" || !ok || fileName == "" || contentType == "" || in.ByteSize <= 0 || in.PresignTTL <= 0 || in.MaxBytes <= 0 || in.ByteSize > in.MaxBytes {
		return api.UploadPresign{}, ErrValidationFailed
	}

	fileObjectID := s.newID()
	objectKey := objectKeyFor(userID, purpose, fileObjectID, fileName)
	presign, err := s.objects.Presign(ctx, objectKey, contentType, in.ByteSize, in.PresignTTL)
	if err != nil {
		return api.UploadPresign{}, err
	}
	now := s.now()
	if err := s.repository.Create(ctx, store.CreateInput{
		ID:               fileObjectID,
		UserID:           userID,
		Purpose:          purpose,
		ObjectKey:        objectKey,
		OriginalFileName: fileName,
		ContentType:      contentType,
		ByteSize:         in.ByteSize,
		Now:              now,
	}); err != nil {
		return api.UploadPresign{}, err
	}
	return api.UploadPresign{
		FileObjectId: fileObjectID,
		UploadUrl:    presign.URL,
		Method:       presign.Method,
		Headers:      headersToAPI(presign.Headers),
		ExpiresAt:    presign.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

type RegisterFileObjectInput struct {
	FileObjectID    string
	ExpectedPurpose store.Purpose
	OwnerUserID     string
}

func (s *Service) RegisterFileObject(ctx context.Context, in RegisterFileObjectInput) (store.FileObject, error) {
	if s == nil || s.repository == nil || s.objects == nil {
		return store.FileObject{}, fmt.Errorf("upload register service is not configured")
	}
	rec, err := s.repository.RegisterUploaded(ctx, in.FileObjectID, in.OwnerUserID, in.ExpectedPurpose, s.now(), s.objects.Exists)
	if errors.Is(err, store.ErrObjectMissing) || errors.Is(err, store.ErrInvalidStateTransition) {
		return store.FileObject{}, ErrValidationFailed
	}
	return rec, err
}

func (s *Service) DeleteFileObjectsForUser(ctx context.Context, userID string) ([]store.DeletedFileObject, error) {
	if s == nil || s.repository == nil || s.objects == nil {
		return nil, fmt.Errorf("upload delete service is not configured")
	}
	repo, ok := s.repository.(PrivacyRepository)
	if !ok {
		return nil, fmt.Errorf("upload privacy repository is not configured")
	}
	objects, ok := s.objects.(DeleteObjectStore)
	if !ok {
		return nil, fmt.Errorf("upload delete object store is not configured")
	}
	files, err := repo.ListFileObjectsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	for _, file := range files {
		if err := objects.Delete(ctx, file.ObjectKey); err != nil {
			return nil, fmt.Errorf("%w: delete object %s: %v", ErrRetryableDelete, file.ID, err)
		}
		if err := repo.HardDelete(ctx, file.ID); err != nil {
			return nil, err
		}
		if err := repo.InsertAuditTombstone(ctx, store.AuditTombstoneInput{
			AuditEventID: s.newID(),
			UserID:       userID,
			FileObjectID: file.ID,
			Purpose:      file.Purpose,
			DeletedAt:    now,
		}); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func publicPurpose(raw string) (store.Purpose, bool) {
	switch store.Purpose(raw) {
	case store.PurposeResume, store.PurposeTargetJobAttachment, store.PurposePrivacyExport:
		return store.Purpose(raw), true
	default:
		return "", false
	}
}

func objectKeyFor(userID string, purpose store.Purpose, fileObjectID string, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	return strings.Join([]string{userID, string(purpose), fileObjectID + ext}, "/")
}

func headersToAPI(headers map[string]string) map[string]any {
	out := make(map[string]any, len(headers))
	for key, value := range headers {
		out[key] = value
	}
	return out
}
