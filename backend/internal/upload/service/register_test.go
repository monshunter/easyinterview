package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
	"github.com/monshunter/easyinterview/backend/internal/upload/service"
	"github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestCreateUploadPresignCreatesPendingFileObjectAndPresignsObject(t *testing.T) {
	repo := &fakeRepository{}
	objects := &fakeObjectStore{presign: objectstore.PresignResult{
		URL:       "https://uploads.example/user-1/resume/file-1.pdf",
		Method:    "PUT",
		Headers:   map[string]string{"Content-Type": "application/pdf"},
		ExpiresAt: fixedNow().Add(10 * time.Minute),
	}}
	svc := service.New(service.Options{
		Repository: repo,
		Objects:    objects,
		Now:        fixedNow,
		NewID:      func() string { return "file-1" },
	})

	out, err := svc.CreateUploadPresign(context.Background(), service.CreatePresignInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		Purpose:        string(store.PurposeResume),
		FileName:       "resume.pdf",
		ContentType:    "application/pdf",
		ByteSize:       1024,
		PresignTTL:     10 * time.Minute,
		MaxBytes:       10485760,
	})
	if err != nil {
		t.Fatalf("CreateUploadPresign: %v", err)
	}
	want := api.UploadPresign{
		FileObjectId: "file-1",
		UploadUrl:    "https://uploads.example/user-1/resume/file-1.pdf",
		Method:       "PUT",
		Headers:      map[string]any{"Content-Type": "application/pdf"},
		ExpiresAt:    "2026-05-12T02:10:00Z",
	}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("response = %+v", out)
	}
	if repo.created.ID != "file-1" ||
		repo.created.UserID != "user-1" ||
		repo.created.Purpose != store.PurposeResume ||
		repo.created.ObjectKey != "user-1/resume/file-1.pdf" ||
		repo.created.OriginalFileName != "resume.pdf" ||
		repo.created.ContentType != "application/pdf" ||
		repo.created.ByteSize != 1024 {
		t.Fatalf("created input = %+v", repo.created)
	}
	if objects.presignObjectKey != "user-1/resume/file-1.pdf" ||
		objects.presignContentType != "application/pdf" ||
		objects.presignByteSize != 1024 ||
		objects.presignTTL != 10*time.Minute {
		t.Fatalf("presign call key=%q contentType=%q byteSize=%d ttl=%s", objects.presignObjectKey, objects.presignContentType, objects.presignByteSize, objects.presignTTL)
	}
}

func TestCreateUploadPresignRejectsResumeDOCXBeforePresign(t *testing.T) {
	repo := &fakeRepository{}
	objects := &fakeObjectStore{}
	svc := service.New(service.Options{
		Repository: repo,
		Objects:    objects,
		Now:        fixedNow,
		NewID:      func() string { return "file-1" },
	})

	_, err := svc.CreateUploadPresign(context.Background(), service.CreatePresignInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		Purpose:        string(store.PurposeResume),
		FileName:       "resume.docx",
		ContentType:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		ByteSize:       1024,
		PresignTTL:     10 * time.Minute,
		MaxBytes:       10485760,
	})
	if !errors.Is(err, service.ErrValidationFailed) {
		t.Fatalf("err = %v, want ErrValidationFailed", err)
	}
	if repo.created.ID != "" {
		t.Fatalf("must not create file_object for docx: %+v", repo.created)
	}
	if objects.presignObjectKey != "" {
		t.Fatalf("must not presign docx object: %q", objects.presignObjectKey)
	}
}

func TestCreateUploadPresignRejectsRemovedTargetJobAttachmentPurpose(t *testing.T) {
	repo := &fakeRepository{}
	objects := &fakeObjectStore{}
	svc := service.New(service.Options{
		Repository: repo,
		Objects:    objects,
		Now:        fixedNow,
		NewID:      func() string { return "file-1" },
	})

	_, err := svc.CreateUploadPresign(context.Background(), service.CreatePresignInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		Purpose:        "target_job_attachment",
		FileName:       "job.txt",
		ContentType:    "text/plain",
		ByteSize:       1024,
		PresignTTL:     10 * time.Minute,
		MaxBytes:       10485760,
	})
	if !errors.Is(err, service.ErrValidationFailed) {
		t.Fatalf("err = %v, want ErrValidationFailed", err)
	}
	if repo.created.ID != "" || objects.presignObjectKey != "" {
		t.Fatalf("removed purpose must not create or presign: created=%+v key=%q", repo.created, objects.presignObjectKey)
	}
}

