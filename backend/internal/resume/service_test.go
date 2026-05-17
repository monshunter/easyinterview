package resume_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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

func TestUpdateResumeVersionSanitizesPatchAndMapsResponse(t *testing.T) {
	now := time.Date(2026, 5, 17, 19, 15, 0, 0, time.UTC)
	focusAngle := "Reliability leadership"
	matchScore := 0.82
	store := &fakeRegisterStore{updateVersionOut: resumestore.VersionRecord{
		ID:            "version-1",
		UserID:        "user-1",
		ResumeAssetID: "asset-1",
		VersionType:   sharedtypes.ResumeVersionTypeStructuredMaster,
		DisplayName:   "Updated master",
		FocusAngle:    &focusAngle,
		MatchScore:    &matchScore,
		StructuredProfile: json.RawMessage(`{
			"headline":"Senior engineer",
			"summary":"new summary",
			"skills":["Go","Postgres"],
			"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}
		}`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "p", RubricVersion: "r", ModelID: "m", Language: "en", FeatureFlag: "f", DataSourceVersion: "d",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }})
	displayName := " Updated master "
	inFocusAngle := " Reliability leadership "

	got, err := svc.UpdateResumeVersion(context.Background(), resume.UpdateVersionRequest{
		UserID:         " user-1 ",
		VersionID:      " version-1 ",
		DisplayName:    &displayName,
		DisplayNameSet: true,
		FocusAngle:     &inFocusAngle,
		FocusAngleSet:  true,
		MatchScore:     &matchScore,
		MatchScoreSet:  true,
		StructuredProfile: map[string]any{
			"summary": "new summary",
			"provenance": map[string]any{
				"promptVersion": "client-controlled",
			},
		},
		StructuredProfileSet: true,
	})
	if err != nil {
		t.Fatalf("UpdateResumeVersion: %v", err)
	}
	if store.updateVersionIn.UserID != "user-1" || store.updateVersionIn.VersionID != "version-1" {
		t.Fatalf("update scope = %+v", store.updateVersionIn)
	}
	if store.updateVersionIn.DisplayName == nil || *store.updateVersionIn.DisplayName != "Updated master" {
		t.Fatalf("displayName input = %#v", store.updateVersionIn.DisplayName)
	}
	if store.updateVersionIn.FocusAngle == nil || *store.updateVersionIn.FocusAngle != "Reliability leadership" {
		t.Fatalf("focusAngle input = %#v", store.updateVersionIn.FocusAngle)
	}
	if _, ok := store.updateVersionIn.StructuredProfilePatch["provenance"]; ok {
		t.Fatalf("client provenance leaked into store patch: %#v", store.updateVersionIn.StructuredProfilePatch)
	}
	if got.DisplayName != "Updated master" || got.FocusAngle == nil || *got.FocusAngle != focusAngle || got.MatchScore == nil || *got.MatchScore != matchScore {
		t.Fatalf("response = %+v", got)
	}
	profile, ok := got.StructuredProfile.(map[string]any)
	if !ok || profile["summary"] != "new summary" {
		t.Fatalf("structured profile response = %#v", got.StructuredProfile)
	}
	if got.Provenance.PromptVersion != "p" {
		t.Fatalf("provenance = %+v", got.Provenance)
	}
}

