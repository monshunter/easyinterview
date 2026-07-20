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

func TestIntegrationAccountThemePreferencesUpDownUp(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL must identify an empty disposable PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	root := repoRoot(t)
	migrator, err := newMigrate(Command{
		DatabaseURL:      dsn,
		MigrationsDir:    filepath.Join(root, "migrations"),
		BackfillManifest: filepath.Join(root, "migrations", "backfill", "manifest.yaml"),
	})
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	defer closeMigrate(migrator)
	if _, _, err := migrator.Version(); !errors.Is(err, migrate.ErrNilVersion) {
		t.Fatalf("integration database must be empty, version error=%v", err)
	}
	if err := migrator.Migrate(20); err != nil {
		t.Fatalf("migrate clean database to v20: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	const userID = "019f77f8-0000-7000-8000-000000000021"
	mustExecIntegration(t, ctx, db, `insert into users (id,email,display_name) values ($1,'theme-v21@example.test','Theme V21')`, userID)
	mustExecIntegration(t, ctx, db, `insert into user_settings (user_id,analytics_opt_in) values ($1,false)`, userID)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v20 to v21: %v", err)
	}
	assertMigrationVersion(t, migrator, 21)
	assertAccountTheme(t, ctx, db, userID, "ocean", sql.NullFloat64{}, sql.NullFloat64{})
	assertAnalyticsOptIn(t, ctx, db, userID, false)

	mustExecIntegration(t, ctx, db, `update user_settings set theme='plum',custom_accent_hue=315,custom_accent_chroma=0.18 where user_id=$1`, userID)
	assertAccountTheme(t, ctx, db, userID, "plum", sql.NullFloat64{Float64: 315, Valid: true}, sql.NullFloat64{Float64: 0.18, Valid: true})
	for _, statement := range []string{
		`update user_settings set theme='forest' where user_id=$1`,
		`update user_settings set custom_accent_hue=20,custom_accent_chroma=null where user_id=$1`,
		`update user_settings set custom_accent_hue=360,custom_accent_chroma=0.2 where user_id=$1`,
		`update user_settings set custom_accent_hue=20,custom_accent_chroma=0.281 where user_id=$1`,
	} {
		if _, err := db.ExecContext(ctx, statement, userID); err == nil {
			t.Fatalf("invalid account theme statement succeeded: %s", statement)
		}
	}

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v21 to v22: %v", err)
	}
	assertMigrationVersion(t, migrator, 22)
	mustExecIntegration(t, ctx, db, `update user_settings set theme='forest',custom_accent_hue=null,custom_accent_chroma=null where user_id=$1`, userID)
	assertAccountTheme(t, ctx, db, userID, "forest", sql.NullFloat64{}, sql.NullFloat64{})
	if _, err := db.ExecContext(ctx, `update user_settings set theme='warm' where user_id=$1`, userID); err == nil {
		t.Fatal("unsupported warm account theme succeeded at v22")
	}

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("migrate populated database v22 down to v21: %v", err)
	}
	assertMigrationVersion(t, migrator, 21)
	assertAccountTheme(t, ctx, db, userID, "ocean", sql.NullFloat64{}, sql.NullFloat64{})
	assertAnalyticsOptIn(t, ctx, db, userID, false)
	if err := migrator.Steps(1); err != nil {
		t.Fatalf("migrate populated database v21 back to v22: %v", err)
	}
	assertMigrationVersion(t, migrator, 22)
	assertAccountTheme(t, ctx, db, userID, "ocean", sql.NullFloat64{}, sql.NullFloat64{})
	mustExecIntegration(t, ctx, db, `update user_settings set theme='forest' where user_id=$1`, userID)
	assertAccountTheme(t, ctx, db, userID, "forest", sql.NullFloat64{}, sql.NullFloat64{})

	mustExecIntegration(t, ctx, db, `delete from users where id=$1`, userID)
	var remaining int
	if err := db.QueryRowContext(ctx, `select count(*) from user_settings where user_id=$1`, userID).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("user_settings privacy cascade count=%d err=%v", remaining, err)
	}
	t.Log("ACCOUNT_THEME_PREFERENCES_POPULATED_PASS")
	t.Log("ACCOUNT_THEME_PREFERENCES_CONSTRAINTS_PASS")
	t.Log("ACCOUNT_THEME_FOREST_UP_DOWN_UP_PASS")
	t.Log("ACCOUNT_THEME_PREFERENCES_UP_DOWN_UP_PASS")
}

func assertAccountTheme(t *testing.T, ctx context.Context, db *sql.DB, userID, wantTheme string, wantHue, wantChroma sql.NullFloat64) {
	t.Helper()
	var theme string
	var hue, chroma sql.NullFloat64
	if err := db.QueryRowContext(ctx, `select theme,custom_accent_hue,custom_accent_chroma from user_settings where user_id=$1`, userID).Scan(&theme, &hue, &chroma); err != nil {
		t.Fatalf("read account theme: %v", err)
	}
	if theme != wantTheme || hue != wantHue || chroma != wantChroma {
		t.Fatalf("account theme=%q/%v/%v, want %q/%v/%v", theme, hue, chroma, wantTheme, wantHue, wantChroma)
	}
}
