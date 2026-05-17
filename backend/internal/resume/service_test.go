package resume_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
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

func TestConfirmStructuredMasterCreatesStructuredMasterVersion(t *testing.T) {
	now := time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC)
	store := &fakeRegisterStore{structuredOut: resumestore.VersionRecord{
		ID:                "version-1",
		UserID:            "user-1",
		ResumeAssetID:     "asset-1",
		VersionType:       "structured_master",
		DisplayName:       "Structured master",
		StructuredProfile: json.RawMessage(`{"headline":"Senior engineer","provenance":{"promptVersion":"resume_profile.v1","rubricVersion":"not_applicable","modelId":"model-1","language":"en","featureFlag":"none","dataSourceVersion":"asset.v1"}}`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "resume_profile.v1", RubricVersion: "not_applicable", ModelID: "model-1", Language: "en", FeatureFlag: "none", DataSourceVersion: "asset.v1",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("version-1"),
	})

	got, err := svc.ConfirmStructuredMaster(context.Background(), resume.ConfirmStructuredMasterInput{
		UserID:        "user-1",
		ResumeAssetID: "asset-1",
		DisplayName:   " Structured master ",
		Language:      "en",
		StructuredProfile: map[string]any{
			"headline": "Senior engineer",
			"provenance": map[string]any{
				"promptVersion": "resume_profile.v1", "rubricVersion": "not_applicable", "modelId": "model-1", "language": "en", "featureFlag": "none", "dataSourceVersion": "asset.v1",
			},
		},
	})
	if err != nil {
		t.Fatalf("ConfirmStructuredMaster: %v", err)
	}
	if store.structuredIn.VersionID != "version-1" || store.structuredIn.UserID != "user-1" || store.structuredIn.ResumeAssetID != "asset-1" || store.structuredIn.DisplayName != "Structured master" {
		t.Fatalf("store input = %+v", store.structuredIn)
	}
	if store.structuredIn.Provenance.PromptVersion != "resume_profile.v1" || store.structuredIn.Provenance.ModelID != "model-1" {
		t.Fatalf("store provenance = %+v", store.structuredIn.Provenance)
	}
	if got.Id != "version-1" || got.VersionType != "structured_master" || got.ParentVersionId != nil || got.SeedStrategy != nil || len(got.Suggestions) != 0 {
		t.Fatalf("response = %+v", got)
	}
	if got.Provenance.PromptVersion != "resume_profile.v1" || got.PromptVersion == nil || *got.PromptVersion != "resume_profile.v1" {
		t.Fatalf("response provenance = %+v prompt=%v", got.Provenance, got.PromptVersion)
	}
}

