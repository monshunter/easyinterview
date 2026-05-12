package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

var ErrValidationFailed = errors.New("upload validation failed")

type Repository interface {
	LockForRegister(ctx context.Context, fileObjectID, ownerUserID string, expectedPurpose store.Purpose) (store.FileObject, error)
	MarkUploaded(ctx context.Context, fileObjectID string, now time.Time) error
}

type ObjectStore interface {
	Exists(ctx context.Context, objectKey string) (bool, error)
}

type Options struct {
	Repository Repository
	Objects    ObjectStore
	Now        func() time.Time
}

type Service struct {
	repository Repository
	objects    ObjectStore
	now        func() time.Time
}

func New(opts Options) *Service {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{repository: opts.Repository, objects: opts.Objects, now: now}
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
