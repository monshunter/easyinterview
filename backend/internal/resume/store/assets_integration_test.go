//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestResumeAssetsIntegrationCRUDStateIsolationPaginationAndRollback(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureResumeAssetsErrorCodeColumn(t, ctx, db)

	repo := resumestore.NewRepository(db)
	now := time.Date(2026, 5, 13, 6, 0, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a002"
	fileID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a003"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a004"
	jobID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a005"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-store-a@example.com', 'active'), ($2, 'resume-store-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into file_objects(id, user_id, purpose, object_key, original_file_name, content_type, byte_size, upload_status) values ($1, $2, 'resume', 'user-a/resume/file.pdf', 'file.pdf', 'application/pdf', 128, 'uploaded')`, fileID, userA)

	created, err := repo.CreateWithParseJob(ctx, resumestore.CreateAssetInput{
		AssetID:      assetID,
		UserID:       userA,
		JobID:        jobID,
		DedupeKey:    "resume-integration-dedupe",
		SourceType:   "upload",
		FileObjectID: &fileID,
		Title:        "Integration Resume",
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		JobStatus:    sharedtypes.JobStatusQueued,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("CreateWithParseJob: %v", err)
	}
	if created.AssetID != assetID || created.JobID != jobID {
		t.Fatalf("created = %+v", created)
	}

	got, err := repo.Get(ctx, userA, assetID)
	if err != nil {
		t.Fatalf("Get owner: %v", err)
	}
	if got.FileObjectID == nil || *got.FileObjectID != fileID || got.SourceType == nil || *got.SourceType != "upload" {
		t.Fatalf("asset file/source fields = %+v", got)
	}
	if _, err := repo.Get(ctx, userB, assetID); !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("cross-user Get err = %v, want ErrAssetNotFound", err)
	}

	if err := repo.MarkParsing(ctx, resumestore.StatusUpdateInput{UserID: userA, AssetID: assetID, Now: now.Add(time.Minute)}); err != nil {
		t.Fatalf("MarkParsing: %v", err)
	}
	if err := repo.MarkReady(ctx, resumestore.MarkReadyInput{UserID: userA, AssetID: assetID, ParsedSummary: []byte(`{"basics":{"name":"Alice"}}`), ParsedTextSnapshot: "parsed", Now: now.Add(2 * time.Minute)}); err != nil {
		t.Fatalf("MarkReady: %v", err)
	}
	ready, err := repo.Get(ctx, userA, assetID)
	if err != nil {
		t.Fatalf("Get ready: %v", err)
	}
	if ready.ParseStatus != sharedtypes.TargetJobParseStatusReady || ready.ParsedTextSnapshot == nil || *ready.ParsedTextSnapshot != "parsed" {
		t.Fatalf("ready asset = %+v", ready)
	}

	for i := 0; i < 24; i++ {
		id := resumeIntegrationID(i)
		job := resumeIntegrationJobID(i)
		_, err := repo.CreateWithParseJob(ctx, resumestore.CreateAssetInput{
			AssetID:     id,
			UserID:      userA,
			JobID:       job,
			DedupeKey:   "resume-integration-dedupe-" + id,
			SourceType:  "paste",
			Title:       "Paged Resume",
			Language:    "en",
			RawText:     "resume text",
			ParseStatus: sharedtypes.TargetJobParseStatusQueued,
			JobStatus:   sharedtypes.JobStatusQueued,
			Now:         now.Add(time.Duration(i+3) * time.Minute),
		})
		if err != nil {
			t.Fatalf("CreateWithParseJob page row %d: %v", i, err)
		}
	}
	first, err := repo.List(ctx, userA, resumestore.ListFilter{PageSize: 20})
	if err != nil {
		t.Fatalf("List first page: %v", err)
	}
	if len(first.Items) != 20 || !first.HasMore || first.NextCursor == "" {
		t.Fatalf("first page len=%d hasMore=%v cursor=%q", len(first.Items), first.HasMore, first.NextCursor)
	}
	second, err := repo.List(ctx, userA, resumestore.ListFilter{PageSize: 20, Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("List second page: %v", err)
	}
	if len(second.Items) != 5 || second.HasMore {
		t.Fatalf("second page len=%d hasMore=%v", len(second.Items), second.HasMore)
	}

	badAssetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1aff1"
	_, err = repo.CreateWithParseJob(ctx, resumestore.CreateAssetInput{
		AssetID:     badAssetID,
		UserID:      userA,
		JobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cf1aff2",
		DedupeKey:   "resume-integration-bad-job",
		SourceType:  "paste",
		Title:       "Rollback Resume",
		Language:    "en",
		RawText:     "resume text",
		ParseStatus: sharedtypes.TargetJobParseStatusQueued,
		JobStatus:   "not_allowed",
		Now:         now,
	})
	if err == nil {
		t.Fatal("expected rollback error for invalid async job status")
	}
	var count int
	if err := db.QueryRowContext(ctx, `select count(*) from resume_assets where id = $1`, badAssetID).Scan(&count); err != nil {
		t.Fatalf("count rollback asset: %v", err)
	}
	if count != 0 {
		t.Fatalf("rollback asset count = %d, want 0", count)
	}
}

func openResumeStoreTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping resume_assets integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	return db
}

func ensureResumeAssetsErrorCodeColumn(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var exists bool
	if err := db.QueryRowContext(ctx, `select exists(select 1 from information_schema.columns where table_name = 'resume_assets' and column_name = 'error_code')`).Scan(&exists); err != nil {
		t.Fatalf("check resume_assets.error_code: %v", err)
	}
	if !exists {
		t.Skip("resume_assets.error_code is not migrated; run make migrate-up before live resume store test")
	}
}

func cleanupResumeStoreUsers(t *testing.T, db *sql.DB, userIDs ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, userID := range userIDs {
		_, _ = db.ExecContext(ctx, `delete from async_jobs where resource_id in (select id from resume_assets where user_id = $1)`, userID)
		_, _ = db.ExecContext(ctx, `delete from resume_assets where user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `delete from file_objects where user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `delete from users where id = $1`, userID)
	}
}

func mustExec(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func resumeIntegrationID(i int) string {
	return fmt.Sprintf("0195f2d0-4a44-7fc2-8f77-1f9c4cf1b%03d", i)
}

func resumeIntegrationJobID(i int) string {
	return fmt.Sprintf("0195f2d0-4a44-7fc2-8f77-1f9c4cf1c%03d", i)
}
