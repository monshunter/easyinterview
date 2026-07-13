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

func TestIntegrationReportAndPracticeV020ActivationRollbackReactivate(t *testing.T) {
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
	if err := migrator.Migrate(18); err != nil {
		t.Fatalf("migrate clean database to v18: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	assertV020ActivationState(t, ctx, db, "v0.1.0", false)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate v18 to v19: %v", err)
	}
	assertMigrationVersion(t, migrator, 19)
	assertV020ActivationState(t, ctx, db, "v0.2.0", true)

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("roll back v19 to v18: %v", err)
	}
	assertMigrationVersion(t, migrator, 18)
	assertV020ActivationState(t, ctx, db, "v0.1.0", false)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("reactivate v19: %v", err)
	}
	assertMigrationVersion(t, migrator, 19)
	assertV020ActivationState(t, ctx, db, "v0.2.0", true)

	t.Log("REPORT_PROMPT_V020_PASS")
	t.Log("PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS")
}

func assertV020ActivationState(t *testing.T, ctx context.Context, db *sql.DB, activeVersion string, v020Present bool) {
	t.Helper()
	for _, featureKey := range []string{"report.generate", "practice.session.chat"} {
		for _, table := range []string{"prompt_versions", "rubric_versions"} {
			var version string
			query := `select version from ` + table + ` where feature_key=$1 and language='multi' and is_active`
			if err := db.QueryRowContext(ctx, query, featureKey).Scan(&version); err != nil {
				t.Fatalf("read active %s %s: %v", table, featureKey, err)
			}
			if version != activeVersion {
				t.Fatalf("active %s %s version=%s, want %s", table, featureKey, version, activeVersion)
			}

			var count int
			query = `select count(*) from ` + table + ` where feature_key=$1 and language='multi' and version='v0.2.0'`
			if err := db.QueryRowContext(ctx, query, featureKey).Scan(&count); err != nil {
				t.Fatalf("count v0.2 %s %s: %v", table, featureKey, err)
			}
			wantCount := 0
			if v020Present {
				wantCount = 1
			}
			if count != wantCount {
				t.Fatalf("v0.2 %s %s count=%d, want %d", table, featureKey, count, wantCount)
			}
		}
	}

	if !v020Present {
		return
	}
	for featureKey, wantHash := range map[string]string{
		"report.generate":       "e99faa33b00842c9320068faa2207f9dedb7af5e4742d7e51c2c81c4542e2fed",
		"practice.session.chat": "d361c6401bc440825393bdaf093d42be53892de50bf4d5c4e9cdba0562a2bc9e",
	} {
		var gotHash string
		if err := db.QueryRowContext(ctx, `
select template_hash from prompt_versions
where feature_key=$1 and language='multi' and version='v0.2.0'`, featureKey).Scan(&gotHash); err != nil {
			t.Fatalf("read v0.2 prompt hash %s: %v", featureKey, err)
		}
		if gotHash != wantHash {
			t.Fatalf("v0.2 prompt hash %s=%s, want %s", featureKey, gotHash, wantHash)
		}
	}

	var practiceRubricsEqual bool
	if err := db.QueryRowContext(ctx, `
select (v010.schema_json - 'version') = (v020.schema_json - 'version')
from rubric_versions v010
join rubric_versions v020 using (feature_key, language)
where v010.feature_key='practice.session.chat'
  and v010.version='v0.1.0'
  and v020.version='v0.2.0'`).Scan(&practiceRubricsEqual); err != nil {
		t.Fatalf("compare practice rubric versions: %v", err)
	}
	if !practiceRubricsEqual {
		t.Fatal("practice v0.2 rubric must be content-identical to v0.1")
	}
}
