package migrations

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
)

func TestResumeVersionsCheckConstraints(t *testing.T) {
	db := openMigrationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ensureResumeVersionTables(t, ctx, db)

	userID := "0195f2d0-4a44-7fc2-8f77-1f9c4ce1b001"
	resumeAssetID := "0195f2d0-4a44-7fc2-8f77-1f9c4ce1b002"
	targetJobID := "0195f2d0-4a44-7fc2-8f77-1f9c4ce1b003"
	resumeVersionID := "0195f2d0-4a44-7fc2-8f77-1f9c4ce1b004"
	tailorRunID := "0195f2d0-4a44-7fc2-8f77-1f9c4ce1b005"

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = db.ExecContext(cleanupCtx, `DELETE FROM users WHERE id = $1`, userID)
	})

	mustExec(t, ctx, db, `INSERT INTO users(id, email, status) VALUES ($1, 'resume-constraint@example.com', 'active')`, userID)
	mustExec(t, ctx, db, `INSERT INTO resume_assets(id, user_id, title, language, parse_status) VALUES ($1, $2, 'Constraint Resume', 'en', 'ready')`, resumeAssetID, userID)
	mustExec(t, ctx, db, `INSERT INTO target_jobs(id, user_id, source_type) VALUES ($1, $2, 'manual_text')`, targetJobID, userID)
	mustExec(t, ctx, db, `INSERT INTO resume_versions(id, user_id, resume_asset_id, version_type, display_name) VALUES ($1, $2, $3, 'structured_master', 'Master')`, resumeVersionID, userID, resumeAssetID)
	mustExec(t, ctx, db, `INSERT INTO resume_tailor_runs(id, user_id, target_job_id, resume_asset_id, mode, status) VALUES ($1, $2, $3, $4, 'gap_review', 'ready')`, tailorRunID, userID, targetJobID, resumeAssetID)

	expectExecCheckViolation(t, ctx, db, `INSERT INTO resume_versions(id, user_id, resume_asset_id, version_type, display_name) VALUES ('0195f2d0-4a44-7fc2-8f77-1f9c4ce1b101', $1, $2, 'foo', 'Invalid')`, userID, resumeAssetID)
	expectExecCheckViolation(t, ctx, db, `INSERT INTO resume_versions(id, user_id, resume_asset_id, version_type, seed_strategy, display_name) VALUES ('0195f2d0-4a44-7fc2-8f77-1f9c4ce1b102', $1, $2, 'targeted', 'bar', 'Invalid')`, userID, resumeAssetID)
	expectExecCheckViolation(t, ctx, db, `UPDATE resume_assets SET source_type = 'unknown' WHERE id = $1`, resumeAssetID)
	expectExecCheckViolation(t, ctx, db, `INSERT INTO resume_version_suggestions(id, resume_version_id, tailor_run_id, original_bullet, suggested_bullet, status) VALUES ('0195f2d0-4a44-7fc2-8f77-1f9c4ce1b103', $1, $2, 'old', 'new', 'unknown')`, resumeVersionID, tailorRunID)
}

