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
	store := &fakeStore{out: resumestore.CreateAssetResult{AssetID: "resume-1", JobID: "job-1", JobStatus: sharedtypes.JobStatusQueued}}
	upload := &fakeUploadRegistrar{out: uploadstore.FileObject{ID: "file-1"}}
	svc := resume.NewService(resume.ServiceOptions{
		Store:          store,
		UploadRegister: upload,
		Now:            func() time.Time { return time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC) },
		NewID:          sequenceIDs("resume-1", "job-1"),
	})

	out, err := svc.RegisterResume(context.Background(), resume.RegisterInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		SourceType:     "upload",
		FileObjectID:   "file-input",
		Title:          "Resume",
		Language:       "en",
	})
	if err != nil {
		t.Fatalf("RegisterResume: %v", err)
	}
	if out.ResumeId != "resume-1" || out.Job.JobType != "resume_parse" {
		t.Fatalf("response = %+v", out)
	}
	if upload.in.ExpectedPurpose != uploadstore.PurposeResume || upload.in.OwnerUserID != "user-1" {
		t.Fatalf("upload register input = %+v", upload.in)
	}
	if store.createIn.FileObjectID == nil || *store.createIn.FileObjectID != "file-1" {
		t.Fatalf("store input file object = %+v", store.createIn.FileObjectID)
	}
}

func TestRegisterPasteRejectsBlankIdempotencyKey(t *testing.T) {
	svc := resume.NewService(resume.ServiceOptions{Store: &fakeStore{}, NewID: sequenceIDs("resume-1", "job-1")})
	_, err := svc.RegisterResume(context.Background(), resume.RegisterInput{
		UserID:     "user-1",
		SourceType: "paste",
		RawText:    "resume text",
		Title:      "Resume",
		Language:   "en",
	})
	if !errors.Is(err, resume.ErrValidationFailed) {
		t.Fatalf("err = %v, want ErrValidationFailed", err)
	}
}