func TestRegisterFileObjectMarksPendingUploadedAfterObjectExists(t *testing.T) {
	repo := &fakeRepository{record: fileObject("file-1", store.StatusPending)}
	objects := &fakeObjectStore{exists: true, statSize: 1024}
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
		size   int64
	}{
		"missing object": {status: store.StatusPending, exists: false},
		"size mismatch":  {status: store.StatusPending, exists: true, size: 2048},
		"scan failed":    {status: store.StatusScanFailed, exists: true, size: 1024},
		"deleted":        {status: store.StatusDeleted, exists: true, size: 1024},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepository{record: fileObject("file-1", setup.status)}
			svc := service.New(service.Options{Repository: repo, Objects: &fakeObjectStore{exists: setup.exists, statSize: setup.size}, Now: fixedNow})
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
		ByteSize:  1024,
		Status:    status,
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 5, 12, 2, 0, 0, 0, time.UTC)
}

type fakeRepository struct {
	record       store.FileObject
	created      store.CreateInput
	markUploaded bool
}

func (r *fakeRepository) Create(_ context.Context, in store.CreateInput) error {
	r.created = in
	return nil
}

func (r *fakeRepository) LockForRegister(_ context.Context, fileObjectID, ownerUserID string, expectedPurpose store.Purpose) (store.FileObject, error) {
	if r.record.ID != fileObjectID || r.record.UserID != ownerUserID || r.record.Purpose != expectedPurpose {
		return store.FileObject{}, store.ErrFileObjectNotFound
	}
	return r.record, nil
}

func (r *fakeRepository) RegisterUploaded(ctx context.Context, fileObjectID, ownerUserID string, expectedPurpose store.Purpose, now time.Time, stat func(context.Context, string) (store.ObjectStat, error)) (store.FileObject, error) {
	rec, err := r.LockForRegister(ctx, fileObjectID, ownerUserID, expectedPurpose)
	if err != nil {
		return store.FileObject{}, err
	}
	if rec.Status == store.StatusUploaded {
		return rec, nil
	}
	if rec.Status != store.StatusPending {
		return store.FileObject{}, store.ErrInvalidStateTransition
	}
	objectStat, err := stat(ctx, rec.ObjectKey)
	if err != nil {
		return store.FileObject{}, err
	}
	if !objectStat.Exists {
		return store.FileObject{}, store.ErrObjectMissing
	}
	if objectStat.Size != rec.ByteSize {
		return store.FileObject{}, store.ErrObjectSizeMismatch
	}
	if err := r.MarkUploaded(ctx, fileObjectID, now); err != nil {
		return store.FileObject{}, err
	}
	rec.Status = store.StatusUploaded
	return rec, nil
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
	exists             bool
	existsKey          string
	statSize           int64
	presign            objectstore.PresignResult
	presignObjectKey   string
	presignContentType string
	presignByteSize    int64
	presignTTL         time.Duration
}

func (s *fakeObjectStore) Presign(_ context.Context, objectKey, contentType string, byteSize int64, ttl time.Duration) (objectstore.PresignResult, error) {
	s.presignObjectKey = objectKey
	s.presignContentType = contentType
	s.presignByteSize = byteSize
	s.presignTTL = ttl
	return s.presign, nil
}

func (s *fakeObjectStore) Exists(_ context.Context, objectKey string) (bool, error) {
	s.existsKey = objectKey
	return s.exists, nil
}

func (s *fakeObjectStore) Stat(_ context.Context, objectKey string) (objectstore.ObjectInfo, error) {
	s.existsKey = objectKey
	if !s.exists {
		return objectstore.ObjectInfo{}, objectstore.ErrObjectNotFound
	}
	return objectstore.ObjectInfo{Size: s.statSize}, nil
}
