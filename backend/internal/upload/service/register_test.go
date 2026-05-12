package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/upload/service"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestRegisterFileObjectMarksPendingUploadedAfterObjectExists(t *testing.T) {
	repo := &fakeRepository{record: fileObject("file-1", store.StatusPending)}
	objects := &fakeObjectStore{exists: true}
	svc := service.New(service.Options{Repository: repo, Objects: objects, Now: fixedNow})

	rec, err := svc.RegisterFileObject(context.Background(), service.RegisterFileObjectInput{
		FileObjectID:    "file-1",
		OwnerUserID:     "user-1",
		ExpectedPurpose: store.PurposeResume,
	})
	if err != nil {
		t.Fatalf("RegisterFileObject: %v", err)
	}
	if rec.ID != "file-1" || !repo.markUploaded {
		t.Fatalf("record=%+v markUploaded=%v", rec, repo.markUploaded)
	}
	if objects.existsKey != "user-1/resume/file-1.pdf" {
		t.Fatalf("exists key = %q", objects.existsKey)
	}
}

func TestRegisterFileObjectUploadedIsIdempotent(t *testing.T) {
	repo := &fakeRepository{record: fileObject("file-1", store.StatusUploaded)}
	svc := service.New(service.Options{Repository: repo, Objects: &fakeObjectStore{exists: false}, Now: fixedNow})

	if _, err := svc.RegisterFileObject(context.Background(), service.RegisterFileObjectInput{
		FileObjectID:    "file-1",
		OwnerUserID:     "user-1",
		ExpectedPurpose: store.PurposeResume,
	}); err != nil {
		t.Fatalf("RegisterFileObject uploaded idempotent: %v", err)
	}
	if repo.markUploaded {
		t.Fatal("uploaded row must not be marked again")
	}
}

func TestRegisterFileObjectRejectsMissingObjectAndIllegalStates(t *testing.T) {
	for name, setup := range map[string]struct {
		status store.UploadStatus
		exists bool
	}{
		"missing object": {status: store.StatusPending, exists: false},
		"scan failed":    {status: store.StatusScanFailed, exists: true},
		"deleted":        {status: store.StatusDeleted, exists: true},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepository{record: fileObject("file-1", setup.status)}
			svc := service.New(service.Options{Repository: repo, Objects: &fakeObjectStore{exists: setup.exists}, Now: fixedNow})
			_, err := svc.RegisterFileObject(context.Background(), service.RegisterFileObjectInput{
				FileObjectID:    "file-1",
				OwnerUserID:     "user-1",
				ExpectedPurpose: store.PurposeResume,
			})
			if !errors.Is(err, service.ErrValidationFailed) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if repo.markUploaded {
				t.Fatal("must not mark uploaded on invalid register")
			}
		})
	}
}

func fileObject(id string, status store.UploadStatus) store.FileObject {
	return store.FileObject{
		ID:        id,
		UserID:    "user-1",
		Purpose:   store.PurposeResume,
		ObjectKey: "user-1/resume/file-1.pdf",
		Status:    status,
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 5, 12, 2, 0, 0, 0, time.UTC)
}

type fakeRepository struct {
	record       store.FileObject
	markUploaded bool
}

func (r *fakeRepository) LockForRegister(_ context.Context, fileObjectID, ownerUserID string, expectedPurpose store.Purpose) (store.FileObject, error) {
	if r.record.ID != fileObjectID || r.record.UserID != ownerUserID || r.record.Purpose != expectedPurpose {
		return store.FileObject{}, store.ErrFileObjectNotFound
	}
	return r.record, nil
}

func (r *fakeRepository) MarkUploaded(_ context.Context, fileObjectID string, _ time.Time) error {
	if r.record.ID != fileObjectID {
		return store.ErrFileObjectNotFound
	}
	r.markUploaded = true
	r.record.Status = store.StatusUploaded
	return nil
}

type fakeObjectStore struct {
	exists    bool
	existsKey string
}

func (s *fakeObjectStore) Exists(_ context.Context, objectKey string) (bool, error) {
	s.existsKey = objectKey
	return s.exists, nil
}