func TestRegisterUploadPathRejectsMissingObjectBeforeCreate(t *testing.T) {
	store := &fakeStore{}
	upload := &fakeUploadRegistrar{err: uploadservice.ErrValidationFailed}
	svc := resume.NewService(resume.ServiceOptions{
		Store:          store,
		UploadRegister: upload,
		NewID:          sequenceIDs("resume-1", "job-1"),
	})

	_, err := svc.RegisterResume(context.Background(), resume.RegisterInput{
		UserID:         "user-1",
		IdempotencyKey: "idem-1",
		SourceType:     "upload",
		FileObjectID:   "missing",
		Title:          "Resume",
		Language:       "en",
	})
	if !errors.Is(err, resume.ErrValidationFailed) {
		t.Fatalf("err = %v, want ErrValidationFailed", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("store create calls = %d, want 0", store.createCalls)
	}
}

func TestGetAndListResumesMapStoreRecordsWithUserScope(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	displayName := "Alice CV"
	rec := resumestore.ResumeRecord{
		ID:                "resume-1",
		UserID:            "user-1",
		Title:             "Resume",
		DisplayName:       &displayName,
		Language:          "en",
		ParseStatus:       sharedtypes.TargetJobParseStatusReady,
		ParsedSummary:     json.RawMessage(`{"headline":"Senior engineer"}`),
		StructuredProfile: json.RawMessage(`{"basics":{"name":"Alice"}}`),
		SourceType:        ptr("paste"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	store := &fakeStore{
		getOut:  rec,
		listOut: resumestore.ListResult{Items: []resumestore.ResumeRecord{rec}, NextCursor: "cursor-2", HasMore: true, PageSize: 20},
	}
	svc := resume.NewService(resume.ServiceOptions{Store: store})

	got, err := svc.GetResume(context.Background(), "user-1", "resume-1")
	if err != nil {
		t.Fatalf("GetResume: %v", err)
	}
	if got.Id != "resume-1" || got.DisplayName != "Alice CV" || got.StructuredProfile == nil {
		t.Fatalf("GetResume mapped = %+v", got)
	}
	if store.getUserID != "user-1" || store.getResumeID != "resume-1" {
		t.Fatalf("store get scope user=%q resume=%q", store.getUserID, store.getResumeID)
	}

	list, err := svc.ListResumes(context.Background(), resume.ListRequest{UserID: "user-1", Cursor: "cursor-1", PageSize: 20})
	if err != nil {
		t.Fatalf("ListResumes: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Id != "resume-1" || list.PageInfo.NextCursor == nil || *list.PageInfo.NextCursor != "cursor-2" || !list.PageInfo.HasMore {
		t.Fatalf("ListResumes mapped = %+v", list)
	}
	if store.listUserID != "user-1" || store.listFilter.Cursor != "cursor-1" || store.listFilter.PageSize != 20 {
		t.Fatalf("store list scope = %q filter=%+v", store.listUserID, store.listFilter)
	}
}

func TestGetResumeMapsStoreNotFound(t *testing.T) {
	store := &fakeStore{getErr: resumestore.ErrAssetNotFound}
	svc := resume.NewService(resume.ServiceOptions{Store: store})
	_, err := svc.GetResume(context.Background(), "user-1", "missing")
	if !errors.Is(err, resume.ErrNotFound) {
		t.Fatalf("GetResume err = %v, want ErrNotFound", err)
	}
}

func TestUpdateResumeOverwritesAndStripsProvenance(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{updateOut: resumestore.ResumeRecord{
		ID:                "resume-1",
		UserID:            "user-1",
		Title:             "Resume",
		Language:          "en",
		ParseStatus:       sharedtypes.TargetJobParseStatusReady,
		StructuredProfile: json.RawMessage(`{"headline":"Staff engineer"}`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }})

	displayName := "Updated CV"
	got, err := svc.UpdateResume(context.Background(), resume.UpdateResumeRequest{
		UserID:         "user-1",
		ResumeID:       "resume-1",
		DisplayName:    &displayName,
		DisplayNameSet: true,
		StructuredProfile: map[string]any{
			"headline":   "Staff engineer",
			"provenance": map[string]any{"promptVersion": "p"},
		},
		StructuredProfileSet: true,
	})
	if err != nil {
		t.Fatalf("UpdateResume: %v", err)
	}
	if got.Id != "resume-1" {
		t.Fatalf("UpdateResume mapped = %+v", got)
	}
	if store.updateIn.DisplayName == nil || *store.updateIn.DisplayName != "Updated CV" {
		t.Fatalf("update displayName = %#v", store.updateIn.DisplayName)
	}
	var profile map[string]any
	if err := json.Unmarshal(store.updateIn.StructuredProfile, &profile); err != nil {
		t.Fatalf("structured profile = %s (err=%v)", store.updateIn.StructuredProfile, err)
	}
	if _, ok := profile["provenance"]; ok {
		t.Fatalf("provenance must be stripped: %+v", profile)
	}
}

func TestUpdateResumeDisplayNameOnlyPreservesStructuredProfile(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	displayNameOut := "Renamed CV"
	store := &fakeStore{updateOut: resumestore.ResumeRecord{
		ID:                "resume-1",
		UserID:            "user-1",
		Title:             "Resume",
		DisplayName:       &displayNameOut,
		Language:          "en",
		ParseStatus:       sharedtypes.TargetJobParseStatusReady,
		StructuredProfile: json.RawMessage(`{"headline":"Existing profile"}`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }})

	displayName := "Renamed CV"
	got, err := svc.UpdateResume(context.Background(), resume.UpdateResumeRequest{
		UserID:         "user-1",
		ResumeID:       "resume-1",
		DisplayName:    &displayName,
		DisplayNameSet: true,
	})
	if err != nil {
		t.Fatalf("UpdateResume: %v", err)
	}
	if got.StructuredProfile == nil {
		t.Fatalf("mapped structured profile is nil")
	}
	if len(store.updateIn.StructuredProfile) != 0 {
		t.Fatalf("displayName-only PATCH must not overwrite structuredProfile, got %s", store.updateIn.StructuredProfile)
	}
	if store.updateIn.DisplayName == nil || *store.updateIn.DisplayName != "Renamed CV" {
		t.Fatalf("update displayName = %#v", store.updateIn.DisplayName)
	}
}

func TestUpdateResumeValidationAndStoreErrors(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		store *fakeStore
		in    resume.UpdateResumeRequest
		want  error
	}{
		{name: "missing resume id", store: &fakeStore{}, in: resume.UpdateResumeRequest{UserID: "user-1", StructuredProfile: map[string]any{"x": 1}, StructuredProfileSet: true}, want: resume.ErrValidationFailed},
		{name: "no editable fields", store: &fakeStore{}, in: resume.UpdateResumeRequest{UserID: "user-1", ResumeID: "resume-1"}, want: resume.ErrValidationFailed},
		{name: "store not found", store: &fakeStore{updateErr: resumestore.ErrAssetNotFound}, in: resume.UpdateResumeRequest{UserID: "user-1", ResumeID: "resume-1", StructuredProfile: map[string]any{"x": 1}, StructuredProfileSet: true}, want: resume.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: tc.store, Now: func() time.Time { return now }})
			_, err := svc.UpdateResume(context.Background(), tc.in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDuplicateResumeAllocatesNewIDAndAppliesProfile(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{duplicateOut: resumestore.ResumeRecord{
		ID:                "resume-new",
		UserID:            "user-1",
		Title:             "Resume",
		Language:          "en",
		ParseStatus:       sharedtypes.TargetJobParseStatusReady,
		StructuredProfile: json.RawMessage(`{"headline":"new"}`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }, NewID: sequenceIDs("resume-new")})

	got, err := svc.DuplicateResume(context.Background(), resume.DuplicateResumeRequest{
		UserID:               "user-1",
		SourceResumeID:       "source-1",
		StructuredProfile:    map[string]any{"headline": "new"},
		StructuredProfileSet: true,
	})
	if err != nil {
		t.Fatalf("DuplicateResume: %v", err)
	}
	if got.Id != "resume-new" {
		t.Fatalf("DuplicateResume mapped = %+v", got)
	}
	if store.duplicateIn.NewResumeID != "resume-new" || store.duplicateIn.SourceResumeID != "source-1" {
		t.Fatalf("duplicate input = %+v", store.duplicateIn)
	}
	if !store.duplicateIn.StructuredProfileSet || string(store.duplicateIn.StructuredProfile) != `{"headline":"new"}` {
		t.Fatalf("structured profile input = set:%v raw:%s", store.duplicateIn.StructuredProfileSet, store.duplicateIn.StructuredProfile)
	}
}

func TestDuplicateResumePreservesExplicitEmptyProfile(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{duplicateOut: resumestore.ResumeRecord{
		ID:                "resume-new",
		UserID:            "user-1",
		Title:             "Resume",
		Language:          "en",
		ParseStatus:       sharedtypes.TargetJobParseStatusReady,
		StructuredProfile: json.RawMessage(`{}`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }, NewID: sequenceIDs("resume-new")})

	got, err := svc.DuplicateResume(context.Background(), resume.DuplicateResumeRequest{
		UserID:               "user-1",
		SourceResumeID:       "source-1",
		StructuredProfile:    map[string]any{},
		StructuredProfileSet: true,
	})
	if err != nil {
		t.Fatalf("DuplicateResume: %v", err)
	}
	if got.Id != "resume-new" {
		t.Fatalf("DuplicateResume mapped = %+v", got)
	}
	if !store.duplicateIn.StructuredProfileSet || string(store.duplicateIn.StructuredProfile) != `{}` {
		t.Fatalf("structured profile input = set:%v raw:%s", store.duplicateIn.StructuredProfileSet, store.duplicateIn.StructuredProfile)
	}
}

func TestDuplicateResumeValidationAndStoreErrors(t *testing.T) {
	tests := []struct {
		name  string
		store *fakeStore
		in    resume.DuplicateResumeRequest
		want  error
	}{
		{name: "missing source", store: &fakeStore{}, in: resume.DuplicateResumeRequest{UserID: "user-1"}, want: resume.ErrValidationFailed},
		{name: "store not found", store: &fakeStore{duplicateErr: resumestore.ErrAssetNotFound}, in: resume.DuplicateResumeRequest{UserID: "user-1", SourceResumeID: "source-1"}, want: resume.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: tc.store, NewID: sequenceIDs("resume-new")})
			_, err := svc.DuplicateResume(context.Background(), tc.in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestArchiveResumeReturnsArchivedStatusAndScopesUser(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{archiveOut: resumestore.ResumeRecord{
		ID:          "resume-1",
		UserID:      "user-1",
		Title:       "Resume",
		Language:    "en",
		ParseStatus: sharedtypes.TargetJobParseStatusReady,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   &now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }})

	got, err := svc.ArchiveResume(context.Background(), "user-1", "resume-1")
	if err != nil {
		t.Fatalf("ArchiveResume: %v", err)
	}
	if got.Status == nil || *got.Status != "archived" {
		t.Fatalf("status = %#v, want archived", got.Status)
	}
	if got.DeletedAt == nil || *got.DeletedAt != now.Format(time.RFC3339) {
		t.Fatalf("deletedAt = %#v, want archive timestamp", got.DeletedAt)
	}
	if store.archiveIn.UserID != "user-1" || store.archiveIn.ResumeID != "resume-1" || !store.archiveIn.Now.Equal(now) {
		t.Fatalf("archive input = %+v", store.archiveIn)
	}

	store.archiveErr = resumestore.ErrAssetNotFound
	if _, err := svc.ArchiveResume(context.Background(), "user-1", "missing"); !errors.Is(err, resume.ErrNotFound) {
		t.Fatalf("ArchiveResume missing err = %v, want ErrNotFound", err)
	}

	store.archiveErr = resumestore.ErrAlreadyArchived
	if _, err := svc.ArchiveResume(context.Background(), "user-1", "resume-1"); !errors.Is(err, resume.ErrAlreadyArchived) {
		t.Fatalf("ArchiveResume archived err = %v, want ErrAlreadyArchived", err)
	}
}

func TestRequestResumeTailorCreatesQueuedRunAndJob(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{tailorCreateOut: resumestore.CreateTailorRunResult{
		TailorRunID:  "tailor-1",
		JobID:        "job-1",
		JobStatus:    sharedtypes.JobStatusQueued,
		JobCreatedAt: now,
		JobUpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }, NewID: sequenceIDs("tailor-1", "job-1")})

	got, err := svc.RequestResumeTailor(context.Background(), resume.RequestTailorRunInput{
		UserID:         "user-1",
		TargetJobID:    "target-1",
		ResumeID:       "resume-1",
		Mode:           "gap_review",
		IdempotencyKey: "idem-tailor",
	})
	if err != nil {
		t.Fatalf("RequestResumeTailor: %v", err)
	}
	if got.TailorRunId != "tailor-1" || got.Job.JobType != "resume_tailor" || got.Job.ResourceType != "resume_tailor_run" {
		t.Fatalf("response = %+v", got)
	}
	if store.tailorCreateIn.ResumeID != "resume-1" || store.tailorCreateIn.TargetJobID != "target-1" || store.tailorCreateIn.Mode != "gap_review" {
		t.Fatalf("tailor create input = %+v", store.tailorCreateIn)
	}
}

func TestRequestResumeTailorValidationAndStoreErrors(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		store *fakeStore
		in    resume.RequestTailorRunInput
		want  error
	}{
		{name: "missing resume id", store: &fakeStore{}, in: resume.RequestTailorRunInput{UserID: "user-1", Mode: "gap_review", IdempotencyKey: "idem"}, want: resume.ErrValidationFailed},
		{name: "invalid mode", store: &fakeStore{}, in: resume.RequestTailorRunInput{UserID: "user-1", ResumeID: "resume-1", Mode: "rewrite", IdempotencyKey: "idem"}, want: resume.ErrValidationFailed},
		{name: "store not found", store: &fakeStore{tailorCreateErr: resumestore.ErrAssetNotFound}, in: resume.RequestTailorRunInput{UserID: "user-1", ResumeID: "resume-1", Mode: "gap_review", IdempotencyKey: "idem"}, want: resume.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: tc.store, Now: func() time.Time { return now }, NewID: sequenceIDs("tailor-1", "job-1")})
			_, err := svc.RequestResumeTailor(context.Background(), tc.in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestGetResumeTailorRunMapsStatusesAndErrors(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{tailorGetOut: resumestore.TailorRunRecord{
		ID:           "tailor-1",
		UserID:       "user-1",
		ResumeID:     "resume-1",
		TargetJobID:  "target-1",
		Mode:         "gap_review",
		Status:       "ready",
		MatchSummary: json.RawMessage(`{"strengths":["Go"],"gaps":["k8s"]}`),
		Suggestions:  json.RawMessage(`[{"originalBullet":"a","suggestedBullet":"b","reason":"impact"}]`),
		Provenance:   resumestore.VersionProvenance{PromptVersion: "p", ModelID: "m", Language: "en"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store})

	got, err := svc.GetResumeTailorRun(context.Background(), " user-1 ", " tailor-1 ")
	if err != nil {
		t.Fatalf("GetResumeTailorRun: %v", err)
	}
	if got.Id != "tailor-1" || got.ResumeId != "resume-1" || got.Status != "ready" {
		t.Fatalf("tailor run = %+v", got)
	}
	if got.TargetJobId == nil || *got.TargetJobId != "target-1" {
		t.Fatalf("target job id = %#v", got.TargetJobId)
	}
	if got.MatchSummary == nil || len(got.Suggestions) != 1 || got.Provenance == nil {
		t.Fatalf("ready run payload = %+v", got)
	}
	if store.tailorGetUserID != "user-1" || store.tailorGetID != "tailor-1" {
		t.Fatalf("tailor get scope user=%q id=%q", store.tailorGetUserID, store.tailorGetID)
	}

	store.tailorGetErr = resumestore.ErrTailorRunNotFound
	if _, err := svc.GetResumeTailorRun(context.Background(), "user-1", "missing"); !errors.Is(err, resume.ErrNotFound) {
		t.Fatalf("GetResumeTailorRun missing err = %v, want ErrNotFound", err)
	}
}

func TestGetResumeTailorRunGeneratingOmitsReadyFields(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{tailorGetOut: resumestore.TailorRunRecord{
		ID:        "tailor-1",
		UserID:    "user-1",
		ResumeID:  "resume-1",
		Mode:      "bullet_suggestions",
		Status:    "generating",
		CreatedAt: now,
		UpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store})

	got, err := svc.GetResumeTailorRun(context.Background(), "user-1", "tailor-1")
	if err != nil {
		t.Fatalf("GetResumeTailorRun: %v", err)
	}
	if got.Status != "generating" || got.MatchSummary != nil || got.Provenance != nil {
		t.Fatalf("generating run = %+v", got)
	}
	if got.Suggestions == nil || len(got.Suggestions) != 0 {
		t.Fatalf("suggestions = %#v, want empty slice", got.Suggestions)
	}
	if got.TargetJobId != nil {
		t.Fatalf("target job id = %#v, want nil", got.TargetJobId)
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

type fakeStore struct {
	createCalls int
	createIn    resumestore.CreateAssetInput
	out         resumestore.CreateAssetResult
	err         error

	getUserID   string
	getResumeID string
	getOut      resumestore.ResumeRecord
	getErr      error

	listUserID string
	listFilter resumestore.ListFilter
	listOut    resumestore.ListResult
	listErr    error

	updateIn  resumestore.UpdateResumeInput
	updateOut resumestore.ResumeRecord
	updateErr error

	duplicateIn  resumestore.DuplicateResumeInput
	duplicateOut resumestore.ResumeRecord
	duplicateErr error

	archiveIn  resumestore.ArchiveResumeInput
	archiveOut resumestore.ResumeRecord
	archiveErr error

	tailorCreateIn  resumestore.CreateTailorRunInput
	tailorCreateOut resumestore.CreateTailorRunResult
	tailorCreateErr error

	tailorGetUserID string
	tailorGetID     string
	tailorGetOut    resumestore.TailorRunRecord
	tailorGetErr    error
}

func (s *fakeStore) CreateWithParseJob(_ context.Context, in resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error) {
	s.createCalls++
	s.createIn = in
	return s.out, s.err
}

func (s *fakeStore) Get(_ context.Context, userID string, resumeID string) (resumestore.ResumeRecord, error) {
	s.getUserID = userID
	s.getResumeID = resumeID
	return s.getOut, s.getErr
}

func (s *fakeStore) List(_ context.Context, userID string, filter resumestore.ListFilter) (resumestore.ListResult, error) {
	s.listUserID = userID
	s.listFilter = filter
	return s.listOut, s.listErr
}

func (s *fakeStore) UpdateResume(_ context.Context, in resumestore.UpdateResumeInput) (resumestore.ResumeRecord, error) {
	s.updateIn = in
	return s.updateOut, s.updateErr
}

func (s *fakeStore) DuplicateResume(_ context.Context, in resumestore.DuplicateResumeInput) (resumestore.ResumeRecord, error) {
	s.duplicateIn = in
	return s.duplicateOut, s.duplicateErr
}

func (s *fakeStore) ArchiveResume(_ context.Context, in resumestore.ArchiveResumeInput) (resumestore.ResumeRecord, error) {
	s.archiveIn = in
	return s.archiveOut, s.archiveErr
}

func (s *fakeStore) CreateTailorRun(_ context.Context, in resumestore.CreateTailorRunInput) (resumestore.CreateTailorRunResult, error) {
	s.tailorCreateIn = in
	return s.tailorCreateOut, s.tailorCreateErr
}

func (s *fakeStore) GetTailorRun(_ context.Context, userID string, tailorRunID string) (resumestore.TailorRunRecord, error) {
	s.tailorGetUserID = userID
	s.tailorGetID = tailorRunID
	return s.tailorGetOut, s.tailorGetErr
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

func ptr(in string) *string {
	return &in
}
