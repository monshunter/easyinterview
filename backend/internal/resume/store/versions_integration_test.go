//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"errors"
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
