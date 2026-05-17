//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	storeai "github.com/monshunter/easyinterview/backend/internal/store/ai"
)

func TestStructuredMasterUniqueCrossUserReadinessAndSoftDelete(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureStructuredMasterUniqueIndex(t, ctx, db)

	repo := resumestore.NewRepository(db)
	now := time.Date(2026, 5, 17, 17, 0, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf2a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf2a002"
	assetReady := "0195f2d0-4a44-7fc2-8f77-1f9c4cf2a003"
	assetProcessing := "0195f2d0-4a44-7fc2-8f77-1f9c4cf2a004"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-version-a@example.com', 'active'), ($2, 'resume-version-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status) values ($1, $2, 'Ready Resume', 'en', 'ready'), ($3, $2, 'Processing Resume', 'en', 'processing')`, assetReady, userA, assetProcessing)

	if _, err := repo.CreateStructuredMasterFromAsset(ctx, structuredMasterInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf2a101", userB, assetReady, now)); !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("cross-user err = %v, want ErrAssetNotFound", err)
	}
	if _, err := repo.CreateStructuredMasterFromAsset(ctx, structuredMasterInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf2a102", userA, assetProcessing, now)); !errors.Is(err, resumestore.ErrAssetParseNotReady) {
		t.Fatalf("processing err = %v, want ErrAssetParseNotReady", err)
	}

	first, err := repo.CreateStructuredMasterFromAsset(ctx, structuredMasterInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf2a103", userA, assetReady, now))
	if err != nil {
		t.Fatalf("CreateStructuredMasterFromAsset first: %v", err)
	}
	if first.VersionType != sharedtypes.ResumeVersionTypeStructuredMaster || first.ResumeAssetID != assetReady {
		t.Fatalf("first version = %+v", first)
	}
	if _, err := repo.CreateStructuredMasterFromAsset(ctx, structuredMasterInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf2a104", userA, assetReady, now.Add(time.Second))); !errors.Is(err, resumestore.ErrStructuredMasterAlreadyExists) {
		t.Fatalf("duplicate err = %v, want ErrStructuredMasterAlreadyExists", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `select count(*) from resume_versions where resume_asset_id = $1 and version_type = 'structured_master' and deleted_at is null`, assetReady).Scan(&count); err != nil {
		t.Fatalf("count active masters: %v", err)
	}
	if count != 1 {
		t.Fatalf("active structured master count = %d, want 1", count)
	}

	mustExec(t, ctx, db, `update resume_versions set deleted_at = $2 where id = $1`, first.ID, now.Add(2*time.Second))
	second, err := repo.CreateStructuredMasterFromAsset(ctx, structuredMasterInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf2a105", userA, assetReady, now.Add(3*time.Second)))
	if err != nil {
		t.Fatalf("CreateStructuredMasterFromAsset after soft delete: %v", err)
	}
	if second.ID == first.ID {
		t.Fatalf("soft-delete replacement reused first id: %s", second.ID)
	}
}

func TestResumeVersionListPaginationCrossUserAndCursor(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureStructuredMasterUniqueIndex(t, ctx, db)

	repo := resumestore.NewRepository(db)
	base := time.Date(2026, 5, 17, 18, 45, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf3a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf3a002"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf3a003"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-version-list-a@example.com', 'active'), ($2, 'resume-version-list-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status) values ($1, $2, 'Ready Resume', 'en', 'ready')`, assetID, userA)
	for i := 0; i < 25; i++ {
		versionID := resumeVersionIntegrationID(i)
		updatedAt := base.Add(-time.Duration(i) * time.Minute)
		versionType := "targeted"
		seedStrategy := "blank"
		if i == 0 {
			versionType = "structured_master"
			seedStrategy = ""
		}
		if seedStrategy == "" {
			mustExec(t, ctx, db, `insert into resume_versions(id, user_id, resume_asset_id, version_type, display_name, structured_profile, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$7)`, versionID, userA, assetID, versionType, "Version", []byte(`{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`), updatedAt)
			continue
		}
		mustExec(t, ctx, db, `insert into resume_versions(id, user_id, resume_asset_id, version_type, display_name, seed_strategy, structured_profile, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, versionID, userA, assetID, versionType, "Version", seedStrategy, []byte(`{"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`), updatedAt)
	}

	first, err := repo.ListVersionsByAsset(ctx, userA, assetID, resumestore.VersionListFilter{PageSize: 20})
	if err != nil {
		t.Fatalf("ListVersionsByAsset first page: %v", err)
	}
	if len(first.Items) != 20 || !first.HasMore || first.NextCursor == "" || first.Items[0].ID != resumeVersionIntegrationID(0) || first.Items[19].ID != resumeVersionIntegrationID(19) {
		t.Fatalf("first page = %+v", first)
	}
	second, err := repo.ListVersionsByAsset(ctx, userA, assetID, resumestore.VersionListFilter{PageSize: 20, Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("ListVersionsByAsset second page: %v", err)
	}
	if len(second.Items) != 5 || second.HasMore || second.Items[0].ID != resumeVersionIntegrationID(20) {
		t.Fatalf("second page = %+v", second)
	}
	if _, err := repo.ListVersionsByAsset(ctx, userB, assetID, resumestore.VersionListFilter{}); !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("cross-user list err = %v, want ErrAssetNotFound", err)
	}
	if _, err := repo.GetVersionByID(ctx, userB, resumeVersionIntegrationID(0)); !errors.Is(err, resumestore.ErrVersionNotFound) {
		t.Fatalf("cross-user get err = %v, want ErrVersionNotFound", err)
	}
	if _, err := repo.ListVersionsByAsset(ctx, userA, assetID, resumestore.VersionListFilter{Cursor: "not-a-cursor"}); !errors.Is(err, resumestore.ErrInvalidCursor) {
		t.Fatalf("invalid cursor err = %v, want ErrInvalidCursor", err)
	}
}

func TestResumeVersionUpdatePatchMergeCrossUserAndDeleted(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureStructuredMasterUniqueIndex(t, ctx, db)

	repo := resumestore.NewRepository(db)
	base := time.Date(2026, 5, 17, 19, 30, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf4a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf4a002"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf4a003"
	versionID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf4b001"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-version-update-a@example.com', 'active'), ($2, 'resume-version-update-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status) values ($1, $2, 'Ready Resume', 'en', 'ready')`, assetID, userA)
	mustExec(t, ctx, db, `insert into resume_versions(id, user_id, resume_asset_id, version_type, display_name, structured_profile, created_at, updated_at, prompt_version, rubric_version, model_id, provider) values ($1,$2,$3,'structured_master','Structured master',$4,$5,$5,'p','r','m','provider')`,
		versionID,
		userA,
		assetID,
		[]byte(`{"headline":"Senior engineer","summary":"old","skills":["Go"],"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`),
		base,
	)

	focusAngle := "Reliability"
	matchScore := 0.91
	updated, err := repo.UpdateVersionPatch(ctx, resumestore.VersionUpdateInput{
		UserID:               userA,
		VersionID:            versionID,
		DisplayName:          stringPointer("Updated master"),
		DisplayNameSet:       true,
		FocusAngle:           &focusAngle,
		FocusAngleSet:        true,
		MatchScore:           &matchScore,
		MatchScoreSet:        true,
		StructuredProfileSet: true,
		StructuredProfilePatch: map[string]any{
			"summary": "new",
			"provenance": map[string]any{
				"promptVersion": "client-controlled",
			},
		},
		Now: base.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("UpdateVersionPatch: %v", err)
	}
	var profile map[string]any
	if err := json.Unmarshal(updated.StructuredProfile, &profile); err != nil {
		t.Fatalf("decode updated profile: %v", err)
	}
	if updated.DisplayName != "Updated master" || updated.FocusAngle == nil || *updated.FocusAngle != focusAngle || updated.MatchScore == nil || *updated.MatchScore != matchScore {
		t.Fatalf("updated record = %+v", updated)
	}
	if profile["summary"] != "new" || profile["headline"] != "Senior engineer" {
		t.Fatalf("merged profile = %+v", profile)
	}
	provenance, _ := profile["provenance"].(map[string]any)
	if provenance["promptVersion"] != "p" {
		t.Fatalf("provenance was not preserved: %+v", provenance)
	}
	if _, err := repo.UpdateVersionPatch(ctx, resumestore.VersionUpdateInput{UserID: userB, VersionID: versionID, DisplayName: stringPointer("bad"), DisplayNameSet: true, Now: base.Add(2 * time.Minute)}); !errors.Is(err, resumestore.ErrVersionNotFound) {
		t.Fatalf("cross-user update err = %v, want ErrVersionNotFound", err)
	}
	mustExec(t, ctx, db, `update resume_versions set deleted_at = $2 where id = $1`, versionID, base.Add(3*time.Minute))
	if _, err := repo.UpdateVersionPatch(ctx, resumestore.VersionUpdateInput{UserID: userA, VersionID: versionID, DisplayName: stringPointer("bad"), DisplayNameSet: true, Now: base.Add(4 * time.Minute)}); !errors.Is(err, resumestore.ErrVersionNotFound) {
		t.Fatalf("deleted update err = %v, want ErrVersionNotFound", err)
	}
}

func TestBranchVersionInsertStrategiesCrossUserAndRollback(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureStructuredMasterUniqueIndex(t, ctx, db)

	repo := resumestore.NewRepository(db)
	base := time.Date(2026, 5, 17, 20, 30, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5a002"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5a003"
	targetA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5a004"
	targetB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5a005"
	parentID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5b001"
	dummyJobID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf5d777"
	t.Cleanup(func() {
		_, _ = db.Exec(`delete from async_jobs where id = $1`, dummyJobID)
		cleanupResumeStoreUsers(t, db, userA, userB)
	})

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-branch-a@example.com', 'active'), ($2, 'resume-branch-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status) values ($1, $2, 'Ready Resume', 'en', 'ready')`, assetID, userA)
	mustExec(t, ctx, db, `insert into target_jobs(id, user_id, source_type, analysis_status) values ($1, $2, 'manual_text', 'ready'), ($3, $4, 'manual_text', 'ready')`, targetA, userA, targetB, userB)
	mustExec(t, ctx, db, `insert into resume_versions(id, user_id, resume_asset_id, version_type, display_name, structured_profile, created_at, updated_at, prompt_version, rubric_version, model_id, provider) values ($1,$2,$3,'structured_master','Structured master',$4,$5,$5,'resume_profile.v1','not_applicable','model-1','provider')`,
		parentID,
		userA,
		assetID,
		[]byte(`{"headline":"Senior engineer","summary":"master","skills":["Go"],"sections":[{"id":"s1"}],"provenance":{"promptVersion":"resume_profile.v1","rubricVersion":"not_applicable","modelId":"model-1","language":"en","featureFlag":"resume-workshop-additive","dataSourceVersion":"resume_asset.v1"}}`),
		base,
	)

	copyOut, err := repo.BranchFromParent(ctx, branchInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf5b002", userA, parentID, targetA, sharedtypes.ResumeSeedStrategyCopyMaster, base.Add(time.Minute)))
	if err != nil {
		t.Fatalf("BranchFromParent copy_master: %v", err)
	}
	if copyOut.Version.VersionType != sharedtypes.ResumeVersionTypeTargeted || copyOut.Version.ParentVersionID == nil || *copyOut.Version.ParentVersionID != parentID || copyOut.Version.TargetJobID == nil || *copyOut.Version.TargetJobID != targetA {
		t.Fatalf("copy version = %+v", copyOut.Version)
	}
	var copyProfile map[string]any
	if err := json.Unmarshal(copyOut.Version.StructuredProfile, &copyProfile); err != nil {
		t.Fatalf("decode copy profile: %v", err)
	}
	if copyProfile["summary"] != "master" {
		t.Fatalf("copy profile did not preserve parent summary: %+v", copyProfile)
	}
	copyProvenance, _ := copyProfile["provenance"].(map[string]any)
	if copyProvenance["promptVersion"] != "resume_branch.copy_master.v1" {
		t.Fatalf("copy provenance = %+v", copyProvenance)
	}

	blankOut, err := repo.BranchFromParent(ctx, branchInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf5b003", userA, parentID, targetA, sharedtypes.ResumeSeedStrategyBlank, base.Add(2*time.Minute)))
	if err != nil {
		t.Fatalf("BranchFromParent blank: %v", err)
	}
	var blankProfile map[string]any
	if err := json.Unmarshal(blankOut.Version.StructuredProfile, &blankProfile); err != nil {
		t.Fatalf("decode blank profile: %v", err)
	}
	if blankProfile["headline"] != "" || len(blankProfile["skills"].([]any)) != 0 {
		t.Fatalf("blank profile = %+v", blankProfile)
	}

	aiIn := branchInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf5b004", userA, parentID, targetA, sharedtypes.ResumeSeedStrategyAiSelect, base.Add(3*time.Minute))
	aiIn.TailorRunID = "0195f2d0-4a44-7fc2-8f77-1f9c4cf5c001"
	aiIn.JobID = "0195f2d0-4a44-7fc2-8f77-1f9c4cf5d001"
	aiIn.DedupeKey = "dedupe-ai"
	aiOut, err := repo.BranchFromParent(ctx, aiIn)
	if err != nil {
		t.Fatalf("BranchFromParent ai_select: %v", err)
	}
	if aiOut.TailorRunID != aiIn.TailorRunID || aiOut.JobID != aiIn.JobID || aiOut.JobStatus != sharedtypes.JobStatusQueued {
		t.Fatalf("ai result = %+v", aiOut)
	}
	var runStatus, jobType, resourceType, jobStatus string
	if err := db.QueryRowContext(ctx, `select rr.status, j.job_type, j.resource_type, j.status from resume_tailor_runs rr join async_jobs j on j.resource_id = rr.id where rr.id = $1`, aiIn.TailorRunID).Scan(&runStatus, &jobType, &resourceType, &jobStatus); err != nil {
		t.Fatalf("query ai_select run/job: %v", err)
	}
	if runStatus != "queued" || jobType != "resume_tailor" || resourceType != "resume_tailor_run" || jobStatus != "queued" {
		t.Fatalf("run/job status = %s %s %s %s", runStatus, jobType, resourceType, jobStatus)
	}

	if _, err := repo.BranchFromParent(ctx, branchInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf5b005", userB, parentID, targetA, sharedtypes.ResumeSeedStrategyCopyMaster, base.Add(4*time.Minute))); !errors.Is(err, resumestore.ErrVersionNotFound) {
		t.Fatalf("cross-user parent err = %v, want ErrVersionNotFound", err)
	}
	if _, err := repo.BranchFromParent(ctx, branchInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf5b006", userA, parentID, targetB, sharedtypes.ResumeSeedStrategyCopyMaster, base.Add(5*time.Minute))); !errors.Is(err, resumestore.ErrVersionNotFound) {
		t.Fatalf("cross-user target err = %v, want ErrVersionNotFound", err)
	}

	mustExec(t, ctx, db, `insert into async_jobs(id, job_type, resource_type, resource_id, status) values ($1, 'resume_tailor', 'resume_tailor_run', $2, 'queued') on conflict (id) do nothing`, dummyJobID, "0195f2d0-4a44-7fc2-8f77-1f9c4cf5c777")
	rollbackIn := branchInput("0195f2d0-4a44-7fc2-8f77-1f9c4cf5b007", userA, parentID, targetA, sharedtypes.ResumeSeedStrategyAiSelect, base.Add(6*time.Minute))
	rollbackIn.TailorRunID = "0195f2d0-4a44-7fc2-8f77-1f9c4cf5c007"
	rollbackIn.JobID = dummyJobID
	rollbackIn.DedupeKey = "dedupe-rollback"
	if _, err := repo.BranchFromParent(ctx, rollbackIn); err == nil {
		t.Fatal("expected duplicate async job id to fail")
	}
	var versionCount, runCount int
	if err := db.QueryRowContext(ctx, `select count(*) from resume_versions where id = $1`, rollbackIn.VersionID).Scan(&versionCount); err != nil {
		t.Fatalf("count rollback version: %v", err)
	}
	if err := db.QueryRowContext(ctx, `select count(*) from resume_tailor_runs where id = $1`, rollbackIn.TailorRunID).Scan(&runCount); err != nil {
		t.Fatalf("count rollback run: %v", err)
	}
	if versionCount != 0 || runCount != 0 {
		t.Fatalf("rollback left rows: version=%d run=%d", versionCount, runCount)
	}
}

func TestResumeTailorRunStoreStateTransitionsIsolationAndClaim(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo := resumestore.NewRepository(db)
	base := time.Date(2026, 5, 18, 10, 30, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf6a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf6a002"
	assetA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf6a003"
	targetA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf6a004"
	targetB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf6a005"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-tailor-a@example.com', 'active'), ($2, 'resume-tailor-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status) values ($1, $2, 'Ready Resume', 'en', 'ready')`, assetA, userA)
	mustExec(t, ctx, db, `insert into target_jobs(id, user_id, source_type, analysis_status) values ($1, $2, 'manual_text', 'ready'), ($3, $4, 'manual_text', 'ready')`, targetA, userA, targetB, userB)

	created, err := repo.CreateTailorRun(ctx, resumestore.CreateTailorRunInput{
		TailorRunID:   "0195f2d0-4a44-7fc2-8f77-1f9c4cf6b001",
		JobID:         "0195f2d0-4a44-7fc2-8f77-1f9c4cf6c001",
		UserID:        userA,
		TargetJobID:   targetA,
		ResumeAssetID: assetA,
		Mode:          "gap_review",
		DedupeKey:     "dedupe-tailor-1",
		Now:           base,
	})
	if err != nil {
		t.Fatalf("CreateTailorRun: %v", err)
	}
	if created.TailorRunID == "" || created.JobID == "" || created.JobStatus != sharedtypes.JobStatusQueued {
		t.Fatalf("create result = %+v", created)
	}
	got, err := repo.GetTailorRun(ctx, userA, created.TailorRunID)
	if err != nil {
		t.Fatalf("GetTailorRun: %v", err)
	}
	if got.Status != "queued" || got.Mode != "gap_review" || got.TargetJobID != targetA || got.ResumeAssetID != assetA {
		t.Fatalf("queued run = %+v", got)
	}
	if _, err := repo.GetTailorRun(ctx, userB, created.TailorRunID); !errors.Is(err, resumestore.ErrTailorRunNotFound) {
		t.Fatalf("cross-user get err = %v, want ErrTailorRunNotFound", err)
	}
	if _, err := repo.CreateTailorRun(ctx, resumestore.CreateTailorRunInput{TailorRunID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6b002", JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6c002", UserID: userB, TargetJobID: targetA, ResumeAssetID: assetA, Mode: "gap_review", Now: base}); !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("cross-user asset/target err = %v, want ErrAssetNotFound", err)
	}
	if _, err := repo.CreateTailorRun(ctx, resumestore.CreateTailorRunInput{TailorRunID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6b003", JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6c003", UserID: userA, TargetJobID: targetB, ResumeAssetID: assetA, Mode: "gap_review", Now: base}); !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("foreign target err = %v, want ErrAssetNotFound", err)
	}

	generating, err := repo.MarkTailorRunGenerating(ctx, resumestore.TailorRunStatusInput{TailorRunID: created.TailorRunID, Now: base.Add(time.Minute)})
	if err != nil {
		t.Fatalf("MarkTailorRunGenerating: %v", err)
	}
	if generating.Status != "generating" {
		t.Fatalf("generating status = %+v", generating)
	}
	if _, err := repo.MarkTailorRunGenerating(ctx, resumestore.TailorRunStatusInput{TailorRunID: created.TailorRunID, Now: base.Add(2 * time.Minute)}); !errors.Is(err, resumestore.ErrInvalidStateTransition) {
		t.Fatalf("second generating err = %v, want ErrInvalidStateTransition", err)
	}

	ready, err := repo.MarkTailorRunReady(ctx, resumestore.TailorRunReadyInput{
		TailorRunID:  created.TailorRunID,
		MatchSummary: json.RawMessage(`{"strengths":["Strong systems evidence"],"gaps":["Add edge runtime detail"]}`),
		Suggestions:  json.RawMessage(`[{"originalBullet":"Led migration.","suggestedBullet":"Led migration across 12 teams.","reason":"Adds scope."}]`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "resume_tailor.v2", RubricVersion: "not_applicable", ModelID: "model-profile:contract.default", Provider: "stub", Language: "zh-CN", FeatureFlag: "none", DataSourceVersion: "target_job.v17",
		},
		Now: base.Add(3 * time.Minute),
	})
	if err != nil {
		t.Fatalf("MarkTailorRunReady: %v", err)
	}
	if ready.Status != "ready" || string(ready.MatchSummary) == "{}" || string(ready.Suggestions) == "[]" || ready.Provenance.PromptVersion != "resume_tailor.v2" {
		t.Fatalf("ready run = %+v", ready)
	}

	failedIn := resumestore.CreateTailorRunInput{TailorRunID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6b004", JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6c004", UserID: userA, TargetJobID: targetA, ResumeAssetID: assetA, Mode: "bullet_suggestions", Now: base.Add(4 * time.Minute)}
	failedCreated, err := repo.CreateTailorRun(ctx, failedIn)
	if err != nil {
		t.Fatalf("CreateTailorRun failed path: %v", err)
	}
	if _, err := repo.MarkTailorRunGenerating(ctx, resumestore.TailorRunStatusInput{TailorRunID: failedCreated.TailorRunID, Now: base.Add(5 * time.Minute)}); err != nil {
		t.Fatalf("MarkTailorRunGenerating failed path: %v", err)
	}
	failed, err := repo.MarkTailorRunFailed(ctx, resumestore.TailorRunFailureInput{TailorRunID: failedCreated.TailorRunID, ErrorCode: "AI_OUTPUT_INVALID", Now: base.Add(6 * time.Minute)})
	if err != nil {
		t.Fatalf("MarkTailorRunFailed: %v", err)
	}
	if failed.Status != "failed" || failed.ErrorCode == nil || *failed.ErrorCode != "AI_OUTPUT_INVALID" {
		t.Fatalf("failed run = %+v", failed)
	}

	concurrent, err := repo.CreateTailorRun(ctx, resumestore.CreateTailorRunInput{TailorRunID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6b005", JobID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf6c005", UserID: userA, TargetJobID: targetA, ResumeAssetID: assetA, Mode: "gap_review", Now: base.Add(7 * time.Minute)})
	if err != nil {
		t.Fatalf("CreateTailorRun concurrent path: %v", err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			_, err := repo.MarkTailorRunGenerating(ctx, resumestore.TailorRunStatusInput{TailorRunID: concurrent.TailorRunID, Now: base.Add(time.Duration(8+offset) * time.Minute)})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if !errors.Is(err, resumestore.ErrInvalidStateTransition) {
			t.Fatalf("concurrent claim err = %v, want ErrInvalidStateTransition", err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent successful claims = %d, want 1", successes)
	}
}

func TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo := resumestore.NewRepository(db)
	base := time.Date(2026, 5, 18, 13, 30, 0, 0, time.UTC)
	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7d001"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7d002"
	targetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7d003"
	versionID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7d004"
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7d005"
	failedRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7d006"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userID) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-tailor-complete@example.com', 'active')`, userID)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status, parsed_summary) values ($1, $2, 'Ready Resume', 'en', 'ready', '{"headline":"Senior engineer"}')`, assetID, userID)
	mustExec(t, ctx, db, `insert into target_jobs(id, user_id, source_type, analysis_status, title, seniority_level, summary) values ($1, $2, 'manual_text', 'ready', 'Staff Backend Engineer', 'staff', '{"requirements":["distributed systems"]}')`, targetID, userID)
	mustExec(t, ctx, db, `insert into resume_versions(id, user_id, resume_asset_id, version_type, target_job_id, display_name, structured_profile, created_at, updated_at) values ($1,$2,$3,'targeted',$4,'Targeted','{"sections":[{"bullets":["Led migration."]}]}',$5,$5)`, versionID, userID, assetID, targetID, base)
	mustExec(t, ctx, db, `insert into resume_tailor_runs(id, user_id, target_job_id, resume_asset_id, mode, status, created_at, updated_at) values ($1,$2,$3,$4,'gap_review','generating',$5,$5), ($6,$2,$3,$4,'gap_review','generating',$5,$5)`, tailorRunID, userID, targetID, assetID, base, failedRunID)

	loaded, err := repo.GetForTailor(ctx, tailorRunID, versionID)
	if err != nil {
		t.Fatalf("GetForTailor: %v", err)
	}
	if loaded.ResumeVersionID != versionID || loaded.UserID != userID || loaded.TargetTitle != "Staff Backend Engineer" || loaded.TargetSeniority != "staff" {
		t.Fatalf("loaded context = %+v", loaded)
	}

	privateText := "PRIVATE_SUGGESTED_BULLET"
	if err := repo.CompleteTailorRunSuccess(ctx, resumestore.CompleteTailorRunSuccessInput{
		TailorRunID:     tailorRunID,
		ResumeVersionID: versionID,
		MatchSummary:    json.RawMessage(`{"strengths":["Strong systems evidence"],"gaps":["Add scale metrics"]}`),
		Suggestions: []resumestore.TailorSuggestionInput{
			{ID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf7e001", OriginalBullet: "Led migration.", SuggestedBullet: privateText, Reason: "Adds scope."},
			{ID: "0195f2d0-4a44-7fc2-8f77-1f9c4cf7e002", OriginalBullet: "Built services.", SuggestedBullet: "Built reliable services.", Reason: "Adds outcome."},
		},
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", ModelID: "fixture-model:resume-tailor", Provider: "stub", Language: "en", FeatureFlag: "none", DataSourceVersion: "target_job.v1",
		},
		OutboxEventID:      "0195f2d0-4a44-7fc2-8f77-1f9c4cf7e003",
		OutboxEventPayload: []byte(`{"tailorRunId":"` + tailorRunID + `","resumeAssetId":"` + assetID + `","targetJobId":"` + targetID + `","mode":"gap_review","status":"ready"}`),
		Now:                base.Add(time.Minute),
	}); err != nil {
		t.Fatalf("CompleteTailorRunSuccess: %v", err)
	}
	var runStatus string
	var suggestionCount int
	if err := db.QueryRowContext(ctx, `select status from resume_tailor_runs where id = $1`, tailorRunID).Scan(&runStatus); err != nil {
		t.Fatalf("query run status: %v", err)
	}
	if runStatus != "ready" {
		t.Fatalf("run status = %s, want ready", runStatus)
	}
	roundtrip, err := repo.GetTailorRun(ctx, userID, tailorRunID)
	if err != nil {
		t.Fatalf("GetTailorRun after success: %v", err)
	}
	if roundtrip.Provenance.PromptVersion != "v0.1.0" ||
		roundtrip.Provenance.RubricVersion != "v0.1.0" ||
		roundtrip.Provenance.ModelID != "fixture-model:resume-tailor" ||
		roundtrip.Provenance.Language != "en" ||
		roundtrip.Provenance.FeatureFlag != "none" ||
		roundtrip.Provenance.DataSourceVersion != "target_job.v1" {
		t.Fatalf("tailor run provenance after DB roundtrip = %+v", roundtrip.Provenance)
	}
	if err := db.QueryRowContext(ctx, `select count(*) from resume_version_suggestions where tailor_run_id = $1 and resume_version_id = $2 and status = 'pending'`, tailorRunID, versionID).Scan(&suggestionCount); err != nil {
		t.Fatalf("count suggestions: %v", err)
	}
	if suggestionCount != 2 {
		t.Fatalf("suggestion count = %d, want 2", suggestionCount)
	}
	var payload []byte
	if err := db.QueryRowContext(ctx, `select payload from outbox_events where aggregate_id = $1 and event_name = 'resume.tailor.completed'`, tailorRunID).Scan(&payload); err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	var outbox map[string]any
	if err := json.Unmarshal(payload, &outbox); err != nil {
		t.Fatalf("decode outbox: %v", err)
	}
	if len(outbox) != 5 || outbox["tailorRunId"] != tailorRunID || outbox["status"] != "ready" {
		t.Fatalf("outbox payload = %+v", outbox)
	}
	if strings.Contains(string(payload), privateText) || strings.Contains(string(payload), "Strong systems evidence") {
		t.Fatalf("outbox payload leaked private content: %s", payload)
	}
	taskRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf7e004"
	if err := storeai.NewTaskRunWriter(db).WriteAITaskRun(ctx, aiclient.AITaskRunRow{
		ID:                  taskRunID,
		UserID:              userID,
		Capability:          aiclient.AITaskRunTaskResumeTailor,
		ResourceType:        aiclient.AITaskRunResourceResumeTailorRun,
		ResourceID:          tailorRunID,
		Provider:            "stub",
		ModelFamily:         "stub",
		ModelID:             "fixture-model:resume-tailor",
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "resume.tailor.default",
		ModelProfileVersion: "1.1.0",
		FeatureKey:          "resume.tailor.gap_review",
		FeatureFlag:         "none",
		DataSourceVersion:   "target_job.v1",
		Language:            "en",
		InputTokens:         12,
		OutputTokens:        7,
		LatencyMs:           20,
		Status:              aiclient.AITaskRunStatusSuccess,
		ValidationStatus:    aiclient.ValidationStatusOK,
		OutputSchemaVersion: "resume.tailor.v1",
		FallbackChain:       []string{"stub/fixture-model:resume-tailor"},
		StartedAt:           base,
		CompletedAt:         base.Add(20 * time.Millisecond),
	}); err != nil {
		t.Fatalf("WriteAITaskRun: %v", err)
	}
	var taskType, resourceType, featureKey, outputSchemaVersion, validationStatus string
	if err := db.QueryRowContext(ctx, `select task_type, resource_type, feature_key, output_schema_version, validation_status from ai_task_runs where id = $1`, taskRunID).Scan(&taskType, &resourceType, &featureKey, &outputSchemaVersion, &validationStatus); err != nil {
		t.Fatalf("query ai_task_runs: %v", err)
	}
	if taskType != "resume_tailor" || resourceType != "resume_tailor_run" || featureKey != "resume.tailor.gap_review" || outputSchemaVersion != "resume.tailor.v1" || validationStatus != "ok" {
		t.Fatalf("ai_task_runs typed columns drift: task=%s resource=%s feature=%s schema=%s validation=%s", taskType, resourceType, featureKey, outputSchemaVersion, validationStatus)
	}

	if _, err := repo.MarkTailorRunFailed(ctx, resumestore.TailorRunFailureInput{TailorRunID: failedRunID, ErrorCode: "AI_OUTPUT_INVALID", Now: base.Add(2 * time.Minute)}); err != nil {
		t.Fatalf("MarkTailorRunFailed: %v", err)
	}
	var failedOutboxes int
	if err := db.QueryRowContext(ctx, `select count(*) from outbox_events where aggregate_id = $1 and event_name = 'resume.tailor.completed'`, failedRunID).Scan(&failedOutboxes); err != nil {
		t.Fatalf("count failed outbox: %v", err)
	}
	if failedOutboxes != 0 {
		t.Fatalf("failed run completed outboxes = %d, want 0", failedOutboxes)
	}
}

func TestResumeSuggestionDecisionCASIsolationAndProfileStability(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo := resumestore.NewRepository(db)
	base := time.Date(2026, 5, 18, 14, 30, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a002"
	assetA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a003"
	assetB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a004"
	targetA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a005"
	targetB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8a006"
	versionA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8b001"
	versionB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8b002"
	runA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8c001"
	runB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8c002"
	acceptID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d001"
	rejectID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d002"
	concurrentID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d003"
	foreignID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf8d004"
	profile := []byte(`{"headline":"Senior engineer","sections":[{"bullets":["Improved reliability."]}],"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","language":"en","featureFlag":"f","dataSourceVersion":"d"}}`)
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-suggestion-a@example.com', 'active'), ($2, 'resume-suggestion-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into resume_assets(id, user_id, title, language, parse_status) values ($1, $2, 'Ready A', 'en', 'ready'), ($3, $4, 'Ready B', 'en', 'ready')`, assetA, userA, assetB, userB)
	mustExec(t, ctx, db, `insert into target_jobs(id, user_id, source_type, analysis_status) values ($1, $2, 'manual_text', 'ready'), ($3, $4, 'manual_text', 'ready')`, targetA, userA, targetB, userB)
	mustExec(t, ctx, db, `insert into resume_versions(id, user_id, resume_asset_id, version_type, target_job_id, display_name, structured_profile, created_at, updated_at, prompt_version, rubric_version, model_id, provider) values ($1,$2,$3,'targeted',$4,'Targeted A',$5,$6,$6,'p','r','m','fixture'), ($7,$8,$9,'targeted',$10,'Targeted B',$5,$6,$6,'p','r','m','fixture')`,
		versionA, userA, assetA, targetA, profile, base,
		versionB, userB, assetB, targetB,
	)
	mustExec(t, ctx, db, `insert into resume_tailor_runs(id, user_id, target_job_id, resume_asset_id, mode, status, prompt_version, rubric_version, model_id, provider, language, feature_flag, data_source_version, created_at, updated_at) values ($1,$2,$3,$4,'gap_review','ready','resume_tailor_suggestion.v1','resume_tailor.rubric.v1','fixture-model:resume-tailor-suggestion','fixture-provider','zh-CN','tailor-flag','target_job.v17',$5,$5), ($6,$7,$8,$9,'gap_review','ready','resume_tailor_suggestion.v1','resume_tailor.rubric.v1','fixture-model:resume-tailor-suggestion','fixture-provider','zh-CN','tailor-flag','target_job.v17',$5,$5)`,
		runA, userA, targetA, assetA, base,
		runB, userB, targetB, assetB,
	)
	for _, row := range []struct {
		id      string
		version string
		run     string
	}{
		{id: acceptID, version: versionA, run: runA},
		{id: rejectID, version: versionA, run: runA},
		{id: concurrentID, version: versionA, run: runA},
		{id: foreignID, version: versionB, run: runB},
	} {
		mustExec(t, ctx, db, `insert into resume_version_suggestions(id, resume_version_id, tailor_run_id, original_bullet, suggested_bullet, reason, status, created_at) values ($1,$2,$3,'Improved reliability.','Improved reliability with release guardrails.','Adds evidence.','pending',$4)`, row.id, row.version, row.run, base)
	}

	accepted, err := repo.DecideResumeSuggestion(ctx, resumestore.DecideSuggestionInput{
		UserID:          userA,
		ResumeVersionID: versionA,
		SuggestionID:    acceptID,
		Decision:        sharedtypes.ResumeTailorSuggestionStatusAccepted,
		Now:             base.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("DecideResumeSuggestion accept: %v", err)
	}
	if !jsonBytesEqual(accepted.StructuredProfile, profile) || !accepted.UpdatedAt.Equal(base.Add(time.Minute)) {
		t.Fatalf("accepted version mutated profile or timestamp incorrectly: %+v profile=%s", accepted, accepted.StructuredProfile)
	}
	assertSuggestionState(t, accepted.Suggestions, acceptID, "accepted", true)
	if _, err := repo.DecideResumeSuggestion(ctx, resumestore.DecideSuggestionInput{UserID: userA, ResumeVersionID: versionA, SuggestionID: acceptID, Decision: sharedtypes.ResumeTailorSuggestionStatusRejected, Now: base.Add(2 * time.Minute)}); !errors.Is(err, resumestore.ErrSuggestionAlreadyDecided) {
		t.Fatalf("already decided err = %v, want ErrSuggestionAlreadyDecided", err)
	}
	if _, err := repo.DecideResumeSuggestion(ctx, resumestore.DecideSuggestionInput{UserID: userB, ResumeVersionID: versionA, SuggestionID: acceptID, Decision: sharedtypes.ResumeTailorSuggestionStatusRejected, Now: base.Add(2 * time.Minute)}); !errors.Is(err, resumestore.ErrSuggestionNotFound) {
		t.Fatalf("cross-user err = %v, want ErrSuggestionNotFound", err)
	}

	rejected, err := repo.DecideResumeSuggestion(ctx, resumestore.DecideSuggestionInput{
		UserID:          userA,
		ResumeVersionID: versionA,
		SuggestionID:    rejectID,
		Decision:        sharedtypes.ResumeTailorSuggestionStatusRejected,
		Now:             base.Add(3 * time.Minute),
	})
	if err != nil {
		t.Fatalf("DecideResumeSuggestion reject: %v", err)
	}
	assertSuggestionState(t, rejected.Suggestions, rejectID, "rejected", true)

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, decision := range []sharedtypes.ResumeTailorSuggestionStatus{sharedtypes.ResumeTailorSuggestionStatusAccepted, sharedtypes.ResumeTailorSuggestionStatusRejected} {
		wg.Add(1)
		go func(decision sharedtypes.ResumeTailorSuggestionStatus) {
			defer wg.Done()
			_, err := repo.DecideResumeSuggestion(ctx, resumestore.DecideSuggestionInput{
				UserID:          userA,
				ResumeVersionID: versionA,
				SuggestionID:    concurrentID,
				Decision:        decision,
				Now:             base.Add(4 * time.Minute),
			})
			errs <- err
		}(decision)
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if !errors.Is(err, resumestore.ErrSuggestionAlreadyDecided) {
			t.Fatalf("concurrent decision err = %v, want ErrSuggestionAlreadyDecided", err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent successful decisions = %d, want 1", successes)
	}
	var profileAfter []byte
	if err := db.QueryRowContext(ctx, `select structured_profile from resume_versions where id = $1`, versionA).Scan(&profileAfter); err != nil {
		t.Fatalf("query profile after decisions: %v", err)
	}
	if !jsonBytesEqual(profileAfter, profile) {
		t.Fatalf("structured_profile changed after decisions: %s", profileAfter)
	}
}

func structuredMasterInput(versionID, userID, assetID string, now time.Time) resumestore.CreateStructuredMasterInput {
	return resumestore.CreateStructuredMasterInput{
		VersionID:         versionID,
		UserID:            userID,
		ResumeAssetID:     assetID,
		DisplayName:       "Structured master",
		StructuredProfile: []byte(`{"headline":"Senior engineer","provenance":{"promptVersion":"resume_profile.v1","rubricVersion":"not_applicable","modelId":"fixture-model:resume-version-profile","language":"en","featureFlag":"resume-workshop-additive","dataSourceVersion":"resume_asset.v1"}}`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion:     "resume_profile.v1",
			RubricVersion:     "not_applicable",
			ModelID:           "fixture-model:resume-version-profile",
			Provider:          "fixture-provider",
			Language:          "en",
			FeatureFlag:       "resume-workshop-additive",
			DataSourceVersion: "resume_asset.v1",
		},
		Now: now,
	}
}

func branchInput(versionID, userID, parentID, targetID string, strategy sharedtypes.ResumeSeedStrategy, now time.Time) resumestore.BranchVersionInput {
	focusAngle := "Platform evidence"
	return resumestore.BranchVersionInput{
		VersionID:       versionID,
		UserID:          userID,
		ParentVersionID: parentID,
		TargetJobID:     targetID,
		SeedStrategy:    strategy,
		DisplayName:     "Targeted",
		FocusAngle:      &focusAngle,
		Provenance: resumestore.VersionProvenance{
			PromptVersion:     "resume_branch." + string(strategy) + ".v1",
			RubricVersion:     "not_applicable",
			ModelID:           "not_applicable",
			Provider:          "system",
			Language:          "en",
			FeatureFlag:       "resume-workshop-additive",
			DataSourceVersion: "resume_version.v1",
		},
		Now: now,
	}
}

func resumeVersionIntegrationID(i int) string {
	return fmt.Sprintf("0195f2d0-4a44-7fc2-8f77-1f9c4cf3b%03d", i)
}

func stringPointer(in string) *string {
	return &in
}

func ensureStructuredMasterUniqueIndex(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var exists bool
	if err := db.QueryRowContext(ctx, `select exists(select 1 from pg_indexes where schemaname = 'public' and indexname = 'uq_resume_versions_structured_master_per_asset')`).Scan(&exists); err != nil {
		t.Fatalf("check structured master unique index: %v", err)
	}
	if !exists {
		t.Skip("structured master unique index is not migrated; run make migrate-up before live resume version test")
	}
}

func assertSuggestionState(t *testing.T, suggestions []any, suggestionID string, wantStatus string, wantDecided bool) {
	t.Helper()
	for _, item := range suggestions {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if m["id"] != suggestionID {
			continue
		}
		if m["status"] != wantStatus {
			t.Fatalf("suggestion %s status = %v, want %s", suggestionID, m["status"], wantStatus)
		}
		_, hasDecidedAt := m["decidedAt"].(string)
		if hasDecidedAt != wantDecided {
			t.Fatalf("suggestion %s decidedAt presence = %v, want %v (%+v)", suggestionID, hasDecidedAt, wantDecided, m)
		}
		provenance, ok := m["provenance"].(map[string]any)
		if !ok {
			t.Fatalf("suggestion %s missing provenance: %+v", suggestionID, m)
		}
		if provenance["promptVersion"] != "resume_tailor_suggestion.v1" ||
			provenance["rubricVersion"] != "resume_tailor.rubric.v1" ||
			provenance["modelId"] != "fixture-model:resume-tailor-suggestion" ||
			provenance["language"] != "zh-CN" ||
			provenance["featureFlag"] != "tailor-flag" ||
			provenance["dataSourceVersion"] != "target_job.v17" {
			t.Fatalf("suggestion %s provenance = %+v", suggestionID, provenance)
		}
		return
	}
	t.Fatalf("suggestion %s not found in %+v", suggestionID, suggestions)
}

func jsonBytesEqual(left, right []byte) bool {
	var leftValue any
	var rightValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		return false
	}
	if err := json.Unmarshal(right, &rightValue); err != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}
