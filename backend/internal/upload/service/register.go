package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

var ErrValidationFailed = errors.New("upload validation failed")
var ErrRetryableDelete = errors.New("upload delete retryable")

type Repository interface {
	LockForRegister(ctx context.Context, fileObjectID, ownerUserID string, expectedPurpose store.Purpose) (store.FileObject, error)
	MarkUploaded(ctx context.Context, fileObjectID string, now time.Time) error
}

type ObjectStore interface {
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

type RegisterFileObjectInput struct {
	FileObjectID    string
	ExpectedPurpose store.Purpose
	OwnerUserID     string
}

func (s *Service) RegisterFileObject(ctx context.Context, in RegisterFileObjectInput) (store.FileObject, error) {
	if s == nil || s.repository == nil || s.objects == nil {
		return store.FileObject{}, fmt.Errorf("upload register service is not configured")
	}
	rec, err := s.repository.LockForRegister(ctx, in.FileObjectID, in.OwnerUserID, in.ExpectedPurpose)
	if err != nil {
		return store.FileObject{}, err
	}
	switch rec.Status {
	case store.StatusUploaded:
		return rec, nil
	case store.StatusPending:
		ok, err := s.objects.Exists(ctx, rec.ObjectKey)
		if err != nil {
			return store.FileObject{}, err
		}
		if !ok {
			return store.FileObject{}, ErrValidationFailed
		}
		if err := s.repository.MarkUploaded(ctx, rec.ID, s.now()); err != nil {
			return store.FileObject{}, err
		}
		rec.Status = store.StatusUploaded
		return rec, nil
	case store.StatusScanFailed, store.StatusDeleted:
		return store.FileObject{}, ErrValidationFailed
	default:
		return store.FileObject{}, ErrValidationFailed
	}
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