func TestUpdateResumeVersionValidationAndStoreErrors(t *testing.T) {
	displayName := "Updated"
	tests := []struct {
		name string
		in   resume.UpdateVersionRequest
		err  error
		want error
	}{
		{name: "empty patch", in: resume.UpdateVersionRequest{UserID: "user-1", VersionID: "version-1"}, want: resume.ErrValidationFailed},
		{name: "missing version", in: resume.UpdateVersionRequest{UserID: "user-1", DisplayName: &displayName, DisplayNameSet: true}, want: resume.ErrValidationFailed},
		{name: "not found", in: resume.UpdateVersionRequest{UserID: "user-1", VersionID: "version-1", DisplayName: &displayName, DisplayNameSet: true}, err: resumestore.ErrVersionNotFound, want: resume.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: &fakeRegisterStore{updateVersionErr: tc.err}})

			_, err := svc.UpdateResumeVersion(context.Background(), tc.in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestBranchResumeVersionRoutesSeedStrategies(t *testing.T) {
	now := time.Date(2026, 5, 17, 20, 15, 0, 0, time.UTC)
	focusAngle := "Platform evidence"
	tests := []struct {
		name       string
		strategy   sharedtypes.ResumeSeedStrategy
		ids        []string
		storeOut   resumestore.BranchVersionResult
		wantAsync  bool
		wantStatus int
	}{
		{
			name:       "copy master",
			strategy:   sharedtypes.ResumeSeedStrategyCopyMaster,
			ids:        []string{"version-copy"},
			storeOut:   branchStoreResult("version-copy", sharedtypes.ResumeSeedStrategyCopyMaster, now, false),
			wantStatus: http.StatusCreated,
		},
		{
			name:       "blank",
			strategy:   sharedtypes.ResumeSeedStrategyBlank,
			ids:        []string{"version-blank"},
			storeOut:   branchStoreResult("version-blank", sharedtypes.ResumeSeedStrategyBlank, now, false),
			wantStatus: http.StatusCreated,
		},
		{
			name:       "ai select",
			strategy:   sharedtypes.ResumeSeedStrategyAiSelect,
			ids:        []string{"version-ai", "tailor-run-1", "job-1"},
			storeOut:   branchStoreResult("version-ai", sharedtypes.ResumeSeedStrategyAiSelect, now, true),
			wantAsync:  true,
			wantStatus: http.StatusAccepted,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeRegisterStore{branchOut: tc.storeOut}
			svc := resume.NewService(resume.ServiceOptions{
				Store: store,
				Now:   func() time.Time { return now },
				NewID: sequenceIDs(tc.ids...),
			})

			got, err := svc.BranchResumeVersion(context.Background(), resume.BranchVersionRequest{
				UserID:          " user-1 ",
				ParentVersionID: " parent-1 ",
				TargetJobID:     " target-1 ",
				SeedStrategy:    tc.strategy,
				DisplayName:     " Targeted ",
				FocusAngle:      &focusAngle,
				IdempotencyKey:  " idem-branch ",
			})
			if err != nil {
				t.Fatalf("BranchResumeVersion: %v", err)
			}
			if store.branchIn.UserID != "user-1" || store.branchIn.ParentVersionID != "parent-1" || store.branchIn.TargetJobID != "target-1" || store.branchIn.DisplayName != "Targeted" {
				t.Fatalf("branch input = %+v", store.branchIn)
			}
			if store.branchIn.VersionID != tc.ids[0] || store.branchIn.SeedStrategy != tc.strategy || store.branchIn.Now != now {
				t.Fatalf("branch id/strategy/time = %+v", store.branchIn)
			}
			if store.branchIn.FocusAngle == nil || *store.branchIn.FocusAngle != focusAngle {
				t.Fatalf("focusAngle = %#v", store.branchIn.FocusAngle)
			}
			if store.branchIn.Provenance.PromptVersion == "" || store.branchIn.Provenance.DataSourceVersion == "" {
				t.Fatalf("provenance = %+v", store.branchIn.Provenance)
			}
			if got.Status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", got.Status, tc.wantStatus)
			}
			if tc.wantAsync {
				if store.branchIn.TailorRunID != "tailor-run-1" || store.branchIn.JobID != "job-1" {
					t.Fatalf("async ids = %+v", store.branchIn)
				}
				if got.Accepted == nil || got.Accepted.Job.JobType != "resume_tailor" || got.Accepted.Job.ResourceType != "resume_tailor_run" || got.Accepted.Job.Status != "queued" {
					t.Fatalf("accepted = %+v", got.Accepted)
				}
			} else if got.Accepted != nil || got.Version.Id != tc.ids[0] {
				t.Fatalf("sync result = %+v", got)
			}
		})
	}
}

