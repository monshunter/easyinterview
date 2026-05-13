package resume_test

import (
	"context"
	"encoding/json"
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

func TestGetAndListResumesMapStoreRecordsWithUserScope(t *testing.T) {
	now := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	fileID := "01918fa0-0000-7000-8000-000000000301"
	sourceType := "upload"
	parsedText := "parsed resume"
	store := &fakeRegisterStore{
		getOut: resumestore.AssetRecord{
			ID:                 "asset-1",
			UserID:             "user-1",
			FileObjectID:       &fileID,
			Title:              "Resume",
			Language:           "en",
			ParseStatus:        "ready",
			ParsedSummary:      json.RawMessage(`{"headline":"Senior engineer"}`),
			ParsedTextSnapshot: &parsedText,
			SourceType:         &sourceType,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		listOut: resumestore.ListResult{
			Items: []resumestore.AssetRecord{{
				ID:          "asset-1",
				UserID:      "user-1",
				Title:       "Resume",
				Language:    "en",
				ParseStatus: "ready",
				CreatedAt:   now,
				UpdatedAt:   now,
			}},
			NextCursor: "cursor-2",
			HasMore:    true,
			PageSize:   20,
		},
	}
	svc := resume.NewService(resume.ServiceOptions{Store: store})

	got, err := svc.GetResume(context.Background(), "user-1", "asset-1")
	if err != nil {
		t.Fatalf("GetResume: %v", err)
	}
	if store.getUserID != "user-1" || store.getAssetID != "asset-1" {
		t.Fatalf("Get scope user=%q asset=%q", store.getUserID, store.getAssetID)
	}
	if got.Id != "asset-1" || got.FileObjectId == nil || *got.FileObjectId != fileID ||
		got.ParsedSummary == nil || (*got.ParsedSummary)["headline"] != "Senior engineer" ||
		got.ParsedTextSnapshot == nil || *got.ParsedTextSnapshot != parsedText ||
		got.Status == nil || *got.Status != "active" {
		t.Fatalf("GetResume mapped = %+v", got)
	}

	list, err := svc.ListResumes(context.Background(), resume.ListRequest{UserID: "user-1", Cursor: "cursor-1", PageSize: 20})
	if err != nil {
		t.Fatalf("ListResumes: %v", err)
	}
	if store.listUserID != "user-1" || store.listFilter.Cursor != "cursor-1" || store.listFilter.PageSize != 20 {
		t.Fatalf("List scope user=%q filter=%+v", store.listUserID, store.listFilter)
	}
	if len(list.Items) != 1 || list.PageInfo.NextCursor == nil || *list.PageInfo.NextCursor != "cursor-2" || !list.PageInfo.HasMore || list.PageInfo.PageSize != 20 {
		t.Fatalf("ListResumes mapped = %+v", list)
	}
}

func TestGetResumeMapsStoreNotFound(t *testing.T) {
	svc := resume.NewService(resume.ServiceOptions{Store: &fakeRegisterStore{getErr: resumestore.ErrAssetNotFound}})

	_, err := svc.GetResume(context.Background(), "user-1", "asset-missing")

	if !errors.Is(err, resume.ErrNotFound) {
		t.Fatalf("GetResume err = %v, want ErrNotFound", err)
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

	getUserID  string
	getAssetID string
	getOut     resumestore.AssetRecord
	getErr     error

	listUserID string
	listFilter resumestore.ListFilter
	listOut    resumestore.ListResult
	listErr    error
}

func (s *fakeRegisterStore) CreateWithParseJob(_ context.Context, in resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error) {
	s.calls++
	s.in = in
	return s.out, s.err
}

func (s *fakeRegisterStore) Get(_ context.Context, userID string, assetID string) (resumestore.AssetRecord, error) {
	s.getUserID = userID
	s.getAssetID = assetID
	return s.getOut, s.getErr
}

func (s *fakeRegisterStore) List(_ context.Context, userID string, filter resumestore.ListFilter) (resumestore.ListResult, error) {
	s.listUserID = userID
	s.listFilter = filter
	return s.listOut, s.listErr
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
