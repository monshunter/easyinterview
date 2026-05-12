//go:build integration

package store_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
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
