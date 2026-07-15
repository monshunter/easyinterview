//go:build integration

package migrations

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

func TestIntegrationUserSettingsDisplayPreferencesUpDownUp(t *testing.T) {
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
	if _, _, err := migrator.Version(); !errors.Is(err, migrate.ErrNilVersion) {
		t.Fatalf("integration database must be empty, version error=%v", err)
	}
	if err := migrator.Migrate(19); err != nil {
		t.Fatalf("migrate clean database to v19: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	const userID = "019f650d-0000-7000-8000-000000000020"
	mustExecIntegration(t, ctx, db, `
insert into users (id,email,display_name) values ($1,'migration-v20@example.test','Migration V20')`, userID)
	mustExecIntegration(t, ctx, db, `
insert into user_settings (
  user_id,ui_language,preferred_practice_language,region,timezone,analytics_opt_in
) values ($1,'fr-FR','zh-CN','EU','Europe/Paris',false)`, userID)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v19 to v20: %v", err)
	}
	assertMigrationVersion(t, migrator, 20)
	assertUserSettingsCurrentColumns(t, ctx, db)
	assertAnalyticsOptIn(t, ctx, db, userID, false)

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("migrate populated database v20 down to v19: %v", err)
	}
	assertMigrationVersion(t, migrator, 19)
	assertAnalyticsOptIn(t, ctx, db, userID, false)
	var uiLanguage, practiceLanguage, timezone string
	var region sql.NullString
	if err := db.QueryRowContext(ctx, `
select ui_language,preferred_practice_language,region,timezone
from user_settings where user_id=$1`, userID).Scan(&uiLanguage, &practiceLanguage, &region, &timezone); err != nil {
		t.Fatalf("read restored display preferences: %v", err)
	}
	if uiLanguage != "zh-CN" || practiceLanguage != "en" || region.Valid || timezone != "UTC" {
		t.Fatalf("down migration restored values=%q/%q/%v/%q, want defaults without deleted-value recovery", uiLanguage, practiceLanguage, region, timezone)
	}

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v19 back to v20: %v", err)
	}
	assertMigrationVersion(t, migrator, 20)
	assertUserSettingsCurrentColumns(t, ctx, db)
	assertAnalyticsOptIn(t, ctx, db, userID, false)

	mustExecIntegration(t, ctx, db, `delete from users where id=$1`, userID)
	var remaining int
	if err := db.QueryRowContext(ctx, `select count(*) from user_settings where user_id=$1`, userID).Scan(&remaining); err != nil {
		t.Fatalf("count privacy-deleted user_settings: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("user_settings survived user hard deletion: count=%d", remaining)
	}

	t.Log("USER_SETTINGS_DISPLAY_PREFERENCES_POPULATED_PASS")
	t.Log("USER_SETTINGS_DISPLAY_PREFERENCES_DOWN_DEFAULTS_PASS")
	t.Log("USER_SETTINGS_PRIVACY_CASCADE_PASS")
}

func assertUserSettingsCurrentColumns(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	rows, err := db.QueryContext(ctx, `
select column_name
from information_schema.columns
where table_schema='public' and table_name='user_settings'
order by ordinal_position`)
	if err != nil {
		t.Fatalf("read user_settings columns: %v", err)
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			t.Fatalf("scan user_settings column: %v", err)
		}
		columns = append(columns, column)
	}
	want := []string{"user_id", "analytics_opt_in", "created_at", "updated_at"}
	if len(columns) != len(want) {
		t.Fatalf("user_settings columns=%v, want %v", columns, want)
	}
	for i := range want {
		if columns[i] != want[i] {
			t.Fatalf("user_settings columns=%v, want %v", columns, want)
		}
	}
}

func assertAnalyticsOptIn(t *testing.T, ctx context.Context, db *sql.DB, userID string, want bool) {
	t.Helper()
	var got bool
	if err := db.QueryRowContext(ctx, `select analytics_opt_in from user_settings where user_id=$1`, userID).Scan(&got); err != nil {
		t.Fatalf("read analytics_opt_in: %v", err)
	}
	if got != want {
		t.Fatalf("analytics_opt_in=%t, want %t", got, want)
	}
}
