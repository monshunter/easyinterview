//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	uploadstore "github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestFileObjectsIntegrationDatabaseAvailable(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping file_objects integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
}

func TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping file_objects integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := uploadstore.NewRepository(db)
	now := time.Date(2026, 5, 12, 1, 15, 0, 0, time.UTC)
	auditID := "018f2a40-0000-7000-9000-000000000401"
	fileID := "018f2a40-0000-7000-9000-000000000402"
	objectKey := "user-1/resume/file-1.pdf"
	t.Cleanup(func() {
		_, _ = db.Exec(`delete from audit_events where id = $1`, auditID)
	})

	err = repo.InsertAuditTombstone(contextWithTimeout(t), uploadstore.AuditTombstoneInput{
		AuditEventID: auditID,
		FileObjectID: fileID,
		Purpose:      uploadstore.PurposeResume,
		ObjectKey:    objectKey,
		DeletedAt:    now,
	})
	if err != nil {
		t.Fatalf("InsertAuditTombstone: %v", err)
	}
	var metadata string
	if err := db.QueryRow(`select metadata::text from audit_events where id = $1`, auditID).Scan(&metadata); err != nil {
		t.Fatalf("select audit metadata: %v", err)
	}
	if strings.Contains(metadata, objectKey) || strings.Contains(metadata, "objectKey") {
		t.Fatalf("audit tombstone leaked object key: %s", metadata)
	}
}

func contextWithTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
