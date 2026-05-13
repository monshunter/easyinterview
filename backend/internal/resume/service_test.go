package resume_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
	uploadstore "github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestRegisterUploadPathVerifiesFileObjectBeforeCreate(t *testing.T) {
	now := time.Date(2026, 5, 13, 2, 0, 0, 0, time.UTC)
	uploads := &fakeUploadRegistrar{out: uploadstore.FileObject{
		ID:      "01918fa0-0000-7000-8000-000000000301",
		UserID:  "user-1",
		Purpose: uploadstore.PurposeResume,
		Status:  uploadstore.StatusUploaded,
	}}
	store := &fakeRegisterStore{out: resumestore.CreateAssetResult{
		AssetID:      "01918fa0-0000-7000-8000-000000000101",
		JobID:        "01918fa0-0000-7000-8000-000000000201",
		JobStatus:    "queued",
		JobCreatedAt: now,
		JobUpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{
		Store:          store,
		UploadRegister: uploads,
		Now:            func() time.Time { return now },
		NewID:          sequenceIDs("asset-new", "job-new"),
	})

	out, err := svc.RegisterResume(context.Background(), resume.RegisterInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		SourceType:     "upload",
		FileObjectID:   "01918fa0-0000-7000-8000-000000000301",
		Title:          "Resume",
		Language:       "en",
	})
	if err != nil {
		t.Fatalf("RegisterResume: %v", err)
	}
	if uploads.in.FileObjectID != "01918fa0-0000-7000-8000-000000000301" ||
		uploads.in.OwnerUserID != "user-1" ||
		uploads.in.ExpectedPurpose != uploadstore.PurposeResume {
		t.Fatalf("upload register input = %+v", uploads.in)
	}
	if store.in.FileObjectID == nil || *store.in.FileObjectID != uploads.out.ID {
		t.Fatalf("store fileObjectId = %#v, want %q", store.in.FileObjectID, uploads.out.ID)
	}
	if out.ResumeAssetId != store.out.AssetID || out.Job.ResourceId != store.out.AssetID {
		t.Fatalf("response = %+v, store out = %+v", out, store.out)
	}
}

func TestRegisterUploadPathRejectsMissingObjectBeforeCreate(t *testing.T) {
	uploads := &fakeUploadRegistrar{err: uploadservice.ErrValidationFailed}
	store := &fakeRegisterStore{}
	svc := resume.NewService(resume.ServiceOptions{
		Store:          store,
		UploadRegister: uploads,
		NewID:          sequenceIDs("asset-new", "job-new"),
	})

	_, err := svc.RegisterResume(context.Background(), resume.RegisterInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		SourceType:     "upload",
		FileObjectID:   "01918fa0-0000-7000-8000-000000000301",
		Title:          "Resume",
		Language:       "en",
	})
	if !errors.Is(err, resume.ErrValidationFailed) {
		t.Fatalf("error = %v, want ErrValidationFailed", err)
	}
	if store.calls != 0 {
		t.Fatalf("store CreateWithParseJob calls = %d, want 0", store.calls)
	}
}

type fakeUploadRegistrar struct {
	in  uploadservice.RegisterFileObjectInput
	out uploadstore.FileObject
	err error
}

func (s *fakeUploadRegistrar) RegisterFileObject(_ context.Context, in uploadservice.RegisterFileObjectInput) (uploadstore.FileObject, error) {
	s.in = in
	return s.out, s.err
}

type fakeRegisterStore struct {
	calls int
	in    resumestore.CreateAssetInput
	out   resumestore.CreateAssetResult
	err   error
}

func (s *fakeRegisterStore) CreateWithParseJob(_ context.Context, in resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error) {
	s.calls++
	s.in = in
	return s.out, s.err
}

func sequenceIDs(ids ...string) func() string {
	i := 0
	return func() string {
		if i >= len(ids) {
			return ids[len(ids)-1]
		}
		id := ids[i]
		i++
		return id
	}
}
