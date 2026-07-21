//go:build integration

package migrations

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegrationPracticeV030ActivationRollbackReactivate(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL must identify an empty disposable PostgreSQL database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := repoRoot(t)
	cmd := Command{
		DatabaseURL:      dsn,
		MigrationsDir:    filepath.Join(root, "migrations"),
		BackfillManifest: filepath.Join(root, "migrations", "backfill", "manifest.yaml"),
	}
	migrator, err := newMigrate(cmd)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	defer closeMigrate(migrator)
	if err := migrator.Migrate(22); err != nil {
		t.Fatalf("migrate clean database to v22: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	assertPracticeV030ActivationState(t, ctx, db, "v0.2.0", false)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate v22 to v23: %v", err)
	}
	assertMigrationVersion(t, migrator, 23)
	assertPracticeV030ActivationState(t, ctx, db, "v0.3.0", true)

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("roll back v23 to v22: %v", err)
	}
	assertMigrationVersion(t, migrator, 22)
	assertPracticeV030ActivationState(t, ctx, db, "v0.2.0", false)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("reactivate v23: %v", err)
	}
	assertMigrationVersion(t, migrator, 23)
	assertPracticeV030ActivationState(t, ctx, db, "v0.3.0", true)

	t.Log("PRACTICE_INTERVIEWER_IDENTITY_V030_DB_PASS")
}

func assertPracticeV030ActivationState(t *testing.T, ctx context.Context, db *sql.DB, activeVersion string, v030Present bool) {
	t.Helper()
	for _, table := range []string{"prompt_versions", "rubric_versions"} {
		var version string
		query := `select version from ` + table + ` where feature_key='practice.session.chat' and language='multi' and is_active`
		if err := db.QueryRowContext(ctx, query).Scan(&version); err != nil {
			t.Fatalf("read active practice %s: %v", table, err)
		}
		if version != activeVersion {
			t.Fatalf("active practice %s version=%s, want %s", table, version, activeVersion)
		}

		var count int
		query = `select count(*) from ` + table + ` where feature_key='practice.session.chat' and language='multi' and version='v0.3.0'`
		if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			t.Fatalf("count practice v0.3 %s: %v", table, err)
		}
		wantCount := 0
		if v030Present {
			wantCount = 1
		}
		if count != wantCount {
			t.Fatalf("practice v0.3 %s count=%d, want %d", table, count, wantCount)
		}
	}

	var reportPrompt, reportRubric string
	if err := db.QueryRowContext(ctx, `select version from prompt_versions where feature_key='report.generate' and language='multi' and is_active`).Scan(&reportPrompt); err != nil {
		t.Fatalf("read active report prompt: %v", err)
	}
	if err := db.QueryRowContext(ctx, `select version from rubric_versions where feature_key='report.generate' and language='multi' and is_active`).Scan(&reportRubric); err != nil {
		t.Fatalf("read active report rubric: %v", err)
	}
	if reportPrompt != "v0.2.0" || reportRubric != "v0.2.0" {
		t.Fatalf("report coordinate changed to %s/%s", reportPrompt, reportRubric)
	}

	if !v030Present {
		return
	}
	var hash string
	if err := db.QueryRowContext(ctx, `
select template_hash from prompt_versions
where feature_key='practice.session.chat' and language='multi' and version='v0.3.0'`).Scan(&hash); err != nil {
		t.Fatalf("read practice v0.3 prompt hash: %v", err)
	}
	if hash != "9fff2605695aed41c3c81efd3f8d35e15b6ecad851a5b3abd482540e402b496d" {
		t.Fatalf("practice v0.3 prompt hash=%s", hash)
	}
	var roleWeight float64
	if err := db.QueryRowContext(ctx, `
select (dimension->>'weight')::double precision
from rubric_versions,
     jsonb_array_elements(schema_json->'dimensions') as dimension
where feature_key='practice.session.chat'
  and language='multi'
  and version='v0.3.0'
  and dimension->>'name'='role_identity'`).Scan(&roleWeight); err != nil {
		t.Fatalf("read practice v0.3 role_identity weight: %v", err)
	}
	if roleWeight != 0.4 {
		t.Fatalf("practice v0.3 role_identity weight=%v, want 0.4", roleWeight)
	}
}
