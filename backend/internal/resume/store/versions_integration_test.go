//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
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
