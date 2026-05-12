package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/upload/service"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestDeleteFileObjectsForUserDeletesObjectsBeforeDBAndWritesAudit(t *testing.T) {
	repo := &fakePrivacyRepository{files: []store.DeletedFileObject{
		{ID: "file-1", ObjectKey: "user-1/resume/file-1.pdf", Purpose: store.PurposeResume},
		{ID: "file-2", ObjectKey: "user-1/resume/file-2.pdf", Purpose: store.PurposeResume},
	}}
	objects := &fakeDeleteObjectStore{}
	svc := service.New(service.Options{Repository: repo, Objects: objects, Now: fixedNow})

	deleted, err := svc.DeleteFileObjectsForUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("DeleteFileObjectsForUser: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("deleted = %+v", deleted)
	}
	if len(objects.deletedKeys) != 2 || len(repo.atomicDeletedIDs) != 2 {
		t.Fatalf("objects=%v atomic=%v", objects.deletedKeys, repo.atomicDeletedIDs)
	}
	if repo.atomicDeletedIDs[0] != "file-1" {
		t.Fatalf("delete/audit order mismatch: atomic=%v", repo.atomicDeletedIDs)
	}
	if len(repo.hardDeletedIDs) != 0 || len(repo.auditIDs) != 0 {
		t.Fatalf("legacy split delete/audit should not be used: hard=%v audit=%v", repo.hardDeletedIDs, repo.auditIDs)
	}
}

func TestDeleteFileObjectsForUserObjectDeleteFailureIsRetryableAndKeepsDBRows(t *testing.T) {
	repo := &fakePrivacyRepository{files: []store.DeletedFileObject{
		{ID: "file-1", ObjectKey: "user-1/resume/file-1.pdf", Purpose: store.PurposeResume},
	}}
	objects := &fakeDeleteObjectStore{deleteErr: errors.New("minio down")}
	svc := service.New(service.Options{Repository: repo, Objects: objects, Now: fixedNow})

	_, err := svc.DeleteFileObjectsForUser(context.Background(), "user-1")
	if !errors.Is(err, service.ErrRetryableDelete) {
		t.Fatalf("expected retryable delete error, got %v", err)
	}
	if len(repo.hardDeletedIDs) != 0 || len(repo.auditIDs) != 0 {
		t.Fatalf("DB row/audit must not be mutated on object failure: hard=%v audit=%v", repo.hardDeletedIDs, repo.auditIDs)
	}
}

func TestDeleteFileObjectsForUserUsesAtomicDBDeleteAndAudit(t *testing.T) {
	repo := &fakePrivacyRepository{files: []store.DeletedFileObject{
		{ID: "file-1", ObjectKey: "user-1/resume/file-1.pdf", Purpose: store.PurposeResume},
	}}
	objects := &fakeDeleteObjectStore{}
	svc := service.New(service.Options{Repository: repo, Objects: objects, Now: fixedNow, NewID: func() string { return "audit-1" }})

	_, err := svc.DeleteFileObjectsForUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("DeleteFileObjectsForUser: %v", err)
	}
	if len(repo.atomicDeletedIDs) != 1 || repo.atomicDeletedIDs[0] != "file-1" {
		t.Fatalf("atomicDeletedIDs=%v", repo.atomicDeletedIDs)
	}
	if len(repo.hardDeletedIDs) != 0 || len(repo.auditIDs) != 0 {
		t.Fatalf("legacy split delete/audit should not be used: hard=%v audit=%v", repo.hardDeletedIDs, repo.auditIDs)
	}
}

type fakePrivacyRepository struct {
	fakeRepository
	files            []store.DeletedFileObject
	hardDeletedIDs   []string
	auditIDs         []string
	atomicDeletedIDs []string
}

func (r *fakePrivacyRepository) ListFileObjectsForUser(ctx context.Context, userID string) ([]store.DeletedFileObject, error) {
	return r.files, nil
}

func (r *fakePrivacyRepository) HardDelete(ctx context.Context, fileObjectID string) error {
	r.hardDeletedIDs = append(r.hardDeletedIDs, fileObjectID)
	return nil
}

func (r *fakePrivacyRepository) InsertAuditTombstone(ctx context.Context, in store.AuditTombstoneInput) error {
	r.auditIDs = append(r.auditIDs, in.FileObjectID)
	if in.ObjectKey != "" {
		return errors.New("audit tombstone leaked object key")
	}
	return nil
}

func (r *fakePrivacyRepository) HardDeleteWithAuditTombstone(ctx context.Context, in store.AuditTombstoneInput) error {
	r.atomicDeletedIDs = append(r.atomicDeletedIDs, in.FileObjectID)
	if in.ObjectKey != "" {
		return errors.New("audit tombstone leaked object key")
	}
	return nil
}

type fakeDeleteObjectStore struct {
	fakeObjectStore
	deletedKeys []string
	deleteErr   error
}

func (s *fakeDeleteObjectStore) Delete(ctx context.Context, objectKey string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deletedKeys = append(s.deletedKeys, objectKey)
	return nil
}

var _ = time.Time{}
