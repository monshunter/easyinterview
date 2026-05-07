package migrations

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Command is the normalized command contract used by cmd/migrate.
type Command struct {
	Name             string
	DatabaseURL      string
	MigrationsDir    string
	BackfillManifest string
	AppEnv           string
	ForceBackfill    bool
	Stdout           io.Writer
}

// RunCommand executes DB-backed migration commands.
func RunCommand(ctx context.Context, cmd Command) error {
	_ = ctx
	if cmd.Stdout == nil {
		cmd.Stdout = io.Discard
	}
	if cmd.Name == "up" || cmd.Name == "down" || cmd.Name == "check" {
		if err := ValidateMigrationFiles(cmd.MigrationsDir); err != nil {
			return err
		}
	}
	switch cmd.Name {
	case "up":
		return runUp(cmd)
	case "down":
		return runDown(cmd)
	case "status":
		return runStatus(cmd)
	case "check":
		if err := runUp(cmd); err != nil {
			return fmt.Errorf("migrate-up: %w", err)
		}
		if err := runDown(cmd); err != nil {
			return fmt.Errorf("migrate-down: %w", err)
		}
		if err := runUp(cmd); err != nil {
			return fmt.Errorf("migrate-up after down: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %s", cmd.Name)
	}
}

func runUp(cmd Command) error {
	m, err := newMigrate(cmd)
	if err != nil {
		return err
	}
	defer closeMigrate(m)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return RunBackfills(cmd)
}

func runDown(cmd Command) error {
	m, err := newMigrate(cmd)
	if err != nil {
		return err
	}
	defer closeMigrate(m)
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func runStatus(cmd Command) error {
	m, err := newMigrate(cmd)
	if err != nil {
		return err
	}
	defer closeMigrate(m)
	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		fmt.Fprintln(cmd.Stdout, "version=nil dirty=false")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.Stdout, "version=%d dirty=%t\n", version, dirty)
	return nil
}

func newMigrate(cmd Command) (*migrate.Migrate, error) {
	dir, err := filepath.Abs(cmd.MigrationsDir)
	if err != nil {
		return nil, err
	}
	sourceURL := (&url.URL{Scheme: "file", Path: dir}).String()
	return migrate.New(sourceURL, cmd.DatabaseURL)
}

func closeMigrate(m *migrate.Migrate) {
	sourceErr, databaseErr := m.Close()
	_, _ = sourceErr, databaseErr
}