func TestResumeVersionsCascadeDelete(t *testing.T) {
	db := openMigrationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ensureResumeVersionTables(t, ctx, db)

	ids := seedResumeVersionGraph(t, ctx, db, "0195f2d0-4a44-7fc2-8f77-1f9c4ce1c")
	t.Cleanup(func() { cleanupUser(db, ids.userID) })

	mustExec(t, ctx, db, `INSERT INTO resume_version_suggestions(id, resume_version_id, tailor_run_id, original_bullet, suggested_bullet) VALUES ($1, $2, $3, 'old', 'new')`, ids.suggestionID, ids.resumeVersionID, ids.tailorRunID)
	mustExec(t, ctx, db, `DELETE FROM resume_versions WHERE id = $1`, ids.resumeVersionID)

	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM resume_version_suggestions WHERE id = $1`, ids.suggestionID).Scan(&count); err != nil {
		t.Fatalf("count suggestion after version delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("resume_version_suggestions rows after deleting parent version = %d, want 0", count)
	}
}

func TestResumeAssetDeleteRequiresVersionCleanup(t *testing.T) {
	db := openMigrationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ensureResumeVersionTables(t, ctx, db)

	ids := seedResumeVersionGraph(t, ctx, db, "0195f2d0-4a44-7fc2-8f77-1f9c4ce1d")
	t.Cleanup(func() { cleanupUser(db, ids.userID) })

	_, err := db.ExecContext(ctx, `DELETE FROM resume_assets WHERE id = $1`, ids.resumeAssetID)
	expectForeignKeyViolation(t, err)
}

func openMigrationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; live migration constraint test skipped")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	return db
}

func ensureResumeVersionTables(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var tableName sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.resume_versions')::text`).Scan(&tableName); err != nil {
		t.Fatalf("check resume_versions table: %v", err)
	}
	if !tableName.Valid {
		t.Skip("resume_versions table is not migrated; run make migrate-up before live constraint test")
	}
}

func mustExec(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

type resumeVersionGraphIDs struct {
	userID          string
	resumeAssetID   string
	targetJobID     string
	resumeVersionID string
	tailorRunID     string
	suggestionID    string
}

func seedResumeVersionGraph(t *testing.T, ctx context.Context, db *sql.DB, prefix string) resumeVersionGraphIDs {
	t.Helper()
	ids := resumeVersionGraphIDs{
		userID:          prefix + "001",
		resumeAssetID:   prefix + "002",
		targetJobID:     prefix + "003",
		resumeVersionID: prefix + "004",
		tailorRunID:     prefix + "005",
		suggestionID:    prefix + "006",
	}
	mustExec(t, ctx, db, `INSERT INTO users(id, email, status) VALUES ($1, $2, 'active')`, ids.userID, ids.userID+"@example.com")
	mustExec(t, ctx, db, `INSERT INTO resume_assets(id, user_id, title, language, parse_status) VALUES ($1, $2, 'FK Resume', 'en', 'ready')`, ids.resumeAssetID, ids.userID)
	mustExec(t, ctx, db, `INSERT INTO target_jobs(id, user_id, source_type) VALUES ($1, $2, 'manual_text')`, ids.targetJobID, ids.userID)
	mustExec(t, ctx, db, `INSERT INTO resume_versions(id, user_id, resume_asset_id, version_type, display_name) VALUES ($1, $2, $3, 'structured_master', 'Master')`, ids.resumeVersionID, ids.userID, ids.resumeAssetID)
	mustExec(t, ctx, db, `INSERT INTO resume_tailor_runs(id, user_id, target_job_id, resume_asset_id, mode, status) VALUES ($1, $2, $3, $4, 'gap_review', 'ready')`, ids.tailorRunID, ids.userID, ids.targetJobID, ids.resumeAssetID)
	return ids
}

func cleanupUser(db *sql.DB, userID string) {
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	_, _ = db.ExecContext(cleanupCtx, `DELETE FROM users WHERE id = $1`, userID)
}

func expectExecCheckViolation(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	_, err := db.ExecContext(ctx, query, args...)
	expectCheckViolation(t, err)
}

func expectCheckViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected check constraint violation")
	}
	pqErr, ok := err.(*pq.Error)
	if !ok {
		t.Fatalf("expected pq error, got %T: %v", err, err)
	}
	if pqErr.Code != "23514" {
		t.Fatalf("expected postgres check violation 23514, got %s: %v", pqErr.Code, err)
	}
}

func expectForeignKeyViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected foreign key violation")
	}
	pqErr, ok := err.(*pq.Error)
	if !ok {
		t.Fatalf("expected pq error, got %T: %v", err, err)
	}
	if pqErr.Code != "23503" {
		t.Fatalf("expected postgres foreign key violation 23503, got %s: %v", pqErr.Code, err)
	}
}