func TestBranchResumeVersionValidationAndStoreErrors(t *testing.T) {
	tests := []struct {
		name string
		in   resume.BranchVersionRequest
		err  error
		want error
	}{
		{name: "missing display name", in: resume.BranchVersionRequest{UserID: "user-1", ParentVersionID: "parent-1", TargetJobID: "target-1", SeedStrategy: sharedtypes.ResumeSeedStrategyCopyMaster}, want: resume.ErrValidationFailed},
		{name: "invalid seed", in: validBranchInput("invalid"), want: resume.ErrValidationFailed},
		{name: "not found", in: validBranchInput(sharedtypes.ResumeSeedStrategyCopyMaster), err: resumestore.ErrVersionNotFound, want: resume.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: &fakeRegisterStore{branchErr: tc.err}, NewID: sequenceIDs("version-1", "tailor-run-1", "job-1")})

			_, err := svc.BranchResumeVersion(context.Background(), tc.in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestRequestResumeTailorCreatesQueuedRunAndJob(t *testing.T) {
	now := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
	store := &fakeRegisterStore{tailorCreateOut: resumestore.CreateTailorRunResult{
		TailorRunID:  "tailor-run-1",
		JobID:        "job-1",
		JobStatus:    sharedtypes.JobStatusQueued,
		JobCreatedAt: now,
		JobUpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("tailor-run-1", "job-1"),
	})

	got, err := svc.RequestResumeTailor(context.Background(), resume.RequestTailorRunInput{
		UserID:         " user-1 ",
		TargetJobID:    " target-1 ",
		ResumeAssetID:  " asset-1 ",
		Mode:           " gap_review ",
		IdempotencyKey: " idem-tailor ",
	})
	if err != nil {
		t.Fatalf("RequestResumeTailor: %v", err)
	}
	if store.tailorCreateIn.UserID != "user-1" || store.tailorCreateIn.TargetJobID != "target-1" || store.tailorCreateIn.ResumeAssetID != "asset-1" || store.tailorCreateIn.Mode != "gap_review" {
		t.Fatalf("tailor create input = %+v", store.tailorCreateIn)
	}
	if store.tailorCreateIn.TailorRunID != "tailor-run-1" || store.tailorCreateIn.JobID != "job-1" || store.tailorCreateIn.DedupeKey == "" || store.tailorCreateIn.Now != now {
		t.Fatalf("tailor generated fields = %+v", store.tailorCreateIn)
	}
	if got.TailorRunId != "tailor-run-1" || got.Job.Id != "job-1" || got.Job.JobType != "resume_tailor" || got.Job.ResourceType != "resume_tailor_run" || got.Job.Status != "queued" {
		t.Fatalf("response = %+v", got)
	}
}

func TestRequestResumeTailorValidationAndStoreErrors(t *testing.T) {
	tests := []struct {
		name string
		in   resume.RequestTailorRunInput
		err  error
		want error
	}{
		{name: "missing asset", in: resume.RequestTailorRunInput{UserID: "user-1", TargetJobID: "target-1", Mode: "gap_review", IdempotencyKey: "idem"}, want: resume.ErrValidationFailed},
		{name: "invalid mode", in: resume.RequestTailorRunInput{UserID: "user-1", TargetJobID: "target-1", ResumeAssetID: "asset-1", Mode: "unsupported", IdempotencyKey: "idem"}, want: resume.ErrValidationFailed},
		{name: "not found", in: validTailorInput(), err: resumestore.ErrAssetNotFound, want: resume.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: &fakeRegisterStore{tailorCreateErr: tc.err}, NewID: sequenceIDs("tailor-run-1", "job-1")})

			_, err := svc.RequestResumeTailor(context.Background(), tc.in)

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestGetResumeTailorRunMapsStatusesAndErrors(t *testing.T) {
	now := time.Date(2026, 5, 18, 10, 15, 0, 0, time.UTC)
	store := &fakeRegisterStore{tailorGetOut: resumestore.TailorRunRecord{
		ID:            "tailor-run-1",
		UserID:        "user-1",
		TargetJobID:   "target-1",
		ResumeAssetID: "asset-1",
		Status:        "ready",
		MatchSummary:  json.RawMessage(`{"strengths":["Strong systems evidence"],"gaps":["Add edge runtime detail"]}`),
		Suggestions:   json.RawMessage(`[{"originalBullet":"Led migration.","suggestedBullet":"Led migration across 12 teams.","reason":"Adds scope."}]`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "resume_tailor.v2", RubricVersion: "not_applicable", ModelID: "model-profile:contract.default", Language: "zh-CN", FeatureFlag: "none", DataSourceVersion: "target_job.v17",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}}
	svc := resume.NewService(resume.ServiceOptions{Store: store})

	got, err := svc.GetResumeTailorRun(context.Background(), " user-1 ", " tailor-run-1 ")
	if err != nil {
		t.Fatalf("GetResumeTailorRun: %v", err)
	}
	if store.tailorGetUserID != "user-1" || store.tailorGetID != "tailor-run-1" {
		t.Fatalf("get scope user=%q run=%q", store.tailorGetUserID, store.tailorGetID)
	}
	if got.Status != "ready" || got.MatchSummary == nil || len(got.Suggestions) != 1 || got.Provenance == nil || got.Provenance.PromptVersion != "resume_tailor.v2" {
		t.Fatalf("run response = %+v", got)
	}

	store.tailorGetErr = resumestore.ErrTailorRunNotFound
	if _, err := svc.GetResumeTailorRun(context.Background(), "user-1", "missing"); !errors.Is(err, resume.ErrNotFound) {
		t.Fatalf("not found err = %v, want ErrNotFound", err)
	}
}

func TestResumeSuggestionDecisionRoutesAcceptRejectToStore(t *testing.T) {
	now := time.Date(2026, 5, 18, 11, 30, 0, 0, time.UTC)
	for _, tc := range []struct {
		name         string
		call         func(*resume.Service) (string, error)
		wantDecision sharedtypes.ResumeTailorSuggestionStatus
	}{
		{
			name: "accept",
			call: func(s *resume.Service) (string, error) {
				got, err := s.AcceptResumeTailorSuggestion(context.Background(), validSuggestionDecisionInput())
				return suggestionStatusFromVersion(t, got.Suggestions), err
			},
			wantDecision: sharedtypes.ResumeTailorSuggestionStatusAccepted,
		},
		{
			name: "reject",
			call: func(s *resume.Service) (string, error) {
				got, err := s.RejectResumeTailorSuggestion(context.Background(), validSuggestionDecisionInput())
				return suggestionStatusFromVersion(t, got.Suggestions), err
			},
			wantDecision: sharedtypes.ResumeTailorSuggestionStatusRejected,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeRegisterStore{decideSuggestionOut: suggestionDecisionStoreVersion("version-1", "suggestion-1", string(tc.wantDecision), now)}
			svc := resume.NewService(resume.ServiceOptions{Store: store, Now: func() time.Time { return now }})

			status, err := tc.call(svc)
			if err != nil {
				t.Fatalf("%s suggestion: %v", tc.name, err)
			}
			if store.decideSuggestionIn.UserID != "user-1" || store.decideSuggestionIn.ResumeVersionID != "version-1" || store.decideSuggestionIn.SuggestionID != "suggestion-1" {
				t.Fatalf("store scope = %+v", store.decideSuggestionIn)
			}
			if store.decideSuggestionIn.Decision != tc.wantDecision || store.decideSuggestionIn.Now != now {
				t.Fatalf("store decision = %+v", store.decideSuggestionIn)
			}
			if status != string(tc.wantDecision) {
				t.Fatalf("response suggestion status = %q, want %q", status, tc.wantDecision)
			}
		})
	}
}

func TestResumeSuggestionDecisionValidationAndStoreErrors(t *testing.T) {
	tests := []struct {
		name string
		in   resume.SuggestionDecisionRequest
		err  error
		want error
	}{
		{name: "missing version", in: resume.SuggestionDecisionRequest{UserID: "user-1", SuggestionID: "suggestion-1", IdempotencyKey: "idem"}, want: resume.ErrValidationFailed},
		{name: "missing idempotency key", in: resume.SuggestionDecisionRequest{UserID: "user-1", ResumeVersionID: "version-1", SuggestionID: "suggestion-1"}, want: resume.ErrValidationFailed},
		{name: "not found", in: validSuggestionDecisionInput(), err: resumestore.ErrSuggestionNotFound, want: resume.ErrNotFound},
		{name: "already decided", in: validSuggestionDecisionInput(), err: resumestore.ErrSuggestionAlreadyDecided, want: resume.ErrSuggestionAlreadyDecided},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := resume.NewService(resume.ServiceOptions{Store: &fakeRegisterStore{decideSuggestionErr: tc.err}})

			_, err := svc.AcceptResumeTailorSuggestion(context.Background(), tc.in)

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

	updateVersionIn  resumestore.VersionUpdateInput
	updateVersionOut resumestore.VersionRecord
	updateVersionErr error

	branchIn  resumestore.BranchVersionInput
	branchOut resumestore.BranchVersionResult
	branchErr error

	tailorCreateIn  resumestore.CreateTailorRunInput
	tailorCreateOut resumestore.CreateTailorRunResult
	tailorCreateErr error

	tailorGetUserID string
	tailorGetID     string
	tailorGetOut    resumestore.TailorRunRecord
	tailorGetErr    error

	decideSuggestionIn  resumestore.DecideSuggestionInput
	decideSuggestionOut resumestore.VersionRecord
	decideSuggestionErr error
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

func (s *fakeRegisterStore) UpdateVersionPatch(_ context.Context, in resumestore.VersionUpdateInput) (resumestore.VersionRecord, error) {
	s.updateVersionIn = in
	return s.updateVersionOut, s.updateVersionErr
}

func (s *fakeRegisterStore) BranchFromParent(_ context.Context, in resumestore.BranchVersionInput) (resumestore.BranchVersionResult, error) {
	s.branchIn = in
	return s.branchOut, s.branchErr
}

func (s *fakeRegisterStore) CreateTailorRun(_ context.Context, in resumestore.CreateTailorRunInput) (resumestore.CreateTailorRunResult, error) {
	s.tailorCreateIn = in
	return s.tailorCreateOut, s.tailorCreateErr
}

func (s *fakeRegisterStore) GetTailorRun(_ context.Context, userID string, tailorRunID string) (resumestore.TailorRunRecord, error) {
	s.tailorGetUserID = userID
	s.tailorGetID = tailorRunID
	return s.tailorGetOut, s.tailorGetErr
}

func (s *fakeRegisterStore) MarkTailorRunGenerating(context.Context, resumestore.TailorRunStatusInput) (resumestore.TailorRunRecord, error) {
	return resumestore.TailorRunRecord{}, errors.New("not implemented")
}

func (s *fakeRegisterStore) MarkTailorRunReady(context.Context, resumestore.TailorRunReadyInput) (resumestore.TailorRunRecord, error) {
	return resumestore.TailorRunRecord{}, errors.New("not implemented")
}

func (s *fakeRegisterStore) MarkTailorRunFailed(context.Context, resumestore.TailorRunFailureInput) (resumestore.TailorRunRecord, error) {
	return resumestore.TailorRunRecord{}, errors.New("not implemented")
}

func (s *fakeRegisterStore) DecideResumeSuggestion(_ context.Context, in resumestore.DecideSuggestionInput) (resumestore.VersionRecord, error) {
	s.decideSuggestionIn = in
	return s.decideSuggestionOut, s.decideSuggestionErr
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

func validTailorInput() resume.RequestTailorRunInput {
	return resume.RequestTailorRunInput{
		UserID:         "user-1",
		TargetJobID:    "target-1",
		ResumeAssetID:  "asset-1",
		Mode:           "gap_review",
		IdempotencyKey: "idem-tailor",
	}
}

func validSuggestionDecisionInput() resume.SuggestionDecisionRequest {
	return resume.SuggestionDecisionRequest{
		UserID:          " user-1 ",
		ResumeVersionID: " version-1 ",
		SuggestionID:    " suggestion-1 ",
		IdempotencyKey:  " idem-suggestion ",
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

func suggestionDecisionStoreVersion(versionID, suggestionID, status string, now time.Time) resumestore.VersionRecord {
	return resumestore.VersionRecord{
		ID:                versionID,
		UserID:            "user-1",
		ResumeAssetID:     "asset-1",
		VersionType:       sharedtypes.ResumeVersionTypeTargeted,
		DisplayName:       "Targeted",
		StructuredProfile: json.RawMessage(`{"headline":"Senior engineer","provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "p", RubricVersion: "r", ModelID: "m", Language: "en", FeatureFlag: "f", DataSourceVersion: "d",
		},
		Suggestions: []any{map[string]any{
			"id":              suggestionID,
			"originalBullet":  "Improved reliability.",
			"suggestedBullet": "Improved reliability with release guardrails.",
			"status":          status,
			"provenance": map[string]any{
				"promptVersion": "p", "rubricVersion": "r", "modelId": "m", "language": "en", "featureFlag": "f", "dataSourceVersion": "d",
			},
			"createdAt": now.Format(time.RFC3339),
			"decidedAt": now.Format(time.RFC3339),
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func suggestionStatusFromVersion(t *testing.T, suggestions []any) string {
	t.Helper()
	if len(suggestions) != 1 {
		t.Fatalf("suggestions = %#v", suggestions)
	}
	first, ok := suggestions[0].(map[string]any)
	if !ok {
		t.Fatalf("suggestion type = %T", suggestions[0])
	}
	status, _ := first["status"].(string)
	if status == "" {
		t.Fatalf("suggestion missing status: %+v", first)
	}
	return status
}

func validBranchInput(strategy sharedtypes.ResumeSeedStrategy) resume.BranchVersionRequest {
	return resume.BranchVersionRequest{
		UserID:          "user-1",
		ParentVersionID: "parent-1",
		TargetJobID:     "target-1",
		SeedStrategy:    strategy,
		DisplayName:     "Targeted",
		IdempotencyKey:  "idem-branch",
	}
}

func branchStoreResult(versionID string, strategy sharedtypes.ResumeSeedStrategy, now time.Time, async bool) resumestore.BranchVersionResult {
	parentID := "parent-1"
	targetID := "target-1"
	focusAngle := "Platform evidence"
	result := resumestore.BranchVersionResult{
		Version: resumestore.VersionRecord{
			ID:                versionID,
			UserID:            "user-1",
			ResumeAssetID:     "asset-1",
			ParentVersionID:   &parentID,
			VersionType:       sharedtypes.ResumeVersionTypeTargeted,
			TargetJobID:       &targetID,
			DisplayName:       "Targeted",
			SeedStrategy:      &strategy,
			FocusAngle:        &focusAngle,
			StructuredProfile: json.RawMessage(`{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`),
			Provenance: resumestore.VersionProvenance{
				PromptVersion: "p", RubricVersion: "r", ModelID: "m", Language: "en", FeatureFlag: "f", DataSourceVersion: "d",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	if async {
		result.TailorRunID = "tailor-run-1"
		result.JobID = "job-1"
		result.JobStatus = sharedtypes.JobStatusQueued
		result.JobCreatedAt = now
		result.JobUpdatedAt = now
	}
	return result
}