func TestConfirmStructuredMasterValidationAndStoreErrors(t *testing.T) {
	tests := []struct {
		name string
		in   resume.ConfirmStructuredMasterInput
		err  error
		want error
	}{
		{name: "missing display name", in: resume.ConfirmStructuredMasterInput{UserID: "user-1", ResumeAssetID: "asset-1", StructuredProfile: validStructuredProfile()}, want: resume.ErrValidationFailed},
		{name: "missing provenance", in: resume.ConfirmStructuredMasterInput{UserID: "user-1", ResumeAssetID: "asset-1", DisplayName: "Master", StructuredProfile: map[string]any{"headline": "Senior engineer"}}, want: resume.ErrValidationFailed},
		{name: "not found", in: validConfirmInput(), err: resumestore.ErrAssetNotFound, want: resume.ErrNotFound},
		{name: "parse not ready", in: validConfirmInput(), err: resumestore.ErrAssetParseNotReady, want: resume.ErrAssetParseNotReady},
		{name: "already exists", in: validConfirmInput(), err: resumestore.ErrStructuredMasterAlreadyExists, want: resume.ErrStructuredMasterAlreadyExists},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: &fakeRegisterStore{structuredErr: tc.err}, NewID: sequenceIDs("version-1")})

			_, err := svc.ConfirmStructuredMaster(context.Background(), tc.in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestGetAndListResumeVersions(t *testing.T) {
	now := time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC)
	store := &fakeRegisterStore{
		versionOut: resumestore.VersionRecord{
			ID:                "version-1",
			UserID:            "user-1",
			ResumeAssetID:     "asset-1",
			VersionType:       sharedtypes.ResumeVersionTypeStructuredMaster,
			DisplayName:       "Structured master",
			StructuredProfile: json.RawMessage(`{"headline":"Senior engineer","provenance":{"promptVersion":"resume_profile.v1","rubricVersion":"not_applicable","modelId":"model-1","language":"en","featureFlag":"none","dataSourceVersion":"asset.v1"}}`),
			Provenance: resumestore.VersionProvenance{
				PromptVersion: "resume_profile.v1", RubricVersion: "not_applicable", ModelID: "model-1", Language: "en", FeatureFlag: "none", DataSourceVersion: "asset.v1",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		versionListOut: resumestore.VersionListResult{
			Items: []resumestore.VersionRecord{
				{ID: "version-2", UserID: "user-1", ResumeAssetID: "asset-1", VersionType: sharedtypes.ResumeVersionTypeTargeted, DisplayName: "Targeted", StructuredProfile: json.RawMessage(`{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`), CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)},
				{ID: "version-1", UserID: "user-1", ResumeAssetID: "asset-1", VersionType: sharedtypes.ResumeVersionTypeStructuredMaster, DisplayName: "Structured master", StructuredProfile: json.RawMessage(`{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`), CreatedAt: now, UpdatedAt: now},
			},
			NextCursor: "cursor-2",
			HasMore:    true,
			PageSize:   2,
		},
	}
	svc := resume.NewService(resume.ServiceOptions{Store: store})

	got, err := svc.GetResumeVersion(context.Background(), "user-1", "version-1")
	if err != nil {
		t.Fatalf("GetResumeVersion: %v", err)
	}
	if store.versionUserID != "user-1" || store.versionID != "version-1" {
		t.Fatalf("version get scope user=%q id=%q", store.versionUserID, store.versionID)
	}
	if got.Id != "version-1" || got.Provenance.PromptVersion != "resume_profile.v1" {
		t.Fatalf("version response = %+v", got)
	}

	list, err := svc.ListResumeVersions(context.Background(), resume.ListVersionRequest{UserID: "user-1", ResumeAssetID: "asset-1", Cursor: "cursor-1", PageSize: 2})
	if err != nil {
		t.Fatalf("ListResumeVersions: %v", err)
	}
	if store.versionListUserID != "user-1" || store.versionListAssetID != "asset-1" || store.versionListFilter.Cursor != "cursor-1" || store.versionListFilter.PageSize != 2 {
		t.Fatalf("version list scope user=%q asset=%q filter=%+v", store.versionListUserID, store.versionListAssetID, store.versionListFilter)
	}
	if len(list.Items) != 2 || list.PageInfo.NextCursor == nil || *list.PageInfo.NextCursor != "cursor-2" || !list.PageInfo.HasMore || list.PageInfo.PageSize != 2 {
		t.Fatalf("version list response = %+v", list)
	}
}

func TestResumeVersionReadMapsStoreErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(*resume.Service) error
		err  error
		want error
	}{
		{name: "get not found", err: resumestore.ErrVersionNotFound, want: resume.ErrNotFound, call: func(s *resume.Service) error {
			_, err := s.GetResumeVersion(context.Background(), "user-1", "missing")
			return err
		}},
		{name: "list asset not found", err: resumestore.ErrAssetNotFound, want: resume.ErrNotFound, call: func(s *resume.Service) error {
			_, err := s.ListResumeVersions(context.Background(), resume.ListVersionRequest{UserID: "user-1", ResumeAssetID: "missing"})
			return err
		}},
		{name: "list invalid cursor", err: resumestore.ErrInvalidCursor, want: resume.ErrInvalidCursor, call: func(s *resume.Service) error {
			_, err := s.ListResumeVersions(context.Background(), resume.ListVersionRequest{UserID: "user-1", ResumeAssetID: "asset-1", Cursor: "bad"})
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeRegisterStore{versionErr: tc.err, versionListErr: tc.err}
			svc := resume.NewService(resume.ServiceOptions{Store: store})
			err := tc.call(svc)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
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

	structuredIn  resumestore.CreateStructuredMasterInput
	structuredOut resumestore.VersionRecord
	structuredErr error

	versionUserID string
	versionID     string
	versionOut    resumestore.VersionRecord
	versionErr    error

	versionListUserID  string
	versionListAssetID string
	versionListFilter  resumestore.VersionListFilter
	versionListOut     resumestore.VersionListResult
	versionListErr     error
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

func (s *fakeRegisterStore) CreateStructuredMasterFromAsset(_ context.Context, in resumestore.CreateStructuredMasterInput) (resumestore.VersionRecord, error) {
	s.structuredIn = in
	return s.structuredOut, s.structuredErr
}

func (s *fakeRegisterStore) GetVersionByID(_ context.Context, userID string, versionID string) (resumestore.VersionRecord, error) {
	s.versionUserID = userID
	s.versionID = versionID
	return s.versionOut, s.versionErr
}

func (s *fakeRegisterStore) ListVersionsByAsset(_ context.Context, userID string, assetID string, filter resumestore.VersionListFilter) (resumestore.VersionListResult, error) {
	s.versionListUserID = userID
	s.versionListAssetID = assetID
	s.versionListFilter = filter
	return s.versionListOut, s.versionListErr
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

func validConfirmInput() resume.ConfirmStructuredMasterInput {
	return resume.ConfirmStructuredMasterInput{
		UserID:            "user-1",
		ResumeAssetID:     "asset-1",
		DisplayName:       "Structured master",
		StructuredProfile: validStructuredProfile(),
	}
}

func validStructuredProfile() map[string]any {
	return map[string]any{
		"headline": "Senior engineer",
		"provenance": map[string]any{
			"promptVersion":     "resume_profile.v1",
			"rubricVersion":     "not_applicable",
			"modelId":           "model-1",
			"language":          "en",
			"featureFlag":       "none",
			"dataSourceVersion": "asset.v1",
		},
	}
}
